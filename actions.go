package main

import (
	"time"
)

type action interface {
	initialize([]string) error
	perform(*rule, *entry) error
}

type banAction struct {
	duration time.Duration
}

type logAction struct{}
