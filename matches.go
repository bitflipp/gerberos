package main

import (
	"errors"
	"fmt"
	"net"
	"time"
)

type match struct {
	time time.Time
	line string
	host string
	ipv6 bool
}

func newMatch(r *rule, l string) (*match, error) {
	for _, re := range r.regexp {
		m := re.FindStringSubmatch(l)
		if len(m) == 0 {
			continue
		}

		sm := make(map[string]string)
		for i, name := range re.SubexpNames() {
			if i != 0 && name != "" {
				sm[name] = m[i]
			}
		}
		h := sm["host"]
		ph := net.ParseIP(h)
		if ph == nil {
			return nil, fmt.Errorf(`failed to parse matched host "%s"`, h)
		}

		return &match{
			line: l,
			time: time.Now(),
			host: h,
			ipv6: ph.To4() == nil,
		}, nil
	}

	return nil, errors.New("line does not match any regexp")
}

func (m match) String() string {
	ipv := "IPv4"
	if m.ipv6 {
		ipv = "IPv6"
	}

	return fmt.Sprintf("%s, %s, %s", m.time.Format(time.RFC3339), m.host, ipv)
}
