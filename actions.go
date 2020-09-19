package main

import (
	"errors"
	"fmt"
	"log"
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

	d, err := time.ParseDuration(r.Action[1])
	if err != nil {
		return fmt.Errorf("failed to parse duration parameter: %s", err)
	}
	a.duration = d

	return nil
}

func (a *banAction) perform(m *match) error {
	s := ipset4Name
	if m.ipv6 {
		s = ipset6Name
	}
	d := int64(a.duration.Seconds())
	if _, _, err := execute("ipset", "test", s, m.host); err != nil {
		if _, _, err := execute("ipset", "add", s, m.host, "timeout", fmt.Sprint(d)); err != nil {
			log.Printf(`%s: failed to add "%s" to ipset "%s" with %d second(s) timeout: %s`, a.rule.name, m.host, s, d, err)
		} else {
			log.Printf(`%s: added "%s" to ipset "%s" with %d second(s) timeout`, a.rule.name, m.host, s, d)
		}
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
	log.Printf("%s: %s (%s)", a.rule.name, m, m.regexp.String())

	return nil
}
