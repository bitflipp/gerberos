package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

type backend interface {
	Initialize() error
	Ban(string, bool, time.Duration) error
	Finalize() error
}

type ipsetBackend struct {
	chainName  string
	ipset4Name string
	ipset6Name string
}

func (b *ipsetBackend) deleteIpsetsAndIptablesEntries() error {
	if s, ec, _ := execute("iptables", "-D", b.chainName, "-j", "DROP", "-m", "set", "--match-set", b.ipset4Name, "src"); ec > 2 {
		return fmt.Errorf(`failed to delete iptables entry for set "%s": %s`, b.ipset4Name, s)
	}
	if s, ec, _ := execute("iptables", "-D", "INPUT", "-j", b.chainName); ec > 2 {
		return fmt.Errorf(`failed to delete iptables entry for chain "%s": %s`, b.chainName, s)
	}
	if s, ec, _ := execute("iptables", "-X", b.chainName); ec > 2 {
		return fmt.Errorf(`failed to delete iptables chain "%s": %s`, b.chainName, s)
	}
	if s, ec, _ := execute("ip6tables", "-D", b.chainName, "-j", "DROP", "-m", "set", "--match-set", b.ipset6Name, "src"); ec > 2 {
		return fmt.Errorf(`failed to delete ip6tables entry for set "%s": %s`, b.ipset6Name, s)
	}
	if s, ec, _ := execute("ip6tables", "-D", "INPUT", "-j", b.chainName); ec > 2 {
		return fmt.Errorf(`failed to delete ip6tables entry for chain "%s": %s`, b.chainName, s)
	}
	if s, ec, _ := execute("ip6tables", "-X", b.chainName); ec > 2 {
		return fmt.Errorf(`failed to delete ip6tables chain "%s": %s`, b.chainName, s)
	}
	if s, ec, _ := execute("ipset", "destroy", b.ipset4Name); ec > 1 {
		return fmt.Errorf(`failed to destroy ipset "%s": %s`, b.ipset4Name, s)
	}
	if s, ec, _ := execute("ipset", "destroy", b.ipset6Name); ec > 1 {
		return fmt.Errorf(`failed to destroy ipset "%s": %s`, b.ipset6Name, s)
	}

	return nil
}

func (b *ipsetBackend) createIpsets() error {
	if s, ec, _ := execute("ipset", "create", b.ipset4Name, "hash:ip", "timeout", "0"); ec != 0 {
		return fmt.Errorf(`failed to create ipset "%s": %s`, b.ipset4Name, s)
	}
	if s, ec, _ := execute("ipset", "create", b.ipset6Name, "hash:ip", "family", "inet6", "timeout", "0"); ec != 0 {
		return fmt.Errorf(`failed to create ipset "%s": %s`, b.ipset6Name, s)
	}

	return nil
}

func (b *ipsetBackend) createIptablesEntries() error {
	if s, ec, _ := execute("iptables", "-N", b.chainName); ec != 0 {
		return fmt.Errorf(`failed to create iptables chain "%s": %s`, b.chainName, s)
	}
	if s, ec, _ := execute("iptables", "-I", b.chainName, "-j", "DROP", "-m", "set", "--match-set", b.ipset4Name, "src"); ec != 0 {
		return fmt.Errorf(`failed to create iptables entry for set "%s": %s`, b.ipset4Name, s)
	}
	if s, ec, _ := execute("iptables", "-I", "INPUT", "-j", b.chainName); ec != 0 {
		return fmt.Errorf(`failed to create iptables entry for chain "%s": %s`, b.chainName, s)
	}
	if s, ec, _ := execute("ip6tables", "-N", b.chainName); ec != 0 {
		return fmt.Errorf(`failed to create ip6tables chain "%s": %s`, b.chainName, s)
	}
	if s, ec, _ := execute("ip6tables", "-I", b.chainName, "-j", "DROP", "-m", "set", "--match-set", b.ipset6Name, "src"); ec != 0 {
		return fmt.Errorf(`failed to create ip6tables entry for set "%s": %s`, b.ipset6Name, s)
	}
	if s, ec, _ := execute("ip6tables", "-I", "INPUT", "-j", b.chainName); ec != 0 {
		return fmt.Errorf(`failed to create ip6tables entry for chain "%s": %s`, b.chainName, s)
	}

	return nil
}

