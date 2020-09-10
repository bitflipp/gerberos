package main

import (
	"flag"
	"log"
)

func main() {
	cfp := flag.String("c", "./gerberos.toml", "Path to TOML configuration file")
	flag.Parse()

	if err := readConfigurationFile(*cfp); err != nil {
		log.Fatalf("failed to read configuration file: %s", err)
	}

	for n, r := range configuration.Rules {
		r.name = n
		if err := r.initialize(); err != nil {
			log.Fatalf("failed to initialize rule '%s': %s", n, err)
		}
	}

	for _, r := range configuration.Rules {
		go r.worker()
	}

	select {}
}
