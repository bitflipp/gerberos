package main

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"time"
)

type match struct {
	time   time.Time
	line   string
	host   string
	ipv6   bool
	regexp *regexp.Regexp
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
			line:   l,
			time:   time.Now(),
			host:   h,
			ipv6:   ph.To4() == nil,
			regexp: re,
		}, nil
	}

	return nil, errors.New("line does not match any regexp")
}

func (m match) String() string {
	return m.StringSimple()
}

func (m match) StringSimple() string {
	ipv := "IPv4"
	if m.ipv6 {
		ipv = "IPv6"
	}

	return fmt.Sprintf("time = %s, host = %s, %s", m.time.Format(time.RFC3339), m.host, ipv)
}

func (m match) StringExtended() string {
	return fmt.Sprintf("%s, line = %s, regexp = %s", m, m.line, m.regexp)
}
