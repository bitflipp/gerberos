package main

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"time"
)

type action interface {
	initialize([]string) error
	perform(*rule, *entry) error
}

type banAction struct {
	duration time.Duration
}

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

func (a *banAction) perform(r *rule, e *entry) error {
	s := ipset4Name
	if e.ipv6 {
		s = ipset6Name
	}
	d := int64(a.duration.Seconds())
	if err := exec.Command("ipset", "test", s, e.host).Run(); err != nil {
		log.Printf("%s: adding '%s' to ipset '%s' with %d seconds timeout", r.name, e.host, s, d)
		if err := exec.Command("ipset", "add", s, e.host, "timeout", fmt.Sprint(d)).Run(); err != nil {
			log.Printf("%s: failed to add '%s' to ipset '%s' with %d seconds timeout: %s", r.name, e.host, s, d, err)
			return err
		}
	}

	return nil
}

type logAction struct{}

func (a *logAction) initialize(ps []string) error {
	return nil
}

func (a *logAction) perform(r *rule, e *entry) error {
	log.Printf("%s: %s", r.name, e)
	return nil
}
