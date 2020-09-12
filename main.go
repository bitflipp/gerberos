package main

import (
	"flag"
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

func alreadyRunningInstance() (bool, error) {
	out, err := exec.Command("ps", "axco", "command").Output()
	if err != nil {
		return false, err
	}

	n := path.Base(os.Args[0])
	oc := false
	for _, p := range strings.Split(string(out), "\n") {
		if p == n {
			if oc {
				return true, nil
			} else {
				oc = true
			}
		}
	}

	return false, nil
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

	// Already running instance
	r, err := alreadyRunningInstance()
	if err != nil {
		log.Fatalf("failed to check for an already running instance: %s", err)
	}
	if r {
		log.Fatalf("an instance of gerberos is already running")
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
	defer reset()

	// Initialize rules
	for n, r := range configuration.Rules {
		r.name = n
		if err := r.initialize(); err != nil {
			log.Fatalf(`failed to initialize rule "%s": %s`, n, err)
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
}
