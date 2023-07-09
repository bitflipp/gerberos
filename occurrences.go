package main

import (
	"net"
	"time"
)

type occurrences struct {
	registry map[string][]time.Time
	interval time.Duration
	count    int
}

func (r *occurrences) add(ip net.IP) bool {
	s := ip.String()

	if _, f := r.registry[s]; !f {
		r.registry[s] = []time.Time{time.Now()}
		return false
	}

	r.registry[s] = append(r.registry[s], time.Now())
	if len(r.registry[s]) > r.count {
		r.registry[s] = r.registry[s][1:]
	}

	if len(r.registry[s]) == r.count {
		d := r.registry[s][r.count-1].Sub(r.registry[s][0])
		if d <= r.interval {
			delete(r.registry, s)
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
