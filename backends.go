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
	runner     *runner
	chainName  string
	ipset4Name string
	ipset6Name string
}

func (b *ipsetBackend) deleteIpsetsAndIptablesEntries() error {
	if s, ec, _ := execute(nil, "iptables", "-D", b.chainName, "-j", "DROP", "-m", "set", "--match-set", b.ipset4Name, "src"); ec > 2 {
		return fmt.Errorf(`failed to delete iptables entry for set "%s": %s`, b.ipset4Name, s)
	}
	if s, ec, _ := execute(nil, "iptables", "-D", "INPUT", "-j", b.chainName); ec > 2 {
		return fmt.Errorf(`failed to delete iptables entry for chain "%s": %s`, b.chainName, s)
	}
	if s, ec, _ := execute(nil, "iptables", "-X", b.chainName); ec > 2 {
		return fmt.Errorf(`failed to delete iptables chain "%s": %s`, b.chainName, s)
	}
	if s, ec, _ := execute(nil, "ip6tables", "-D", b.chainName, "-j", "DROP", "-m", "set", "--match-set", b.ipset6Name, "src"); ec > 2 {
		return fmt.Errorf(`failed to delete ip6tables entry for set "%s": %s`, b.ipset6Name, s)
	}
	if s, ec, _ := execute(nil, "ip6tables", "-D", "INPUT", "-j", b.chainName); ec > 2 {
		return fmt.Errorf(`failed to delete ip6tables entry for chain "%s": %s`, b.chainName, s)
	}
	if s, ec, _ := execute(nil, "ip6tables", "-X", b.chainName); ec > 2 {
		return fmt.Errorf(`failed to delete ip6tables chain "%s": %s`, b.chainName, s)
	}
	if s, ec, _ := execute(nil, "ipset", "destroy", b.ipset4Name); ec > 1 {
		return fmt.Errorf(`failed to destroy ipset "%s": %s`, b.ipset4Name, s)
	}
	if s, ec, _ := execute(nil, "ipset", "destroy", b.ipset6Name); ec > 1 {
		return fmt.Errorf(`failed to destroy ipset "%s": %s`, b.ipset6Name, s)
	}

	return nil
}

func (b *ipsetBackend) createIpsets() error {
	if s, ec, _ := execute(nil, "ipset", "create", b.ipset4Name, "hash:ip", "timeout", "0"); ec != 0 {
		return fmt.Errorf(`failed to create ipset "%s": %s`, b.ipset4Name, s)
	}
	if s, ec, _ := execute(nil, "ipset", "create", b.ipset6Name, "hash:ip", "family", "inet6", "timeout", "0"); ec != 0 {
		return fmt.Errorf(`failed to create ipset "%s": %s`, b.ipset6Name, s)
	}

	return nil
}

func (b *ipsetBackend) createIptablesEntries() error {
	if s, ec, _ := execute(nil, "iptables", "-N", b.chainName); ec != 0 {
		return fmt.Errorf(`failed to create iptables chain "%s": %s`, b.chainName, s)
	}
	if s, ec, _ := execute(nil, "iptables", "-I", b.chainName, "-j", "DROP", "-m", "set", "--match-set", b.ipset4Name, "src"); ec != 0 {
		return fmt.Errorf(`failed to create iptables entry for set "%s": %s`, b.ipset4Name, s)
	}
	if s, ec, _ := execute(nil, "iptables", "-I", "INPUT", "-j", b.chainName); ec != 0 {
		return fmt.Errorf(`failed to create iptables entry for chain "%s": %s`, b.chainName, s)
	}
	if s, ec, _ := execute(nil, "ip6tables", "-N", b.chainName); ec != 0 {
		return fmt.Errorf(`failed to create ip6tables chain "%s": %s`, b.chainName, s)
	}
	if s, ec, _ := execute(nil, "ip6tables", "-I", b.chainName, "-j", "DROP", "-m", "set", "--match-set", b.ipset6Name, "src"); ec != 0 {
		return fmt.Errorf(`failed to create ip6tables entry for set "%s": %s`, b.ipset6Name, s)
	}
	if s, ec, _ := execute(nil, "ip6tables", "-I", "INPUT", "-j", b.chainName); ec != 0 {
		return fmt.Errorf(`failed to create ip6tables entry for chain "%s": %s`, b.chainName, s)
	}

	return nil
}

