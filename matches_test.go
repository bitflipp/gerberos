package main

import (
	"testing"
)

func TestMatches(t *testing.T) {
	ml := func(s, l, re string, e bool) *match {
		r := validRule()
		r.Regexp = []string{re}
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

	ml("invalid 4.1", "300.300.300.300", "%host%", false)
	ml("invalid 4.2", "100.100.100", "%host%", false)
	ml("invalid 4.3", "100..100.100.100", "%host%", false)
	ml("invalid 4.4", "start 1000.100.100.100 end", "start %host% end", false)
	ml("invalid 4.5", "start 100.100.100.100.100.100 end", "start %host% end", false)
	ml("invalid 6.1", "affe:affe", "%host%", false)
	ml("invalid 6.2", "1a:1a", "%host%", false)
	ml("invalid 6.3", "start 3ab9:1ea0:c269:5aad:b716:c28d:237d:4d8f:3ab9:1ea0:c269:5aad:b716:c28d:237d:4d8f end", "start %host% end", false)

	ml("valid 4.1", "147.144.139.204", "%host%", true)
	ml("valid 4.2", "49.236.157.198", "%host%", true)
	ml("valid 4.3", "1.1.1.1", "%host%", true)
	ml("valid 4.4", "255.255.255.254", "%host%", true)
	ml("valid 6.1", "a0ca:14f:80b2::77e6:f471:361e", "%host%", true)
	ml("valid 6.2", "35bb:6be1:abae:de1:adbd:aecd:2813:a993", "%host%", true)
	ml("valid 6.3", "3ab9:1ea0:c269:5aad:b716:c28d:237d:4d8f", "%host%", true)
	ml("valid 6.4", "affe::affe", "%host%", true)
	ml("valid 6.5", "1a::1a", "%host%", true)

	em("valid 2.1", "0.0.0.0", false)
	em("valid 2.2", "11.0.0.0", false)
	em("valid 2.3", "129.56.0.0", false)
	em("valid 2.4", "243.8.45.0", false)
	em("valid 2.5", "192.168.172.14", false)
	em("valid 2.6", "1200:0000:AB00:1234:0000:2552:7777:1313", true)
	em("valid 2.7", "21DA:D3:0:2F3B:2AA:FF:FE28:9C5A", true)
	em("valid 2.8", "::1", true)
}
