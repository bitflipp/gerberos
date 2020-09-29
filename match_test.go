package main

import (
	"testing"
)

func TestMatches(t *testing.T) {
	ml := func(s, l, re string, e bool) *match {
		r := validRule()
		r.Aggregate = nil
		r.Regexp = []string{re}
		if err := r.initialize(); err != nil {
			t.Errorf("%s: failed to initialize rule", err)
			t.FailNow()
		}

		m, err := r.match(l)
		if e != (err == nil) {
			t.Errorf(`%s: unexpected result`, s)
			t.FailNow()
		}

		return m
	}

	em := func(s, h string, ipv6 bool) {
		m := ml(s, h, "%ip%", true)
		if h != m.ip {
			t.Errorf(`%s: expected IP "%s", got "%s"`, s, h, m.ip)
		}
		if ipv6 != m.ipv6 {
			t.Errorf("%s: unexpected IPv6 flag", s)
		}
	}

	ml("invalid 4.1", "300.300.300.300", "%ip%", false)
	ml("invalid 4.2", "100.100.100", "%ip%", false)
	ml("invalid 4.3", "100..100.100.100", "%ip%", false)
	ml("invalid 4.4", "start 1000.100.100.100 end", "start %ip% end", false)
	ml("invalid 4.5", "start 100.100.100.100.100.100 end", "start %ip% end", false)
	ml("invalid 4.6", "192.168.0.", "%ip%", false)
	ml("invalid 4.7", "192.168.1.1", "%ip% extra", false)

	ml("invalid 6.1", "affe:affe", "%ip%", false)
	ml("invalid 6.2", "1a:1a", "%ip%", false)
	ml("invalid 6.3", "start 3ab9:1ea0:c269:5aad:b716:c28d:237d:4d8f:3ab9:1ea0:c269:5aad:b716:c28d:237d:4d8f end", "start %ip% end", false)

	em("valid 4.1", "147.144.139.204", false)
	em("valid 4.2", "49.236.157.198", false)
	em("valid 4.3", "1.1.1.1", false)
	em("valid 4.4", "255.255.255.254", false)
	em("valid 4.5", "0.0.0.0", false)
	em("valid 4.6", "11.0.0.0", false)
	em("valid 4.7", "129.56.0.0", false)
	em("valid 4.8", "243.8.45.0", false)
	em("valid 4.9", "192.168.172.14", false)
	ml("valid 4.10", "prefix 192.168.1.1", "prefix.*?%ip%", true)
	ml("valid 4.11", "192.168.1.1 suffix", "%ip%.*?suffix", true)

	em("valid 6.1", "a0ca:14f:80b2::77e6:f471:361e", true)
	em("valid 6.2", "35bb:6be1:abae:de1:adbd:aecd:2813:a993", true)
	em("valid 6.3", "3ab9:1ea0:c269:5aad:b716:c28d:237d:4d8f", true)
	em("valid 6.4", "affe::affe", true)
	em("valid 6.5", "1a::1a", true)
	em("valid 6.6", "1200:0000:AB00:1234:0000:2552:7777:1313", true)
	em("valid 6.7", "21DA:D3:0:2F3B:2AA:FF:FE28:9C5A", true)
	em("valid 6.8", "::1", true)
}
