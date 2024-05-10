package gerberos

import (
	"net"
	"time"

	"github.com/rs/zerolog/log"
)

type occurrences struct {
	registry map[string][]time.Time
	interval time.Duration
	count    int
}

func (r *occurrences) add(ip net.IP) bool {
	ips := ip.String()
	t := time.Now()

	log.Debug().IPAddr("ip", ip).Int("length", len(r.registry[ips])).Msg("updating occurrences")

	if _, f := r.registry[ips]; !f {
		r.registry[ips] = []time.Time{t}
		return false
	}

	r.registry[ips] = append(r.registry[ips], t)
	if len(r.registry[ips]) > r.count {
		r.registry[ips] = r.registry[ips][1:]
	}

	if len(r.registry[ips]) == r.count {
		d := r.registry[ips][r.count-1].Sub(r.registry[ips][0])
		if d <= r.interval {
			delete(r.registry, ips)
			return true
		}
	}

	return false
}

func newOccurrences(interval time.Duration, count int) *occurrences {
	return &occurrences{
		registry: make(map[string][]time.Time),
		interval: interval,
		count:    count,
	}
}
