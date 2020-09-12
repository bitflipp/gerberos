package main

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
)

const (
	ipMagicText = "%host%"
)

var (
	ipMagicRegexp = regexp.MustCompile(ipMagicText)
	ipRegexpText  = `(?P<host>(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}|(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|((?:[0-9A-Fa-f]{1,4}(?::[0-9A-Fa-f]{1,4})*)?)::((?:[0-9A-Fa-f]{1,4}(?::[0-9A-Fa-f]{1,4})*)?))`
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

	return r.source.initialize(r)
}

func (r *rule) initializeRegexp() error {
	if strings.Contains(r.Regexp, "(?P<host>") {
		return errors.New(`regexp must not contain a subexpression named "host" ("(?P<host>")`)
	}

	if len(ipMagicRegexp.FindAllStringIndex(r.Regexp, -1)) != 1 {
		return fmt.Errorf(`"%s" must appear exactly once in regexp`, ipMagicText)
	}
	re := strings.Replace(r.Regexp, ipMagicText, ipRegexpText, 1)

	if e, err := regexp.Compile(re); err != nil {
		return err
	} else {
		r.regexp = e
	}

	return nil
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

	return r.action.initialize(r)
}

func (r *rule) initialize() error {
	if err := r.initializeSource(); err != nil {
		return err
	}

	if err := r.initializeRegexp(); err != nil {
		return err
	}

	if err := r.initializeAction(); err != nil {
		return err
	}

	return nil
}

func (r *rule) worker() {
	c, err := r.source.matches()
	if err != nil {
		log.Printf("%s: failed to initialize matches channel: %s", r.name, err)
		return
	}

	for m := range c {
		if err := r.action.perform(m); err != nil {
			log.Printf("%s: failed to perform action: %s", r.name, err)
		}
	}
}
