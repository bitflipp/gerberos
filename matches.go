package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type match struct {
	time time.Time
	line string
	host string
	ipv6 bool
}

func newMatch(r *rule, l string) (*match, error) {
	m := r.regexp.FindStringSubmatch(l)
	if len(m) == 0 {
		return nil, errors.New("line does not match regular expression")
	}

	sm := make(map[string]string)
	for i, name := range r.regexp.SubexpNames() {
		if i != 0 && name != "" {
			sm[name] = m[i]
		}
	}
	host := sm["host"]
	ipv6 := strings.Contains(host, ":")

	return &match{line: l, time: time.Now(), host: host, ipv6: ipv6}, nil
}

func (m match) String() string {
	ipv := "IPv4"
	if m.ipv6 {
		ipv = "IPv6"
	}

	return fmt.Sprintf("%s, %s, %s", m.time.Format(time.RFC3339), m.host, ipv)
}
