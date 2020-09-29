package main

import (
	"regexp"
	"sync"
	"time"
)

type aggregate struct {
	registry map[string]struct {
		ip    string
		count int
	}
	registryMutex sync.Mutex
	interval      time.Duration
	regexp        []*regexp.Regexp
}
