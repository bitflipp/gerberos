package gerberos

import (
	"testing"
)

func TestRulesValue(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)

	ir := func(f func(r *rule)) {
		r := newTestValidRule()
		f(r)
		if err := r.initialize(rn); err != nil {
			t.Errorf("failed to initialize rule: %s", err)
		}
	}

	ir(func(r *rule) {})
	ir(func(r *rule) {
		r.Aggregate = nil
	})
	ir(func(r *rule) {
		r.Occurrences = nil
	})
	ir(func(r *rule) {
		r.Action = []string{"log", "extended"}
	})
	ir(func(r *rule) {
		r.Source = []string{"systemd", "service"}
	})
	ir(func(r *rule) {
		r.Source = []string{"file", "FILE"}
	})
	ir(func(r *rule) {
		r.Source = []string{"kernel"}
	})
	ir(func(r *rule) {
		r.Source = []string{"process", "kek", "se"}
	})
}

func TestRulesInvalid(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)

	ee := func(s string, f func(r *rule)) {
		r := newTestValidRule()
		f(r)
		if err := r.initialize(rn); err == nil {
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
	ee("log action: missing type parameter", func(r *rule) {
		r.Action = []string{"log"}
	})
	ee("log action: invalid type parameter", func(r *rule) {
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
	ee("invalid IP magic", func(r *rule) {
		r.Regexp = []string{"%cip%"}
	})
	ee("duplicate IP magic", func(r *rule) {
		r.Regexp = []string{"%ip% %ip%"}
	})
	ee("aggregate option used but no ID magic", func(r *rule) {
		r.Regexp = []string{"%ip%"}
	})
	ee("syntactically incorrect regexp", func(r *rule) {
		r.Regexp = []string{"%ip% %id% ["}
	})
	ee(`forbidden subexpression "ip"`, func(r *rule) {
		r.Regexp = []string{"%ip% (?P<ip>.*)"}
	})
	ee(`forbidden subexpression "id"`, func(r *rule) {
		r.Regexp = []string{"%id% (?P<id>.*)"}
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
	ee("process source: missing name", func(r *rule) {
		r.Source = []string{"process"}
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
	ee("aggregate: missing interval parameter", func(r *rule) {
		r.Aggregate = []string{}
	})
	ee("aggregate: invalid interval parameter", func(r *rule) {
		r.Aggregate = []string{"5g"}
	})
	ee("aggregate: missing regexp", func(r *rule) {
		r.Aggregate = []string{"1m"}
	})
	ee("aggregate: missing ID magic", func(r *rule) {
		r.Aggregate = []string{"1m", "regexp"}
	})
	ee(`aggregate: forbidden subexpression "id"`, func(r *rule) {
		r.Aggregate = []string{"1m", "%id% (?P<id>.*)"}
	})
	ee("aggregate: duplicate ID magic", func(r *rule) {
		r.Aggregate = []string{"1m", "%id% %id%"}
	})
	ee("aggregate: syntactically incorrect regexp", func(r *rule) {
		r.Aggregate = []string{"1m", "%ip% %id% ["}
	})
	ee("occurrences: missing interval parameter", func(r *rule) {
		r.Occurrences = []string{"5"}
	})
	ee("occurrences: invalid interval parameter", func(r *rule) {
		r.Occurrences = []string{"5", "5g"}
	})
}
