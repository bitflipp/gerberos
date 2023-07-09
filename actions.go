package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

type action interface {
	initialize(r *rule) error
	perform(m *match) error
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
		return fmt.Errorf("failed to parse duration parameter: %w", err)
	}
	a.duration = d

	if len(r.Action) > 2 {
		return errors.New("superfluous parameter(s)")
	}

	return nil
}

func (a *banAction) perform(m *match) error {
	err := a.rule.runner.backend.ban(m.ip, m.ipv6, a.duration)
	if err != nil {
		log.Warn().Str("rule", a.rule.name).Str("ip", m.ip).Err(err).Msg("failed to ban IP")
	} else {
		log.Info().Str("rule", a.rule.name).Str("ip", m.ip).Dur("duration", a.duration).Msg("banned IP")
	}

	return err
}

type logAction struct {
	rule     *rule
	extended bool
}

func (a *logAction) initialize(r *rule) error {
	a.rule = r

	if len(r.Action) < 2 {
		return errors.New("missing type parameter")
	}

	switch r.Action[1] {
	case "simple":
		a.extended = false
	case "extended":
		a.extended = true
	default:
		return errors.New("invalid type parameter")
	}

	if len(r.Action) > 2 {
		return errors.New("superfluous parameter(s)")
	}

	return nil
}

func (a *logAction) perform(m *match) error {
	var ms string
	if a.extended {
		ms = m.stringExtended()
	} else {
		ms = m.stringSimple()
	}
	log.Info().Str("rule", a.rule.name).Str("match", ms).Msg("")

	return nil
}

type testAction struct {
	rule *rule
}

func (a *testAction) initialize(r *rule) error {
	a.rule = r

	return nil
}

func (a *testAction) perform(m *match) error {
	return errFault
}
