package main

import (
	"flag"
	"log"
	"runtime/debug"
)

var (
	version = "unknown version"
)

func logBuildInfo() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		log.Print("no build info found")
		return
	}

	log.Printf("build info:")
	log.Printf("- built with: %s", bi.GoVersion)
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			l := 7
			if len(s.Value) > 7 {
				s.Value = s.Value[:l]
			}
			log.Printf("- revision: %s", s.Value)
		case "vcs.modified":
			if s.Value == "true" {
				log.Printf("- source files were modified since last commit")
			}
		}
	}
}

func main() {
	// Logging
	log.SetFlags(0)

	// Version and build info
	log.Printf("gerberos %s", version)
	logBuildInfo()

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
