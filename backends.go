package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

type backendInterface interface {
	Initialize() error
	Ban(r *rule, ip string, ipv6 bool, d time.Duration) error
	Finalize() error
}

type ipsetBackend struct {
	chainName  string
	ipset4Name string
	ipset6Name string
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

	// Always ensure file is saved to disk. This should prevent loss of banned IPs.
	err = f.Sync()
	if err != nil {
		return err
	}

	return nil
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
		return fmt.Errorf("failed to delete ipsets: %s", err)
	}
	if configuration.SaveFilePath != nil {
		if err := b.restoreIpsets(); err != nil {
			if err := b.createIpsets(); err != nil {
				return fmt.Errorf("failed to create ipsets: %s", err)
			}
		} else {
			log.Printf(`restored ipsets from "%s"`, *configuration.SaveFilePath)
		}
	} else {
		log.Printf("warning: not persisting ipsets")
		if err := b.createIpsets(); err != nil {
			return fmt.Errorf("failed to create ipsets: %s", err)
		}
	}
	if err := b.createIptablesEntries(); err != nil {
		return fmt.Errorf("failed to create ip(6)tables entries: %s", err)
	}

	return nil
}

func (b *ipsetBackend) Ban(r *rule, ip string, ipv6 bool, d time.Duration) error {
	s := b.ipset4Name
	if ipv6 {
		s = b.ipset6Name
	}
	ds := int64(d.Seconds())
	if _, _, err := execute("ipset", "test", s, ip); err != nil {
		if _, _, err := execute("ipset", "add", s, ip, "timeout", fmt.Sprint(ds)); err != nil {
			log.Printf(`%s: failed to add "%s" to ipset "%s" with %d second(s) timeout: %s`, r.name, ip, s, ds, err)
		} else {
			log.Printf(`%s: added "%s" to ipset "%s" with %d second(s) timeout`, r.name, ip, s, ds)
		}
	}
	return nil
}

func (b *ipsetBackend) Finalize() error {
	if configuration.SaveFilePath != nil {
		if err := b.saveIpsets(); err != nil {
			return fmt.Errorf(`failed to save ipsets to "%s": %s`, *configuration.SaveFilePath, err)
		}
	}
	if err := b.deleteIpsetsAndIptablesEntries(); err != nil {
		return fmt.Errorf("failed to delete ipsets and ip(6)tables entries: %s", err)
	}
	return nil
}
