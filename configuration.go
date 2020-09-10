package main

import (
	"os"

	"github.com/BurntSushi/toml"
)

var (
	configuration struct {
		Rules map[string]*rule
	}
)

func readConfigurationFile(p string) error {
	cf, err := os.Open(p)
	if err != nil {
		return err
	}
	defer cf.Close()

	if _, err := toml.DecodeReader(cf, &configuration); err != nil {
		return err
	}

	return nil
}
