package main

import "testing"

func TestConfigurationReadFileInvalid(t *testing.T) {
	rc := func(n string) {
		c := &configuration{}
		if err := c.readFile(n); err == nil {
			t.Error("expected error")
		}
	}

	rc("")
	rc("test/invalidConfiguration.toml")
}
