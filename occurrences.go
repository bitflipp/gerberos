package main

import "time"

type occurrences struct {
	registry map[string][]time.Time
	interval time.Duration
	count    int
}

func (r *occurrences) add(h string) bool {
	if _, f := r.registry[h]; !f {
		r.registry[h] = []time.Time{time.Now()}
		return false
	}

	r.registry[h] = append(r.registry[h], time.Now())
	if len(r.registry[h]) > r.count {
		r.registry[h] = r.registry[h][1:]
	}

	if len(r.registry[h]) == r.count {
		d := r.registry[h][r.count-1].Sub(r.registry[h][0])
		if d <= r.interval {
			delete(r.registry, h)
			return true
		}
	}

	return false
}

func newOccurrences(i time.Duration, c int) *occurrences {
	return &occurrences{
		registry: make(map[string][]time.Time),
		interval: i,
		count:    c,
	}
}
