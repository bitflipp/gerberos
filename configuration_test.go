package main

import (
	"errors"
	"testing"
	"testing/iotest"
)

func TestConfigurationReadFileInvalid(t *testing.T) {
	rc := func(n string) {
		c := &configuration{}
		testError(t, c.readFile(n))
	}

	rc("")
	rc("test/invalid_configuration.toml")
}

func TestConfigurationReadFileError(t *testing.T) {
	r := iotest.ErrReader(errors.New(""))
	c := &configuration{}
	testError(t, c.read(r))
}