func (b *ipsetBackend) saveIpsets() error {
	f, err := os.Create(b.runner.configuration.SaveFilePath)
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
	f, err := os.Open(b.runner.configuration.SaveFilePath)
	if err != nil {
		return err
	}

	defer func() {
		f.Close()
		if err := os.Remove(b.runner.configuration.SaveFilePath); err != nil {
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
	if s, _, err := execute(nil, "ipset", "list"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errors.New("ipset: command not found")
		}
		return fmt.Errorf("ipset: insufficient privileges: %s", s)
	}
	if s, _, err := execute(nil, "iptables", "-L"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errors.New("iptables: command not found")
		}
		return fmt.Errorf("iptables: insufficient privileges: %s", s)
	}
	if s, _, err := execute(nil, "ip6tables", "-L"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errors.New("ip6tables: command not found")
		}
		return fmt.Errorf("ip6tables: insufficient privileges: %s", s)
	}

	// Initialize ipsets and ip(6)tables entries
	if err := b.deleteIpsetsAndIptablesEntries(); err != nil {
		return fmt.Errorf("failed to delete ipsets: %w", err)
	}
	if b.runner.configuration.SaveFilePath != "" {
		if err := b.restoreIpsets(); err != nil {
			if err := b.createIpsets(); err != nil {
				return fmt.Errorf("failed to create ipsets: %w", err)
			}
		} else {
			log.Printf(`restored ipsets from "%s"`, b.runner.configuration.SaveFilePath)
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
	if _, _, err := execute(nil, "ipset", "test", s, ip); err != nil {
		if _, _, err := execute(nil, "ipset", "add", s, ip, "timeout", fmt.Sprint(ds)); err != nil {
			return err
		}
	}
	return nil
}

func (b *ipsetBackend) Finalize() error {
	if b.runner.configuration.SaveFilePath != "" {
		if err := b.saveIpsets(); err != nil {
			return fmt.Errorf(`failed to save ipsets to "%s": %w`, b.runner.configuration.SaveFilePath, err)
		}
	}
	if err := b.deleteIpsetsAndIptablesEntries(); err != nil {
		return fmt.Errorf("failed to delete ipsets and ip(6)tables entries: %w", err)
	}
	return nil
}

type nftBackend struct {
	runner     *runner
	table4Name string
	table6Name string
	set4Name   string
	set6Name   string
}

func (b *nftBackend) createTables() error {
	if s, _, err := execute(nil, "nft", "add", "table", "ip", b.table4Name); err != nil {
		return fmt.Errorf(`failed to add table "%s": %s`, b.table4Name, s)
	}
	if s, _, err := execute(nil, "nft", "add", "set", "ip", b.table4Name, b.set4Name, "{ type ipv4_addr; flags timeout; }"); err != nil {
		return fmt.Errorf(`failed to add ip set "%s": %s`, b.table4Name, s)
	}
	if s, _, err := execute(nil, "nft", "add", "chain", "ip", b.table4Name, "input", "{ type filter hook input priority 0; policy accept; }"); err != nil {
		return fmt.Errorf(`failed to add input chain: %s`, s)
	}
	if s, _, err := execute(nil, "nft", "flush", "chain", "ip", b.table4Name, "input"); err != nil {
		return fmt.Errorf(`failed to flush input chain: %s`, s)
	}
	if s, _, err := execute(nil, "nft", "add", "rule", "ip", b.table4Name, "input", "ip", "saddr", "@"+b.set4Name, "reject"); err != nil {
		return fmt.Errorf(`failed to add rule: %s`, s)
	}
	if s, _, err := execute(nil, "nft", "add", "table", "ip6", b.table6Name); err != nil {
		return fmt.Errorf(`failed to create ip6 table "%s": %s`, b.table6Name, s)
	}
	if s, _, err := execute(nil, "nft", "add", "set", "ip6", b.table6Name, b.set6Name, "{ type ipv6_addr; flags timeout; }"); err != nil {
		return fmt.Errorf(`failed to add ip set "%s": %s`, b.table6Name, s)
	}
	if s, _, err := execute(nil, "nft", "add", "chain", "ip6", b.table6Name, "input", "{ type filter hook input priority 0; policy accept; }"); err != nil {
		return fmt.Errorf(`failed to add input chain: %s`, s)
	}
	if s, _, err := execute(nil, "nft", "flush", "chain", "ip6", b.table6Name, "input"); err != nil {
		return fmt.Errorf(`failed to flush input chain: %s`, s)
	}
	if s, _, err := execute(nil, "nft", "add", "rule", "ip6", b.table6Name, "input", "ip6", "saddr", "@"+b.set6Name, "reject"); err != nil {
		return fmt.Errorf(`failed to add rule: %s`, s)
	}

	return nil
}

func (b *nftBackend) deleteTables() error {
	if s, _, err := execute(nil, "nft", "delete", "table", "ip", b.table4Name); err != nil {
		return fmt.Errorf(`failed to delete table "%s": %s`, b.table4Name, s)
	}
	if s, _, err := execute(nil, "nft", "delete", "table", "ip6", b.table6Name); err != nil {
		return fmt.Errorf(`failed to delete table "%s": %s`, b.table6Name, s)
	}

	return nil
}

func (b *nftBackend) saveSets() error {
	f, err := os.Create(b.runner.configuration.SaveFilePath)
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
	return exec.Command("nft", "-f", b.runner.configuration.SaveFilePath).Run()
}

func (b *nftBackend) Initialize() error {
	b.table4Name = "gerberos4"
	b.table6Name = "gerberos6"
	b.set4Name = "set4"
	b.set6Name = "set6"

	// Check privileges
	if s, _, err := execute(nil, "nft", "list", "ruleset"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errors.New("nft: command not found")
		}
		return fmt.Errorf("nft: insufficient privileges: %s", s)
	}

	if err := b.createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	if b.runner.configuration.SaveFilePath != "" {
		if err := b.restoreSets(); err != nil {
			log.Printf(`failed to restore sets from "%s": %s`, b.runner.configuration.SaveFilePath, err)
		} else {
			log.Printf(`restored sets from "%s"`, b.runner.configuration.SaveFilePath)
		}
	} else {
		log.Printf("warning: not persisting sets")
	}

	return nil
}

