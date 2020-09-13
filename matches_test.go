package main

import (
	"testing"
)

func TestMatches(t *testing.T) {
	ml := func(s, l, re string, e bool) *match {
		r := validRule()
		r.Regexp = re
		if err := r.initialize(); err != nil {
			t.Errorf("failed to initialize rule: %s", err)
			t.FailNow()
		}

		m, err := newMatch(r, l)
		if e != (err == nil) {
			t.Errorf(`unexpected result for %s`, s)
		}

		return m
	}

	em := func(s, h string, ipv6 bool) {
		m := ml(s, h, "%host%", true)
		if h != m.host {
			t.Errorf(`expected host "%s", got "%s"`, h, m.host)
		}
		if ipv6 != m.ipv6 {
			t.Errorf("unexpected IPv6 flag")
		}
	}

	ml("invalid 1", "192.168.0.", "%host%", false)
	ml("invalid 2", "192.168.1.1", "%host% extra", false)
	ml("invalid 3", "1200:0000:AB00:1234:O000:2552:7777:1313", "%host%", false)

	em("valid 1", "0.0.0.0", false)
	em("valid 2", "11.0.0.0", false)
	em("valid 3", "129.56.0.0", false)
	em("valid 4", "243.8.45.0", false)
	em("valid 5", "192.168.172.14", false)
	em("valid 6", "1200:0000:AB00:1234:0000:2552:7777:1313", true)
	em("valid 7", "21DA:D3:0:2F3B:2AA:FF:FE28:9C5A", true)
	em("valid 8", "::1", true)
}
