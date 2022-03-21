package main

import (
	_ "embed"
	"flag"
	"log"
)

//go:embed VERSION
var version string

func main() {
	// Logging
	log.SetFlags(0)

	// Version
	log.Printf("gerberos %s", version)

	// Flags
	cfp := flag.String("c", "./gerberos.toml", "Path to TOML configuration file")
	flag.Parse()

	// Configuration
	c := &configuration{}
	if err := c.readFile(*cfp); err != nil {
		log.Fatalf("failed to read configuration file: %s", err)
	}

	// Runner
	rn := newRunner(c)
	if err := rn.initialize(); err != nil {
		log.Fatalf("failed to initialize runner: %s", err)
	}
	defer func() {
		if err := rn.finalize(); err != nil {
			log.Fatalf("failed to finalize runner: %s", err)
		}
	}()
	rn.run(true)
}
