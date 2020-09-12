package main

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"time"
)

type action interface {
	initialize(*rule) error
	perform(*match) error
}

type banAction struct {
	rule     *rule
	duration time.Duration
}

func (a *banAction) initialize(r *rule) error {
	a.rule = r

	if len(r.Action) < 2 {
		return errors.New("missing duration parameter")
	}

	if d, err := time.ParseDuration(r.Action[1]); err != nil {
		return fmt.Errorf("failed to parse duration parameter: %s", err)
	} else {
		a.duration = d
	}

	return nil
}

func (a *banAction) perform(e *match) error {
	s := ipset4Name
	if e.ipv6 {
		s = ipset6Name
	}
	d := int64(a.duration.Seconds())
	if err := exec.Command("ipset", "test", s, e.host).Run(); err != nil {
		exec.Command("ipset", "add", s, e.host, "timeout", fmt.Sprint(d)).Run()
		log.Printf(`%s: added "%s" to ipset "%s" with %d second(s) timeout`, a.rule.name, e.host, s, d)
	}

	return nil
}

type logAction struct {
	rule *rule
}

func (a *logAction) initialize(r *rule) error {
	a.rule = r

	return nil
}

func (a *logAction) perform(e *match) error {
	log.Printf("%s: %s", a.rule.name, e)

	return nil
}
