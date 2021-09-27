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
	version = "2.2.0"
)

var (
	configuration struct {
		Backend      string
		SaveFilePath *string
		Verbose      bool
		Rules        map[string]*rule
	}
	activeBackend     backend
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

	// Initialize backend
	switch configuration.Backend {
	case "":
		log.Fatalf("no backend configuration found (is it missing?) - please choose one of the available backends")
	case "ipset":
		activeBackend = &ipsetBackend{}
	case "nft":
		activeBackend = &nftBackend{}
	default:
		log.Fatalf("unknown backend: %s", configuration.Backend)
	}
	if err := activeBackend.Initialize(); err != nil {
		log.Fatalf("failed to initialize backend: %s", err)
	}
	defer func() {
		if err := activeBackend.Finalize(); err != nil {
			log.Printf("failed to finalize backend: %s", err)
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
