package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func (s *fileSource) initialize(ps []string) error {
	if len(ps) == 0 {
		return errors.New("missing path parameter")
	}
	s.path = ps[0]

	if fi, err := os.Stat(ps[0]); err == nil && fi.IsDir() {
		return fmt.Errorf("'%s' is a directory", s.path)
	}

	return nil
}

func (s *fileSource) entries() (chan entry, error) {
	cmd := exec.Command("tail", "-n", "0", "-F", s.path)
	o, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	c := make(chan entry, 1)
	go func() {
		s := bufio.NewScanner(o)
		for s.Scan() {
			c <- newEntry(s.Text(), time.Now())
		}
	}()

	return c, cmd.Start()
}

func (s *systemdSource) initialize(ps []string) error {
	if len(ps) == 0 {
		return errors.New("missing service parameter")
	}
	s.service = ps[0]

	return nil
}

func (s *systemdSource) entries() (chan entry, error) {
	cmd := exec.Command("journalctl", "-f", "-n", "0", "-u", s.service)
	o, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	c := make(chan entry, 1)
	go func() {
		s := bufio.NewScanner(o)
		for s.Scan() {
			c <- newEntry(s.Text(), time.Now())
		}
	}()

	return c, cmd.Start()
}