func (b *nftBackend) Ban(ip string, ipv6 bool, d time.Duration) error {
	ds := int64(d.Seconds())

	if ipv6 {
		if s, ec, err := execute(nil, "nft", "add", "element", "ip6", b.table6Name, b.set6Name, fmt.Sprintf("{ %s timeout %ds }", ip, ds)); err != nil {
			if ec == 1 {
				// This ip is probably already in set. Ignore the error.
				// on netfilters >= 1.0.0, this shouldn't be a problem any more. However, since Ubuntu 20.04. only has v0.9.3, this is needed
				return nil
			}
			return fmt.Errorf(`failed to add element to set "%s": %s`, b.set6Name, s)
		}
	} else {
		if s, ec, err := execute(nil, "nft", "add", "element", "ip", b.table4Name, b.set4Name, fmt.Sprintf("{ %s timeout %ds }", ip, ds)); err != nil {
			if ec == 1 {
				// This ip is probably already in set. Ignore the error.
				// on netfilters >= 1.0.0, this shouldn't be a problem any more. However, since Ubuntu 20.04. only has v0.9.3, this is needed
				return nil
			}
			return fmt.Errorf(`failed to add element to set "%s": %s`, b.set4Name, s)
		}
	}

	return nil
}

func (b *nftBackend) Finalize() error {
	if b.runner.configuration.SaveFilePath != "" {
		if err := b.saveSets(); err != nil {
			return fmt.Errorf(`failed to save sets to "%s": %w`, b.runner.configuration.SaveFilePath, err)
		}
	}

	if err := b.deleteTables(); err != nil {
		return fmt.Errorf("failed to delete tables: %w", err)
	}

	return nil
}

type testBackend struct {
	runner *runner
}

func (b *testBackend) Initialize() error {
	return nil
}

func (b *testBackend) Ban(ip string, ipv6 bool, d time.Duration) error {
	return nil
}

func (b *testBackend) Finalize() error {
	return nil
}
