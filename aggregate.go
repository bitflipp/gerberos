package main

import (
	"net"
	"regexp"
	"sync"
	"time"
)

type aggregate struct {
	registry      map[string]net.IP
	registryMutex sync.Mutex
	interval      time.Duration
	regexp        []*regexp.Regexp
}
