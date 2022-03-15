package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type configuration struct {
	Verbose      bool
	Backend      string
	SaveFilePath string
	Rules        map[string]*rule
}

func (c *configuration) readFile(fp string) error {
	cf, err := os.Open(fp)
	if err != nil {
		return fmt.Errorf("failed to open configuration file: %w", err)
	}
	defer cf.Close()

	if _, err := toml.NewDecoder(cf).Decode(&c); err != nil {
		var terr toml.ParseError
		if errors.As(err, &terr) {
			return fmt.Errorf("failed to decode configuration file: %s", terr.ErrorWithUsage())
		}
		return fmt.Errorf("failed to decode configuration file: %w", err)
	}

	return nil
}
