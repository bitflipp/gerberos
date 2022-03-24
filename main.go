package main

import (
	_ "embed"
	"flag"
	"log"
	"runtime/debug"
)

//go:embed VERSION
var version string

func main() {
	// Logging
	log.SetFlags(0)

	// Version
	log.Printf("gerberos %s", version)
	if bi, ok := debug.ReadBuildInfo(); ok {
		log.Printf("go version: %s", bi.GoVersion)
		for i := range bi.Settings {
			switch bi.Settings[i].Key {
			case "vcs.revision":
				length := 7
				if length > len(bi.Settings[i].Value) {
					length = len(bi.Settings[i].Value)
				}
				log.Printf("revision: %s", bi.Settings[i].Value[:length])
			case "vcs.modified":
				log.Printf("modified files: %s", bi.Settings[i].Value)
			}
		}
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
