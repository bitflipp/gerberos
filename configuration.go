package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type configuration struct {
	Verbose      bool
	Backend      string
	SaveFilePath *string
	Rules        map[string]*rule
}

func (c *configuration) readFile(fp string) error {
	cf, err := os.Open(fp)
	if err != nil {
		return fmt.Errorf("failed to open configuration file: %w", err)
	}
	defer cf.Close()

	if _, err := toml.DecodeReader(cf, &c); err != nil {
		return fmt.Errorf("failed to decode configuration file: %w", err)
	}

	return nil
}
