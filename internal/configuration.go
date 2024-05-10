package gerberos

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Configuration struct {
	Backend      string
	SaveFilePath string
	LogLevel     string
	Rules        map[string]*rule
}

func (c *Configuration) ReadFile(path string) error {
	cf, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open configuration file: %w", err)
	}
	defer cf.Close()

	return c.read(cf)
}

func (c *Configuration) read(r io.Reader) error {
	if _, err := toml.NewDecoder(r).Decode(&c); err != nil {
		var terr toml.ParseError
		if errors.As(err, &terr) {
			return fmt.Errorf("failed to decode configuration file: %s", terr.ErrorWithUsage())
		}
		return fmt.Errorf("failed to decode configuration file: %w", err)
	}

	c.setGlobalLogLevel()

	return nil
}

func (c *Configuration) setGlobalLogLevel() {
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
