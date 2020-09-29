package main

import (
	"testing"
)

func validRule() *rule {
	return &rule{
		Action:      []string{"ban", "1h"},
		Regexp:      []string{"%host%"},
		Source:      []string{"file", "FILE"},
		Occurrences: []string{"5", "10s"},
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
		r.Occurrences = nil
	})
	ir(func(r *rule) {
		r.Action = []string{"log", "extended"}
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
		r.Action = nil
	})
	ee("empty action", func(r *rule) {
		r.Action = []string{}
	})
	ee("unknown action", func(r *rule) {
		r.Action = []string{"unknown"}
	})
	ee("log action: missing level parameter", func(r *rule) {
		r.Action = []string{"log"}
	})
	ee("log action: invalid level parameter", func(r *rule) {
		r.Action = []string{"log", "invalid"}
	})
	ee("log action: superfluous parameter", func(r *rule) {
		r.Action = []string{"log", "simple", "superfluous"}
	})
	ee("ban action: missing duration parameter", func(r *rule) {
		r.Action = []string{"ban"}
	})
	ee("ban action: invalid duration parameter", func(r *rule) {
		r.Action = []string{"ban", "1hour"}
	})
	ee("ban action: superfluous parameter", func(r *rule) {
		r.Action = []string{"ban", "1h", "superfluous"}
	})
	ee("missing regexp", func(r *rule) {
		r.Regexp = nil
	})
	ee("empty regexp", func(r *rule) {
		r.Regexp = []string{}
	})
	ee("invalid host magic", func(r *rule) {
		r.Regexp = []string{"%chost%"}
	})
	ee("duplicate host magic", func(r *rule) {
		r.Regexp = []string{"%host% %host%"}
	})
	ee("syntactically incorrect regexp", func(r *rule) {
		r.Regexp = []string{"%host% ["}
	})
	ee("forbidden subexpression", func(r *rule) {
		r.Regexp = []string{"%host% (?P<host>.*)"}
	})
	ee("missing source", func(r *rule) {
		r.Source = nil
	})
	ee("empty source", func(r *rule) {
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
	ee("file source source: superfluous parameter", func(r *rule) {
		r.Source = []string{"file", "file", "superfluous"}
	})
	ee("systemd source: missing service parameter", func(r *rule) {
		r.Source = []string{"systemd"}
	})
	ee("systemd source: superfluous parameter", func(r *rule) {
		r.Source = []string{"systemd", "service", "superfluous"}
	})
	ee("kernel source: superfluous parameter", func(r *rule) {
		r.Source = []string{"kernel", "superfluous"}
	})
	ee("occurrences: missing count parameter", func(r *rule) {
		r.Occurrences = []string{}
	})
	ee("occurrences: invalid count parameter", func(r *rule) {
		r.Occurrences = []string{"five"}
	})
	ee("occurrences: invalid count parameter 2", func(r *rule) {
		r.Occurrences = []string{"1"}
	})
	ee("occurrences: missing interval parameter", func(r *rule) {
		r.Occurrences = []string{"5"}
	})
	ee("occurrences: invalid interval parameter", func(r *rule) {
		r.Occurrences = []string{"5", "5g"}
	})
}