func (b *ipsetBackend) saveIpsets() error {
	f, err := os.Create(*configuration.SaveFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command("ipset", "save")
	cmd.Stdout = f
	err = cmd.Run()
	if err != nil {
		return err
	}

	// Always ensure file is saved to disk. This should prevent loss of banned IPs on shutdown.
	return f.Sync()
}

func (b *ipsetBackend) restoreIpsets() error {
	f, err := os.Open(*configuration.SaveFilePath)
	if err != nil {
		return err
	}

	defer func() {
		f.Close()
		if err := os.Remove(*configuration.SaveFilePath); err != nil {
			log.Printf("failed to delete save file: %s", err)
		}
	}()

	cmd := exec.Command("ipset", "restore")
	cmd.Stdin = f
	return cmd.Run()
}

func (b *ipsetBackend) Initialize() error {
	b.chainName = "gerberos"
	b.ipset4Name = "gerberos4"
	b.ipset6Name = "gerberos6"

	// Check privileges
	if _, _, err := execute("ipset", "list"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errors.New("ipset: command not found")
		}
		return errors.New("ipset: insufficient privileges")
	}
	if _, _, err := execute("iptables", "-L"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errors.New("iptables: command not found")
		}
		return errors.New("iptables: insufficient privileges")
	}
	if _, _, err := execute("ip6tables", "-L"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errors.New("ip6tables: command not found")
		}
		return errors.New("ip6tables: insufficient privileges")
	}

	// Initialize ipsets and ip(6)tables entries
	if err := b.deleteIpsetsAndIptablesEntries(); err != nil {
		return fmt.Errorf("failed to delete ipsets: %w", err)
	}
	if configuration.SaveFilePath != nil {
		if err := b.restoreIpsets(); err != nil {
			if err := b.createIpsets(); err != nil {
				return fmt.Errorf("failed to create ipsets: %w", err)
			}
		} else {
			log.Printf(`restored ipsets from "%s"`, *configuration.SaveFilePath)
		}
	} else {
		log.Printf("warning: not persisting ipsets")
		if err := b.createIpsets(); err != nil {
			return fmt.Errorf("failed to create ipsets: %w", err)
		}
	}
	if err := b.createIptablesEntries(); err != nil {
		return fmt.Errorf("failed to create ip(6)tables entries: %w", err)
	}

	return nil
}

func (b *ipsetBackend) Ban(ip string, ipv6 bool, d time.Duration) error {
	s := b.ipset4Name
	if ipv6 {
		s = b.ipset6Name
	}
	ds := int64(d.Seconds())
	if _, _, err := execute("ipset", "test", s, ip); err != nil {
		if _, _, err := execute("ipset", "add", s, ip, "timeout", fmt.Sprint(ds)); err != nil {
			return err
		}
	}
	return nil
}

func (b *ipsetBackend) Finalize() error {
	if configuration.SaveFilePath != nil {
		if err := b.saveIpsets(); err != nil {
			return fmt.Errorf(`failed to save ipsets to "%s": %w`, *configuration.SaveFilePath, err)
		}
	}
	if err := b.deleteIpsetsAndIptablesEntries(); err != nil {
		return fmt.Errorf("failed to delete ipsets and ip(6)tables entries: %w", err)
	}
	return nil
}

type nftBackend struct {
	table4Name string
	table6Name string
	set4Name   string
	set6Name   string
}

