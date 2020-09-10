package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
)

const (
	chainName  = "gerberos"
	ipset4Name = "gerberos4"
	ipset6Name = "gerberos6"
)

var (
	configuration struct {
		Rules map[string]*rule
	}
)

func reset() {
	exec.Command("iptables", "-D", chainName, "-j", "DROP", "-m", "set", "--match-set", ipset4Name, "src").Run()
	exec.Command("iptables", "-D", "INPUT", "-j", chainName).Run()
	exec.Command("ip6tables", "-D", chainName, "-j", "DROP", "-m", "set", "--match-set", ipset6Name, "src").Run()
	exec.Command("ip6tables", "-D", "INPUT", "-j", chainName).Run()
	exec.Command("ipset", "destroy", ipset4Name).Run()
	exec.Command("ipset", "destroy", ipset6Name).Run()
}

func main() {
	// Logging
	log.SetFlags(0)

	// Check privileges
	if err := exec.Command("ipset", "list").Run(); err != nil {
		log.Fatalf("ipset: insufficient privileges")
	}
	if err := exec.Command("iptables", "-L").Run(); err != nil {
		log.Fatalf("iptables: insufficient privileges")
	}
	if err := exec.Command("ip6tables", "-L").Run(); err != nil {
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

	// Create ipsets and ip(6)tables entries
	reset()
	exec.Command("ipset", "create", ipset4Name, "hash:ip", "timeout", "0").Run()
	exec.Command("ipset", "create", ipset6Name, "hash:ip", "family", "inet6", "timeout", "0").Run()
	exec.Command("iptables", "-N", chainName).Run()
	exec.Command("iptables", "-I", chainName, "-j", "DROP", "-m", "set", "--match-set", ipset4Name, "src").Run()
	exec.Command("iptables", "-I", "INPUT", "-j", chainName).Run()
	exec.Command("ip6tables", "-N", chainName).Run()
	exec.Command("ip6tables", "-I", chainName, "-j", "DROP", "-m", "set", "--match-set", ipset6Name, "src").Run()
	exec.Command("ip6tables", "-I", "INPUT", "-j", chainName).Run()

	// Initialize rules
	for n, r := range configuration.Rules {
		r.name = n
		if err := r.initialize(); err != nil {
			log.Fatalf("failed to initialize rule '%s': %s", n, err)
		}
	}

	// Spawn workers
	for _, r := range configuration.Rules {
		go r.worker()
	}

	// Wait for SIGINT or SIGTERM
	ic := make(chan os.Signal, 1)
	signal.Notify(ic, syscall.SIGINT, syscall.SIGTERM)
	<-ic

	reset()
}
