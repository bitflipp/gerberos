package main

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type match struct {
	time   time.Time
	line   string
	ip     string
	ipv6   bool
	regexp *regexp.Regexp
}

func (r *rule) matchSimple(line string) (*match, error) {
	for _, re := range r.regexp {
		m := re.FindStringSubmatch(line)
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
		h = strings.Trim(h, "[]")
		ph := net.ParseIP(h)
		if ph == nil {
			return nil, fmt.Errorf(`failed to parse matched IP "%s"`, h)
		}

		return &match{
			line:   line,
			time:   time.Now(),
			ip:     h,
			ipv6:   ph.To4() == nil,
			regexp: re,
		}, nil
	}

	return nil, fmt.Errorf(`line "%s" does not match any regexp`, line)
}

func (r *rule) matchAggregate(line string) (*match, error) {
	a := r.aggregate

	for _, re := range a.regexp {
		m := re.FindStringSubmatch(line)
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
				line:   line,
				time:   time.Now(),
				ip:     ip.String(),
				ipv6:   ip.To4() == nil,
				regexp: re,
			}, nil
		}
		a.registryMutex.Unlock()
	}

	for _, re := range r.regexp {
		m := re.FindStringSubmatch(line)
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
		h = strings.Trim(h, "[]")
		ip := net.ParseIP(h)
		if ip == nil {
			return nil, fmt.Errorf(`failed to parse matched IP "%s"`, h)
		}

		id := sm["id"]
		if id == "" {
			return nil, fmt.Errorf(`failed to match ID`)
		}

		a.registryMutex.Lock()
		a.registry[id] = ip
		log.Debug().Str("rule", r.name).Str("id", id).IPAddr("ip", ip).Msg("added ID to registry")
		a.registryMutex.Unlock()

		go func(id string) {
			time.Sleep(a.interval)
			a.registryMutex.Lock()
			if ip, e := a.registry[id]; e {
				delete(a.registry, id)
				log.Debug().Str("rule", r.name).Str("id", id).IPAddr("ip", ip).Msg("removed ID from registry")
			}
			a.registryMutex.Unlock()
		}(id)

		return nil, errors.New("incomplete aggregate")
	}

	return nil, fmt.Errorf(`line "%s" does not match any regexp`, line)
}

func (r *rule) match(line string) (*match, error) {
	if r.aggregate != nil {
		return r.matchAggregate(line)
	}

	return r.matchSimple(line)
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
