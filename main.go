package main

import (
	"flag"
	"io"
	"log"
	"os"
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
	// Flags
	cfp := flag.String("c", "./gerberos.toml", "Path to TOML configuration file")
	flag.Parse()

	// Configuration
	c := &configuration{}
	if err := c.readFile(*cfp); err != nil {
		log.Fatalf("failed to read configuration file: %s", err)
	}

	// Logging
	log.SetFlags(0)
	if c.LogFilePath != "" {
		lf, err := os.OpenFile(c.LogFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed to open log file: %s", err)
		}
		defer lf.Close()
		lw := logWriter{
			clock:  &realTimeClock{},
			writer: lf,
		}
		log.SetOutput(io.MultiWriter(os.Stderr, lw))
	}

	// Version and build info
	log.Printf("gerberos %s", version)
	logBuildInfo()

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
