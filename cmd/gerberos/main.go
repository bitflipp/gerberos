package main

import (
	"flag"
	"runtime/debug"

	gerberos "github.com/bitflipp/gerberos/internal"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	version = "unknown version"
)

func logVersionAndBuildInfo() {
	ev := log.Info().Str("version", version)

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		ev.Msg("no build info found")
		return
	}

	ev = ev.Str("goVersion", bi.GoVersion)
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			l := 7
			if len(s.Value) > 7 {
				s.Value = s.Value[:l]
			}
			ev = ev.Str("revision", s.Value)
		case "vcs.modified":
			ev = ev.Bool("sourceFilesModified", s.Value == "true")
		}
	}

	ev.Msg("")
}

func setGlobalLogLevel(c *gerberos.Configuration) {
	switch c.LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		log.Warn().Str("logLevel", c.LogLevel).Msg("unknown log level, defaulting to info")
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func main() {
	cfp := flag.String("c", "./gerberos.toml", "Path to TOML configuration file")
	flag.Parse()

	c := &gerberos.Configuration{}
	if err := c.ReadFile(*cfp); err != nil {
		log.Fatal().Err(err).Msg("failed to read configuration file")
	}

	setGlobalLogLevel(c)
	logVersionAndBuildInfo()

	rn := gerberos.NewRunner(c)
	if err := rn.Initialize(); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize runner")
	}
	defer func() {
		if err := rn.Finalize(); err != nil {
			log.Fatal().Err(err).Msg("failed to finalize runner")
		}
	}()
	rn.Run(true)
}
