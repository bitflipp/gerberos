package main

import (
	"errors"
	"regexp"
	"time"
)

var (
	// https://gist.github.com/syzdek/6086792
	hostRegexp4 = regexp.MustCompile(`((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])`)
	hostRegexp6 = regexp.MustCompile(`(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`)
)

type entry struct {
	line string
	host string
	ipv6 bool
	time time.Time
	err  error
}

func newEntry(l string, t time.Time) entry {
	var host string
	ipv6 := false

	m4 := hostRegexp4.FindAllString(l, 2)
	if m4 == nil {
		m6 := hostRegexp6.FindAllString(l, 2)
		if m6 == nil {
			return entry{err: errors.New("no IPv4 or IPv6 host")}
		} else {
			if len(m6) > 1 {
				return entry{err: errors.New("multiple IPv6 hosts")}
			}
			host = m6[0]
			ipv6 = true
		}
	} else {
		if len(m4) > 1 {
			return entry{err: errors.New("multiple IPv4 hosts")}
		}
		host = m4[0]
	}

	return entry{line: l, time: t, host: host, ipv6: ipv6}
}
