package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func execute(in io.Reader, n string, args ...string) (string, int, error) {
	cmd := exec.Command(n, args...)
	cmd.Stdin = in
	log.Printf("executing: %s", cmd)
	b, err := cmd.CombinedOutput()
	if err != nil {
		eerr, ok := err.(*exec.ExitError)
		if ok && eerr != nil {
			return string(b), eerr.ExitCode(), eerr
		}
		return "", -1, err
	}

	return string(b), 0, nil
}

// Testing requires the instance name n to be dynamic.
// It defaults to os.Args[0].
func isInstanceAlreadyRunning(n string) (bool, error) {
	s, _, err := execute(nil, "ps", "axco", "command")
	if err != nil {
		return false, err
	}

	if n == "" {
		n = os.Args[0]
	}
	n = path.Base(n)
	oc := false
	for _, p := range strings.Split(s, "\n") {
		if p == n {
			if oc {
				return true, nil
			}
			oc = true
		}
	}

	return false, nil
}
