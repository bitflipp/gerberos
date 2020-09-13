package main

import (
	"testing"
)

func validRule() *rule {
	return &rule{
		Action: []string{"ban", "1h"},
		Regexp: "%host%",
		Source: []string{"file", "FILE"},
	}
}

func TestValidRules(t *testing.T) {
	{
		r := validRule()
		if err := r.initialize(); err != nil {
			t.Errorf("failed to initialize rule: %s", err)
		}
	}
	{
		r := validRule()
		r.Action = []string{"log"}
		r.Source = []string{"systemd", "service"}
		if err := r.initialize(); err != nil {
			t.Errorf("failed to initialize rule: %s", err)
		}
	}
}

func TestInvalidRules(t *testing.T) {
	er := func(s string, f func(r *rule)) {
		r := validRule()
		f(r)
		if err := r.initialize(); err == nil {
			t.Errorf("expected error because of %s", s)
		}
	}

	er("missing action", func(r *rule) {
		r.Action = []string{}
	})
	er("unknown action", func(r *rule) {
		r.Action = []string{"unknown"}
	})
	er("ban action: missing duration parameter", func(r *rule) {
		r.Action = []string{"ban"}
	})
	er("ban action: invalid duration parameter", func(r *rule) {
		r.Action = []string{"ban", "1hour"}
	})
	er("invalid host magic", func(r *rule) {
		r.Regexp = "%chost%"
	})
	er("duplicate host magic", func(r *rule) {
		r.Regexp = "%host% %host%"
	})
	er("syntactically incorrect regexp", func(r *rule) {
		r.Regexp = "%host% ["
	})
	er("forbidden subexpression", func(r *rule) {
		r.Regexp = "%host% (?P<host>.*)"
	})
	er("missing source", func(r *rule) {
		r.Source = []string{}
	})
	er("unknown source", func(r *rule) {
		r.Source = []string{"unknown"}
	})
	er("file source: missing path parameter", func(r *rule) {
		r.Source = []string{"file"}
	})
	er("file source: path is a directory", func(r *rule) {
		r.Source = []string{"file", "/"}
	})
	er("systemd source: missing service parameter", func(r *rule) {
		r.Source = []string{"systemd"}
	})
}
