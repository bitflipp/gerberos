package main

import (
	"testing"
	"time"
)

func testNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func newTestConfiguration() (*configuration, error) {
	c := &configuration{}

	return c, c.readFile("test/configuration.toml")
}

func newTestRunner() (*runner, error) {
	c, err := newTestConfiguration()
	if err != nil {
		return nil, err
	}

	return newRunner(c), nil
}

func newTestValidRule() *rule {
	return &rule{
		Action:      []string{"ban", "1h"},
		Regexp:      []string{`%ip%\s%id%`},
		Source:      []string{"test"},
		Aggregate:   []string{"1s", `a\s%id%`, `%id%\sb`},
		Occurrences: []string{"5", "10s"},

		name: "test",
	}
}

func newTestOccurrences() *occurrences {
	return newOccurrences(100*time.Millisecond, 10)
}
