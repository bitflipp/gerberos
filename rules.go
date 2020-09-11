package main

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
)

const (
	ipv4MagicText = "%ip4%"
	ipv6MagicText = "%ip6%"
)

var (
	ipv4MagicRegexp = regexp.MustCompile(ipv4MagicText)
	ipv6MagicRegexp = regexp.MustCompile(ipv6MagicText)
	ipv4RegexpText  = `(?P<host>((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`
	ipv6RegexpText  = `(?P<host>(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])))`
)

type rule struct {
	Source []string
	Regexp string
	Action []string

	name   string
	source source
	regexp *regexp.Regexp
	ipv6   bool
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
		return errors.New("regexp must not contain a subexpression named 'host' ('(?P<host>')")
	}

	var re string
	o4 := len(ipv4MagicRegexp.FindAllStringIndex(r.Regexp, -1))
	o6 := len(ipv6MagicRegexp.FindAllStringIndex(r.Regexp, -1))
	switch {
	case o4 == 1 && o6 == 0:
		re = strings.Replace(r.Regexp, ipv4MagicText, ipv4RegexpText, 1)
	case o4 == 0 && o6 == 1:
		re = strings.Replace(r.Regexp, ipv6MagicText, ipv6RegexpText, 1)
		r.ipv6 = true
	default:
		return fmt.Errorf("regexp must contain exactly one of '%s' or '%s'", ipv4MagicText, ipv6MagicText)
	}

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
