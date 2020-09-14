package main

import (
	"log"
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
			t.FailNow()
		}

		return m
	}

	em := func(s, h string, ipv6 bool) {
		m := ml(s, h, "%host%", true)
		log.Println("m is", m)
		if h != m.host {
			t.Errorf(`expected host "%s", got "%s"`, h, m.host)
		}
		if ipv6 != m.ipv6 {
			t.Errorf("unexpected IPv6 flag")
		}
	}

	ml("invalid 1", "192.168.0.", "%host%", false)
	ml("invalid 2", "192.168.1.1", "%host% extra", false)

	if m := ml("valid 1.1", "prefix 192.168.1.1", "prefix.*?%host%", true); m.host != "192.168.1.1" {
		t.Errorf(`expected host "192.168.1.1", got "%s"`, m.host)
	}
	if m := ml("valid 1.2", "192.168.1.1 suffix", "%host%.*?suffix", true); m.host != "192.168.1.1" {
		t.Errorf(`expected host "192.168.1.1", got "%s"`, m.host)
	}

	em("valid 2.1", "0.0.0.0", false)
	em("valid 2.2", "11.0.0.0", false)
	em("valid 2.3", "129.56.0.0", false)
	em("valid 2.4", "243.8.45.0", false)
	em("valid 2.5", "192.168.172.14", false)
	em("valid 2.6", "1200:0000:AB00:1234:0000:2552:7777:1313", true)
	em("valid 2.7", "21DA:D3:0:2F3B:2AA:FF:FE28:9C5A", true)
	em("valid 2.8", "::1", true)
}
