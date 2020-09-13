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
	ir := func(f func(r *rule)) {
		r := validRule()
		f(r)
		if err := r.initialize(); err != nil {
			t.Errorf("failed to initialize rule: %s", err)
		}
	}

	ir(func(r *rule) {})
	ir(func(r *rule) {
		r.Action = []string{"log"}
	})
	ir(func(r *rule) {
		r.Source = []string{"systemd", "service"}
	})
}

func TestInvalidRules(t *testing.T) {
	ee := func(s string, f func(r *rule)) {
		r := validRule()
		f(r)
		if err := r.initialize(); err == nil {
			t.Errorf("expected error because of %s", s)
		}
	}

	ee("missing action", func(r *rule) {
		r.Action = []string{}
	})
	ee("unknown action", func(r *rule) {
		r.Action = []string{"unknown"}
	})
	ee("ban action: missing duration parameter", func(r *rule) {
		r.Action = []string{"ban"}
	})
	ee("ban action: invalid duration parameter", func(r *rule) {
		r.Action = []string{"ban", "1hour"}
	})
	ee("invalid host magic", func(r *rule) {
		r.Regexp = "%chost%"
	})
	ee("duplicate host magic", func(r *rule) {
		r.Regexp = "%host% %host%"
	})
	ee("syntactically incorrect regexp", func(r *rule) {
		r.Regexp = "%host% ["
	})
	ee("forbidden subexpression", func(r *rule) {
		r.Regexp = "%host% (?P<host>.*)"
	})
	ee("missing source", func(r *rule) {
		r.Source = []string{}
	})
	ee("unknown source", func(r *rule) {
		r.Source = []string{"unknown"}
	})
	ee("file source: missing path parameter", func(r *rule) {
		r.Source = []string{"file"}
	})
	ee("file source: path is a directory", func(r *rule) {
		r.Source = []string{"file", "/"}
	})
	ee("systemd source: missing service parameter", func(r *rule) {
		r.Source = []string{"systemd"}
	})
}
