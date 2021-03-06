package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	ipMagicText = "%ip%"
	idMagicText = "%id%"
)

var (
	ipMagicRegexp = regexp.MustCompile(ipMagicText)
	ipRegexpText  = `(?P<ip>(\d?\d?\d\.){3}\d?\d?\d|\[?([0-9A-Fa-f]{0,4}::?){1,6}[0-9A-Fa-f]{0,4}::?[0-9A-Fa-f]{0,4})\]?`
	idMagicRegexp = regexp.MustCompile(idMagicText)
	idRegexpText  = `(?P<id>(.*))`
)

type rule struct {
	Source      []string
	Regexp      []string
	Action      []string
	Aggregate   []string
	Occurrences []string

	name        string
	source      source
	regexp      []*regexp.Regexp
	action      action
	aggregate   *aggregate
	occurrences *occurrences
}

func (r *rule) initializeSource() error {
	if r.Source == nil {
		return errors.New("missing source")
	}

	if len(r.Source) == 0 {
		return errors.New("empty source")
	}

	switch r.Source[0] {
	case "file":
		r.source = &fileSource{}
	case "systemd":
		r.source = &systemdSource{}
	case "kernel":
		r.source = &kernelSource{}
	default:
		return errors.New("unknown source")
	}

	return r.source.initialize(r)
}

func (r *rule) initializeRegexp() error {
	if r.Regexp == nil {
		return errors.New("missing regexp")
	}

	if len(r.Regexp) == 0 {
		return errors.New("empty regexp")
	}

	r.regexp = make([]*regexp.Regexp, 0)
	for _, s := range r.Regexp {
		if strings.Contains(s, "(?P<ip>") {
			return errors.New(`regexp must not contain a subexpression named "ip" ("(?P<ip>")`)
		}

		if strings.Contains(s, "(?P<id>") {
			return errors.New(`regexp must not contain a subexpression named "id" ("(?P<id>")`)
		}

		if len(ipMagicRegexp.FindAllStringIndex(s, -1)) != 1 {
			return fmt.Errorf(`"%s" must appear exactly once in regexp`, ipMagicText)
		}

		if r.Aggregate != nil && len(idMagicRegexp.FindAllStringIndex(s, -1)) != 1 {
			return fmt.Errorf(`"%s" must appear exactly once in regexp if the aggregate option is used`, idMagicText)
		}

		t := strings.Replace(s, ipMagicText, ipRegexpText, 1)
		t = strings.Replace(t, idMagicText, idRegexpText, 1)
		re, err := regexp.Compile(t)
		if err != nil {
			return err
		}
		r.regexp = append(r.regexp, re)
	}

	return nil
}

func (r *rule) initializeAction() error {
	if r.Action == nil {
		return errors.New("missing action")
	}

	if len(r.Action) == 0 {
		return errors.New("empty action")
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

func (r *rule) initializeAggregate() error {
	if r.Aggregate == nil {
		return nil
	}

	if len(r.Aggregate) < 1 {
		return errors.New("missing interval parameter")
	}
	i, err := time.ParseDuration(r.Aggregate[0])
	if err != nil {
		return fmt.Errorf("failed to parse interval parameter: %s", err)
	}

	if len(r.Aggregate) < 2 {
		return errors.New("missing regexp")
	}

	res := make([]*regexp.Regexp, 0)
	for _, s := range r.Aggregate[1:] {
		if strings.Contains(s, "(?P<id>") {
			return errors.New(`regexp must not contain a subexpression named "id" ("(?P<id>")`)
		}

		if len(idMagicRegexp.FindAllStringIndex(s, -1)) != 1 {
			return fmt.Errorf(`"%s" must appear exactly once in regexp`, idMagicRegexp)
		}

		re, err := regexp.Compile(strings.Replace(s, idMagicText, idRegexpText, 1))
		if err != nil {
			return err
		}
		res = append(res, re)
	}

	r.aggregate = &aggregate{
		registry: make(map[string]net.IP, 0),
		interval: i,
		regexp:   res,
	}

	return nil
}

func (r *rule) initializeOccurrences() error {
	if r.Occurrences == nil {
		return nil
	}

	if len(r.Occurrences) < 1 {
		return errors.New("missing count parameter")
	}
	c, err := strconv.Atoi(r.Occurrences[0])
	if err != nil {
		return fmt.Errorf("failed to parse count parameter: %s", err)
	}
	if c < 2 {
		return errors.New("invalid count parameter: must be > 1")
	}

	if len(r.Occurrences) < 2 {
		return errors.New("missing interval parameter")
	}
	i, err := time.ParseDuration(r.Occurrences[1])
	if err != nil {
		return fmt.Errorf("failed to parse interval parameter: %s", err)
	}

	r.occurrences = &occurrences{
		registry: make(map[string][]time.Time, 0),
		interval: i,
		count:    c,
	}

	return nil
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

	if err := r.initializeAggregate(); err != nil {
		return err
	}

	if err := r.initializeOccurrences(); err != nil {
		return err
	}

	return nil
}

func (r *rule) processScanner(n string, args ...string) (chan *match, error) {
	cmd := exec.Command(n, args...)
	o, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	e, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	log.Printf(`%s: scanning process stdout and stderr: "%s"`, r.name, cmd)

	c := make(chan *match, 1)
	go func() {
		sc := bufio.NewScanner(o)
		for sc.Scan() {
			if m, err := r.match(sc.Text()); err == nil {
				c <- m
			} else {
				if configuration.Verbose {
					log.Printf("%s: failed to create match: %s", r.name, err)
				}
			}
		}
		close(c)
		if err = sc.Err(); err != nil {
			log.Printf(`%s: error while scanning command "%s": %s`, r.name, cmd, err.Error())
		}
		if err = cmd.Wait(); err != nil {
			log.Printf(`%s: error while executing command "%s": %s`, r.name, cmd, err.Error())
		}
	}()
	go func() {
		sc := bufio.NewScanner(e)
		for sc.Scan() {
			log.Printf(`%s: process stderr: "%s"`, r.name, sc.Text())
		}
	}()

	return c, nil
}

func (r *rule) worker() {
	c, err := r.source.matches()
	if err != nil {
		log.Printf("%s: failed to initialize matches channel: %s", r.name, err)
		return
	}

	for m := range c {
		p := true
		if r.occurrences != nil {
			p = r.occurrences.add(m.ip)
		}

		if p {
			if err := r.action.perform(m); err != nil {
				log.Printf("%s: failed to perform action: %s", r.name, err)
			}
		}
	}

	time.Sleep(5 * time.Second)
	log.Printf("%s: queuing worker for respawn", r.name)
	respawnWorkerChan <- r
}
