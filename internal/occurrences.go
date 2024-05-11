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

func (o *occurrences) add(ip net.IP) bool {
	ips := ip.String()
	t := time.Now()

	log.Debug().IPAddr("ip", ip).Int("length", len(o.registry[ips])).Msg("updating occurrences")

	if _, f := o.registry[ips]; !f {
		o.registry[ips] = []time.Time{t}
		return false
	}

	o.registry[ips] = append(o.registry[ips], t)
	if len(o.registry[ips]) > o.count {
		o.registry[ips] = o.registry[ips][1:]
	}

	if len(o.registry[ips]) == o.count {
		d := o.registry[ips][o.count-1].Sub(o.registry[ips][0])
		if d <= o.interval {
			delete(o.registry, ips)
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
