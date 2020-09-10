package main

import (
	"flag"
	"log"
	"os"
	"os/exec"

	"github.com/BurntSushi/toml"
)

const (
	ipset4Name = "gerberos4"
	ipset6Name = "gerberos6"
)

var (
	configuration struct {
		Rules map[string]*rule
	}
)

func main() {
	// Logging
	log.SetFlags(log.Lshortfile &^ (log.Ldate | log.Ltime))

	// Privileges
	if err := exec.Command("ipset", "list").Run(); err != nil {
		log.Fatalf("ipset: insufficient privileges")
	}
	if err := exec.Command("iptables", "-L").Run(); err != nil {
		log.Fatalf("iptables: insufficient privileges")
	}

	// Flags
	cfp := flag.String("c", "./gerberos.toml", "Path to TOML configuration file")
	flag.Parse()

	// Configuration
	cf, err := os.Open(*cfp)
	if err != nil {
		log.Fatalf("failed to open configuration file: %s", err)
	}
	defer cf.Close()

	if _, err := toml.DecodeReader(cf, &configuration); err != nil {
		log.Fatalf("failed to open configuration file: %s", err)
	}

	// ipsets
	if err := exec.Command("ipset", "list", ipset4Name).Run(); err != nil {
		if err := exec.Command("ipset", "create", ipset4Name, "hash:ip", "timeout", "0").Run(); err != nil {
			log.Fatalf("failed to create ipset '%s': %s", ipset4Name, err)
		}
	}
	if err := exec.Command("ipset", "list", ipset6Name).Run(); err != nil {
		if err := exec.Command("ipset", "create", ipset6Name, "hash:ip", "family", "inet6", "timeout", "0").Run(); err != nil {
			log.Fatalf("failed to create ipset '%s': %s", ipset6Name, err)
		}
	}
	if err := exec.Command("iptables", "-C", "INPUT", "-j", "DROP", "-m", "set", "--match-set", ipset4Name, "src").Run(); err != nil {
		if err := exec.Command("iptables", "-A", "INPUT", "-j", "DROP", "-m", "set", "--match-set", ipset4Name, "src").Run(); err != nil {
			log.Fatalf("failed to create iptables entry for set '%s': %s", ipset4Name, err)
		}
	}
	if err := exec.Command("ip6tables", "-C", "INPUT", "-j", "DROP", "-m", "set", "--match-set", ipset6Name, "src").Run(); err != nil {
		if err := exec.Command("ip6tables", "-A", "INPUT", "-j", "DROP", "-m", "set", "--match-set", ipset6Name, "src").Run(); err != nil {
			log.Fatalf("failed to create ip6tables entry for set 'gerberos6': %s", err)
		}
	}

	// Rules
	for n, r := range configuration.Rules {
		r.name = n
		if err := r.initialize(); err != nil {
			log.Fatalf("failed to initialize rule '%s': %s", n, err)
		}
	}

	// Workers
	for _, r := range configuration.Rules {
		go r.worker()
	}

	// Wait indefinitely
	select {}
}
