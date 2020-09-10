package main

import (
	"time"
)

type action interface {
	initialize(ps []string) error
	perform(r *rule, m entry) error
}

type banAction struct {
	duration time.Duration
}

type logAction struct{}
