package gerberos

import (
	"errors"
	"testing"
	"testing/iotest"
)

func TestConfigurationReadFileInvalid(t *testing.T) {
	rc := func(n string) {
		c := &Configuration{}
		testError(t, c.ReadFile(n))
	}

	rc("")
	rc("test/invalid_configuration.toml")
}

func TestConfigurationReadFileError(t *testing.T) {
	r := iotest.ErrReader(errors.New(""))
	c := &Configuration{}
	testError(t, c.read(r))
}
