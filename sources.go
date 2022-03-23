package main

import (
	"errors"
	"fmt"
	"os"
)

type source interface {
	initialize(r *rule) error
	matches() (chan *match, error)
}

type fileSource struct {
	rule *rule
	path string
}

func (s *fileSource) initialize(r *rule) error {
	s.rule = r

	if len(r.Source) < 2 {
		return errors.New("missing path parameter")
	}
	s.path = r.Source[1]

	if fi, err := os.Stat(s.path); err == nil && fi.IsDir() {
		return fmt.Errorf(`"%s" is a directory`, s.path)
	}

	if len(r.Source) > 2 {
		return errors.New("superfluous parameter(s)")
	}

	return nil
}

func (s *fileSource) matches() (chan *match, error) {
	return s.rule.processScanner("tail", "-n", "0", "-F", s.path)
}

type systemdSource struct {
	rule    *rule
	service string
}

func (s *systemdSource) initialize(r *rule) error {
	s.rule = r

	if len(r.Source) < 2 {
		return errors.New("missing service parameter")
	}
	s.service = r.Source[1]

	if len(r.Source) > 2 {
		return errors.New("superfluous parameter(s)")
	}

	return nil
}

func (s *systemdSource) matches() (chan *match, error) {
	return s.rule.processScanner("journalctl", "-n", "0", "-f", "-u", s.service)
}

type kernelSource struct {
	rule *rule
}

func (k *kernelSource) initialize(r *rule) error {
	k.rule = r

	if len(r.Source) > 1 {
		return errors.New("superfluous parameter(s)")
	}

	return nil
}

func (k *kernelSource) matches() (chan *match, error) {
	return k.rule.processScanner("journalctl", "-kf", "-n", "0")
}

type testSource struct {
	rule        *rule
	matchesErr  error
	processPath string
}

func (s *testSource) initialize(r *rule) error {
	s.rule = r

	return nil
}

func (s *testSource) matches() (chan *match, error) {
	if s.matchesErr != nil {
		return nil, s.matchesErr
	}

	p := "test/producer"
	if s.processPath != "" {
		p = s.processPath
	}
	return s.rule.processScanner(p)
}

type processSource struct {
	rule *rule
	name string
	args []string
}

func (s *processSource) initialize(r *rule) error {
	s.rule = r

	if len(r.Source) < 2 {
		return errors.New("missing process name")
	}
	s.name = r.Source[1]

	s.args = r.Source[2:]

	return nil
}

func (s *processSource) matches() (chan *match, error) {
	return s.rule.processScanner(s.name, s.args...)
}
