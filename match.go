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
	ip     string
	ipv6   bool
	id     string
	regexp *regexp.Regexp
}

func (r *rule) match(l string) (*match, error) {
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
		h := sm["ip"]
		ph := net.ParseIP(h)
		if ph == nil {
			return nil, fmt.Errorf(`failed to parse matched IP "%s"`, h)
		}

		return &match{
			line:   l,
			time:   time.Now(),
			ip:     h,
			ipv6:   ph.To4() == nil,
			regexp: re,
		}, nil
	}

	return nil, errors.New("line does not match any regexp")
}

func (m match) stringSimple() string {
	ipv := "IPv4"
	if m.ipv6 {
		ipv = "IPv6"
	}

	return fmt.Sprintf(`time = %s, IP = "%s", %s`, m.time.Format(time.RFC3339), m.ip, ipv)
}

func (m match) stringExtended() string {
	return fmt.Sprintf(`%s, line = "%s", regexp = "%s"`, m, m.line, m.regexp)
}

func (m match) String() string {
	return m.stringSimple()
}
