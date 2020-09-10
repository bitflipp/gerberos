package main

import (
	"errors"
	"log"
	"time"
)

func (a *banAction) initialize(ps []string) error {
	if len(ps) == 0 {
		return errors.New("missing duration parameter")
	}

	if d, err := time.ParseDuration(ps[0]); err != nil {
		return errors.New("invalid duration parameter")
	} else {
		a.duration = d
	}

	return nil
}

func (a *banAction) perform(r *rule, m match) error {
	log.Printf("%s: (not) banning: %+v", r.name, m)
	return nil
}

func (a *logAction) initialize(ps []string) error {
	return nil
}

func (a *logAction) perform(r *rule, m match) error {
	log.Printf("%s: %+v", r.name, m)
	return nil
}
