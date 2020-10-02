package main

import (
	"testing"
)

func TestMatches(t *testing.T) {
	// Simple match
	ml := func(s, re string, e bool, l string) *match {
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

	// Aggregate match
	mla := func(s string, e bool, ls ...string) *match {
		r := validRule()

		if err := r.initialize(); err != nil {
			t.Errorf("%s: failed to initialize rule", err)
			t.FailNow()
		}

		for i, l := range ls {
			m, err := r.match(l)
			if i == len(ls)-1 {
				if e != (err == nil) {
					t.Errorf(`%s: unexpected result`, s)
					t.FailNow()
				}
				return m
			}
		}

		return nil
	}

	// IPv4/6
	em := func(s, h string, ipv6 bool) {
		m := ml(s, "%ip%", true, h)
		if h != m.ip {
			t.Errorf(`%s: expected IP "%s", got "%s"`, s, h, m.ip)
		}
		if ipv6 != m.ipv6 {
			t.Errorf("%s: unexpected IPv6 flag", s)
		}
	}

	ml("invalid 4.1", "%ip%", false, "300.300.300.300")
	ml("invalid 4.2", "%ip%", false, "100.100.100")
	ml("invalid 4.3", "%ip%", false, "100..100.100.100")
	ml("invalid 4.4", "start %ip% end", false, "start 1000.100.100.100 end")
	ml("invalid 4.5", "start %ip% end", false, "start 100.100.100.100.100.100 end")
	ml("invalid 4.6", "%ip%", false, "192.168.0.")
	ml("invalid 4.7", "%ip% extra", false, "192.168.1.1")
	mla("invalid 4.8", false, "192.168.1.1 id")
	mla("invalid 4.9", false, "192.168.1.1 id", "192.168.1.1 id")
	mla("invalid 4.10", false, "192.168.1.1 id", "id c")

	ml("invalid 6.1", "%ip%", false, "affe:affe")
	ml("invalid 6.2", "%ip%", false, "1a:1a")
	ml("invalid 6.3", "start %ip% end", false, "start 3ab9:1ea0:c269:5aad:b716:c28d:237d:4d8f:3ab9:1ea0:c269:5aad:b716:c28d:237d:4d8f end")
	mla("invalid 6.4", false, "::1 id")
	mla("invalid 6.5", false, "::1 id", "::1 id")
	mla("invalid 6.6", false, "::1 id", "id c")

	em("valid 4.1", "147.144.139.204", false)
	em("valid 4.2", "49.236.157.198", false)
	em("valid 4.3", "1.1.1.1", false)
	em("valid 4.4", "255.255.255.254", false)
	em("valid 4.5", "0.0.0.0", false)
	em("valid 4.6", "11.0.0.0", false)
	em("valid 4.7", "129.56.0.0", false)
	em("valid 4.8", "243.8.45.0", false)
	em("valid 4.9", "192.168.172.14", false)
	ml("valid 4.10", "prefixseparator%ip%", true, "prefixseparator192.168.1.1")
	ml("valid 4.11", "%ip%separatorsuffix", true, "192.168.1.1separatorsuffix")
	mla("valid 4.12", true, "192.168.1.1 id", "a id")
	mla("valid 4.12", true, "192.168.1.1 id", "id b")

	em("valid 6.1", "a0ca:14f:80b2::77e6:f471:361e", true)
	em("valid 6.2", "35bb:6be1:abae:de1:adbd:aecd:2813:a993", true)
	em("valid 6.3", "3ab9:1ea0:c269:5aad:b716:c28d:237d:4d8f", true)
	em("valid 6.4", "affe::affe", true)
	em("valid 6.5", "1a::1a", true)
	em("valid 6.6", "1200:0000:AB00:1234:0000:2552:7777:1313", true)
	em("valid 6.7", "21DA:D3:0:2F3B:2AA:FF:FE28:9C5A", true)
	em("valid 6.8", "::1", true)
	mla("valid 6.9", true, "::1 id", "a id")
	mla("valid 6.10", true, "::1 id", "id b")
}
