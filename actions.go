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

func (a *banAction) perform(m *match) error {
	s := ipset4Name
	if m.ipv6 {
		s = ipset6Name
	}
	d := int64(a.duration.Seconds())
	if err := exec.Command("ipset", "test", s, m.host).Run(); err != nil {
		exec.Command("ipset", "add", s, m.host, "timeout", fmt.Sprint(d)).Run()
		log.Printf(`%s: added "%s" to ipset "%s" with %d second(s) timeout`, a.rule.name, m.host, s, d)
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

func (a *logAction) perform(m *match) error {
	log.Printf("%s: %s", a.rule.name, m)

	return nil
}
