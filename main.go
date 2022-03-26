package main

import (
	"flag"
	"log"
	"runtime/debug"
)

var tag = "?"

func main() {
	// Logging
	log.SetFlags(0)

	// Version
	log.Printf("gerberos %s", tag)
	if bi, ok := debug.ReadBuildInfo(); ok {
		log.Printf("built with: %s", bi.GoVersion)
	} else {
		log.Print("no build info found")
	}

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
