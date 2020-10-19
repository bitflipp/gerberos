package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/BurntSushi/toml"
)

const (
	version    = "2.1"
	chainName  = "gerberos"
	ipset4Name = "gerberos4"
	ipset6Name = "gerberos6"
)

var (
	configuration struct {
		SaveFilePath *string
		Verbose      bool
		Rules        map[string]*rule
	}
	respawnWorkerChan = make(chan *rule, 1)
)

func execute(n string, args ...string) (string, int, error) {
	cmd := exec.Command(n, args...)
	if configuration.Verbose {
		log.Printf("executing: %s", cmd)
	}

	b, err := cmd.CombinedOutput()
	if err != nil {
		eerr, ok := err.(*exec.ExitError)
		if ok && eerr != nil {
			return string(b), eerr.ExitCode(), eerr
		}
		return "", -1, err
	}

	return string(b), 0, nil
}

func saveIpsets() error {
	f, err := os.Create(*configuration.SaveFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command("ipset", "save")
	cmd.Stdout = f

	return cmd.Run()
}

func restoreIpsets() error {
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

func deleteIpsetsAndIptablesEntries() error {
	if s, ec, _ := execute("iptables", "-D", chainName, "-j", "DROP", "-m", "set", "--match-set", ipset4Name, "src"); ec > 2 {
		return fmt.Errorf(`failed to delete iptables entry for set "%s": %s`, ipset4Name, s)
	}
	if s, ec, _ := execute("iptables", "-D", "INPUT", "-j", chainName); ec > 2 {
		return fmt.Errorf(`failed to delete iptables entry for chain "%s": %s`, chainName, s)
	}
	if s, ec, _ := execute("iptables", "-X", chainName); ec > 2 {
		return fmt.Errorf(`failed to delete iptables chain "%s": %s`, chainName, s)
	}
	if s, ec, _ := execute("ip6tables", "-D", chainName, "-j", "DROP", "-m", "set", "--match-set", ipset6Name, "src"); ec > 2 {
		return fmt.Errorf(`failed to delete ip6tables entry for set "%s": %s`, ipset6Name, s)
	}
	if s, ec, _ := execute("ip6tables", "-D", "INPUT", "-j", chainName); ec > 2 {
		return fmt.Errorf(`failed to delete ip6tables entry for chain "%s": %s`, chainName, s)
	}
	if s, ec, _ := execute("ip6tables", "-X", chainName); ec > 2 {
		return fmt.Errorf(`failed to delete ip6tables chain "%s": %s`, chainName, s)
	}
	if s, ec, _ := execute("ipset", "destroy", ipset4Name); ec > 1 {
		return fmt.Errorf(`failed to destroy ipset "%s": %s`, ipset4Name, s)
	}
	if s, ec, _ := execute("ipset", "destroy", ipset6Name); ec > 1 {
		return fmt.Errorf(`failed to destroy ipset "%s": %s`, ipset6Name, s)
	}

	return nil
}

func createIpsets() error {
	if s, ec, _ := execute("ipset", "create", ipset4Name, "hash:ip", "timeout", "0"); ec != 0 {
		return fmt.Errorf(`failed to create ipset "%s": %s`, ipset4Name, s)
	}
	if s, ec, _ := execute("ipset", "create", ipset6Name, "hash:ip", "family", "inet6", "timeout", "0"); ec != 0 {
		return fmt.Errorf(`failed to create ipset "%s": %s`, ipset6Name, s)
	}

	return nil
}

func createIptablesEntries() error {
	if s, ec, _ := execute("iptables", "-N", chainName); ec != 0 {
		return fmt.Errorf(`failed to create iptables chain "%s": %s`, chainName, s)
	}
	if s, ec, _ := execute("iptables", "-I", chainName, "-j", "DROP", "-m", "set", "--match-set", ipset4Name, "src"); ec != 0 {
		return fmt.Errorf(`failed to create iptables entry for set "%s": %s`, ipset4Name, s)
	}
	if s, ec, _ := execute("iptables", "-I", "INPUT", "-j", chainName); ec != 0 {
		return fmt.Errorf(`failed to create iptables entry for chain "%s": %s`, chainName, s)
	}
	if s, ec, _ := execute("ip6tables", "-N", chainName); ec != 0 {
		return fmt.Errorf(`failed to create ip6tables chain "%s": %s`, chainName, s)
	}
	if s, ec, _ := execute("ip6tables", "-I", chainName, "-j", "DROP", "-m", "set", "--match-set", ipset6Name, "src"); ec != 0 {
		return fmt.Errorf(`failed to create ip6tables entry for set "%s": %s`, ipset6Name, s)
	}
	if s, ec, _ := execute("ip6tables", "-I", "INPUT", "-j", chainName); ec != 0 {
		return fmt.Errorf(`failed to create ip6tables entry for chain "%s": %s`, chainName, s)
	}

	return nil
}

func isInstanceAlreadyRunning() (bool, error) {
	s, _, err := execute("ps", "axco", "command")
	if err != nil {
		return false, err
	}

	n := path.Base(os.Args[0])
	oc := false
	for _, p := range strings.Split(s, "\n") {
		if p == n {
			if oc {
				return true, nil
			}
			oc = true
		}
	}

	return false, nil
}

func spawnWorker(r *rule) {
	go r.worker()
	log.Printf("%s: spawned worker", r.name)
}

func main() {
	// Logging
	log.SetFlags(0)

	log.Printf("gerberos %s", version)

	// Check privileges
	if _, _, err := execute("ipset", "list"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			log.Fatalf("ipset: command not found")
		}
		log.Fatalf("ipset: insufficient privileges")
	}
	if _, _, err := execute("iptables", "-L"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			log.Fatalf("iptables: command not found")
		}
		log.Fatalf("iptables: insufficient privileges")
	}
	if _, _, err := execute("ip6tables", "-L"); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			log.Fatalf("ip6tables: command not found")
		}
		log.Fatalf("ip6tables: insufficient privileges")
	}

	// Parse flags
	cfp := flag.String("c", "./gerberos.toml", "Path to TOML configuration file")
	flag.Parse()

	// Read configuration file
	cf, err := os.Open(*cfp)
	if err != nil {
		log.Fatalf("failed to open configuration file: %s", err)
	}
	defer cf.Close()
	if _, err := toml.DecodeReader(cf, &configuration); err != nil {
		log.Fatalf("failed to read configuration file: %s", err)
	}

	// Already running instance
	r, err := isInstanceAlreadyRunning()
	if err != nil {
		log.Fatalf("failed to check for an already running instance: %s", err)
	}
	if r {
		log.Fatalf("an instance is already running")
	}

	// Initialize ipsets and ip(6)tables entries
	if err := deleteIpsetsAndIptablesEntries(); err != nil {
		log.Fatalf("failed to delete ipsets: %s", err)
	}
	if configuration.SaveFilePath != nil {
		if err := restoreIpsets(); err != nil {
			log.Printf(`failed to restore ipsets from "%s": %s`, *configuration.SaveFilePath, err)
			if err := createIpsets(); err != nil {
				log.Fatalf("failed to create ipsets: %s", err)
			}
		} else {
			log.Printf(`restored ipsets from "%s"`, *configuration.SaveFilePath)
		}
	} else {
		log.Printf("warning: not persisting ipsets")
		if err := createIpsets(); err != nil {
			log.Fatalf("failed to create ipsets: %s", err)
		}
	}
	if err := createIptablesEntries(); err != nil {
		log.Fatalf("failed to create ip(6)tables entries: %s", err)
	}
	defer func() {
		if configuration.SaveFilePath != nil {
			if err := saveIpsets(); err != nil {
				log.Printf(`failed to save ipsets to "%s": %s`, *configuration.SaveFilePath, err)
			}
		}
		if err := deleteIpsetsAndIptablesEntries(); err != nil {
			log.Printf("failed to delete ipsets and ip(6)tables entries: %s", err)
		}
	}()

	// Initialize rules
	for n, r := range configuration.Rules {
		r.name = n
		if err := r.initialize(); err != nil {
			log.Fatalf(`failed to initialize rule "%s": %s`, n, err)
		}
	}

	// Spawn workers
	for _, r := range configuration.Rules {
		spawnWorker(r)
	}

	// Wait for SIGINT or SIGTERM and respawn workers
	ic := make(chan os.Signal, 1)
	signal.Notify(ic, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ic:
			return
		case r := <-respawnWorkerChan:
			spawnWorker(r)
		}
	}
}