func (b *nftBackend) createTables() error {
	if s, _, err := execute("nft", "add", "table", "ip", b.table4Name); err != nil {
		return fmt.Errorf(`failed to add table "%s": %s`, b.table4Name, s)
	}
	if s, _, err := execute("nft", "add", "set", "ip", b.table4Name, b.set4Name, "{ type ipv4_addr; flags timeout; }"); err != nil {
		return fmt.Errorf(`failed to add ip set "%s": %s`, b.table4Name, s)
	}
	if s, _, err := execute("nft", "add", "chain", "ip", b.table4Name, "input", "{ type filter hook input priority filter; policy accept; }"); err != nil {
		return fmt.Errorf(`failed to add input chain: %s`, s)
	}
	if s, _, err := execute("nft", "flush", "chain", "ip", b.table4Name, "input"); err != nil {
		return fmt.Errorf(`failed to flush input chain: %s`, s)
	}
	if s, _, err := execute("nft", "add", "rule", "ip", b.table4Name, "input", "ip", "saddr", "@"+b.set4Name, "drop"); err != nil {
		return fmt.Errorf(`failed to add rule: %s`, s)
	}
	if s, _, err := execute("nft", "add", "table", "ip6", b.table6Name); err != nil {
		return fmt.Errorf(`failed to create ip6 table "%s": %s`, b.table6Name, s)
	}
	if s, _, err := execute("nft", "add", "set", "ip6", b.table6Name, b.set6Name, "{ type ipv6_addr; flags timeout; }"); err != nil {
		return fmt.Errorf(`failed to add ip set "%s": %s`, b.table6Name, s)
	}
	if s, _, err := execute("nft", "add", "chain", "ip6", b.table6Name, "input", "{ type filter hook input priority filter; policy accept; }"); err != nil {
		return fmt.Errorf(`failed to add input chain: %s`, s)
	}
	if s, _, err := execute("nft", "flush", "chain", "ip6", b.table6Name, "input"); err != nil {
		return fmt.Errorf(`failed to flush input chain: %s`, s)
	}
	if s, _, err := execute("nft", "add", "rule", "ip6", b.table6Name, "input", "ip6", "saddr", "@"+b.set6Name, "drop"); err != nil {
		return fmt.Errorf(`failed to add rule: %s`, s)
	}

	return nil
}

func (b *nftBackend) saveSets() error {
	f, err := os.Create(*configuration.SaveFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command("nft", "list", "set", "ip", b.table4Name, b.set4Name)
	cmd.Stdout = f
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("nft", "list", "set", "ip6", b.table6Name, b.set6Name)
	cmd.Stdout = f
	err = cmd.Run()
	if err != nil {
		return err
	}

	// Always ensure file is saved to disk. This should prevent loss of banned IPs on shutdown.
	return f.Sync()
}

func (b *nftBackend) restoreSets() error {
	return exec.Command("nft", "-f", *configuration.SaveFilePath).Run()
}

func (b *nftBackend) Initialize() error {
	b.table4Name = "gerberos4"
	b.table6Name = "gerberos6"
	b.set4Name = "set4"
	b.set6Name = "set6"

	// Check privileges
	if _, _, err := execute("nft", "list", "ruleset"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errors.New("nft: command not found")
		}
		return errors.New("nft: insufficient privileges")
	}

	if err := b.createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	if configuration.SaveFilePath != nil {
		if err := b.restoreSets(); err != nil {
			log.Printf(`failed to restore sets from "%s": %s`, *configuration.SaveFilePath, err)
		} else {
			log.Printf(`restored sets from "%s"`, *configuration.SaveFilePath)
		}
	} else {
		log.Printf("warning: not persisting sets")
	}

	return nil
}

func (b *nftBackend) Ban(ip string, ipv6 bool, d time.Duration) error {
	ds := int64(d.Seconds())

	if ipv6 {
		if _, _, err := execute("nft", "add", "element", "ip6", b.table6Name, b.set6Name, fmt.Sprintf("{ %s timeout %ds }", ip, ds)); err != nil {
			return fmt.Errorf(`failed to add element to set "%s": %w`, b.set6Name, err)
		}
	} else {
		if _, _, err := execute("nft", "add", "element", "ip", b.table4Name, b.set4Name, fmt.Sprintf("{ %s timeout %ds }", ip, ds)); err != nil {
			return fmt.Errorf(`failed to add element to set "%s": %w`, b.set4Name, err)
		}
	}

	return nil
}

func (b *nftBackend) Finalize() error {
	if configuration.SaveFilePath != nil {
		if err := b.saveSets(); err != nil {
			return fmt.Errorf(`failed to save sets to "%s": %w`, *configuration.SaveFilePath, err)
		}
	}

	return nil
}
