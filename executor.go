package main

import (
	"bytes"
	"io"
	"os/exec"
)

type executor interface {
	execute(name string, args ...string) (string, int, error)
	executeWithStd(stdin io.Reader, stdout io.Writer, name string, args ...string) (string, int, error)
}

type defaultExecutor struct{}

func (e *defaultExecutor) execute(name string, args ...string) (string, int, error) {
	return e.executeWithStd(nil, nil, name, args...)
}

func (e *defaultExecutor) executeWithStd(stdin io.Reader, stdout io.Writer, name string, args ...string) (string, int, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	var (
		b   []byte
		err error
	)
	if stdout == nil {
		b, err = cmd.CombinedOutput()
	} else {
		bf := &bytes.Buffer{}
		cmd.Stderr = bf
		b, err = bf.Bytes(), cmd.Run()
	}
	if err != nil {
		eerr, ok := err.(*exec.ExitError)
		if ok && eerr != nil {
			return string(b), eerr.ExitCode(), eerr
		}
		return "", -1, err
	}

	return string(b), 0, nil
}
