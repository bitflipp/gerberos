package main

import (
	"errors"
	"log"
	"regexp"
)

type rule struct {
	Source []string
	Regexp string
	Action []string

	name   string
	source source
	regexp *regexp.Regexp
	action action
}

func (r *rule) initializeSource() error {
	if len(r.Source) == 0 {
		return errors.New("missing source")
	}

	switch r.Source[0] {
	case "file":
		r.source = &fileSource{}
	case "systemd":
		r.source = &systemdSource{}
	default:
		return errors.New("unknown source")
	}

	return r.source.initialize(r.Source[1:])
}

func (r *rule) initializeAction() error {
	if len(r.Action) == 0 {
		return errors.New("missing action")
	}

	switch r.Action[0] {
	case "ban":
		r.action = &banAction{}
	case "log":
		r.action = &logAction{}
	default:
		return errors.New("unknown action")
	}

	return r.action.initialize(r.Action[1:])
}

func (r *rule) initialize() error {
	if err := r.initializeSource(); err != nil {
		return err
	}

	if e, err := regexp.Compile(r.Regexp); err != nil {
		return err
	} else {
		r.regexp = e
	}

	if err := r.initializeAction(); err != nil {
		return err
	}

	return nil
}

func (r *rule) worker() {
	c, err := r.source.entries()
	if err != nil {
		log.Printf("%s: failed to initialize entries channel: %s", r.name, err)
		return
	}
	for m := range c {
		if m.err == nil {
			if !r.regexp.MatchString(m.line) {
				continue
			}

			if err := r.action.perform(r, m); err != nil {
				log.Printf("%s: failed to perform action: %s", r.name, err)
			}
		}
	}
}
