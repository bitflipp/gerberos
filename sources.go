package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type source interface {
	initialize(*rule) error
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

	return nil
}

func (s *fileSource) matches() (chan *match, error) {
	cmd := exec.Command("tail", "-n", "0", "-F", s.path)
	o, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	c := make(chan *match, 1)
	go func() {
		sc := bufio.NewScanner(o)
		for sc.Scan() {
			if m, err := newMatch(s.rule, sc.Text()); err == nil {
				c <- m
			}
		}
	}()

	return c, cmd.Start()
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

	return nil
}

func (s *systemdSource) matches() (chan *match, error) {
	cmd := exec.Command("journalctl", "-n", "0", "-f", "-u", s.service)
	o, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	c := make(chan *match, 1)
	go func() {
		sc := bufio.NewScanner(o)
		for sc.Scan() {
			if m, err := newMatch(s.rule, sc.Text()); err == nil {
				c <- m
			}
		}
	}()

	return c, cmd.Start()
}
