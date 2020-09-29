package main

import (
	"errors"
	"fmt"
	"log"
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

func (r *rule) matchSimple(l string) (*match, error) {
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

func (r *rule) matchAggregate(l string) (*match, error) {
	a := r.aggregate

	for _, re := range a.regexp {
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
		id := sm["id"]

		a.registryMutex.Lock()
		if ip, e := a.registry[id]; e {
			delete(a.registry, id)
			a.registryMutex.Unlock()

			return &match{
				line:   l,
				time:   time.Now(),
				ip:     ip.String(),
				ipv6:   ip.To4() == nil,
				regexp: re,
			}, nil
		}
		a.registryMutex.Unlock()
	}

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
		pip := net.ParseIP(h)
		if pip == nil {
			return nil, fmt.Errorf(`failed to parse matched IP "%s"`, h)
		}

		id := sm["id"]
		if id == "" {
			return nil, fmt.Errorf(`failed to match ID`)
		}

		a.registryMutex.Lock()
		a.registry[id] = pip
		if configuration.Verbose {
			log.Printf(`%s: added ID "%s" to registry with IP %s`, r.name, id, pip)
		}
		a.registryMutex.Unlock()

		go func(id string) {
			time.Sleep(a.interval)
			a.registryMutex.Lock()
			defer a.registryMutex.Unlock()
			if ip, e := a.registry[id]; e {
				delete(a.registry, id)
				log.Printf(`%s: removed ID "%s" with IP %s from registry`, r.name, id, ip)
			}
		}(id)
	}

	return nil, errors.New("line does not match any regexp")
}

func (r *rule) match(l string) (*match, error) {
	if r.aggregate != nil {
		return r.matchAggregate(l)
	}

	return r.matchSimple(l)
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
