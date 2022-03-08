package main

import (
	"io"
	"os/exec"
	"strings"
	"testing"
)

func TestExecute(t *testing.T) {
	ts := "test"
	r := strings.NewReader(ts)
	o, c, err := execute(r, "cat")
	testNoError(t, err)
	if o != ts {
		t.Errorf(`expected output "%s", got "%s"`, ts, o)
	}
	if c != 0 {
		t.Errorf(`expected exit code 0, got %d`, c)
	}
}

func TestExecuteExitError(t *testing.T) {
	_, c, err := execute(nil, "cat", "--invalid-flag")
	if err == nil {
		t.Error("expected error")
	}
	if c != 1 {
		t.Errorf(`expected exit code 1, got %d`, c)
	}
}

func TestExecuteUnknownCommandFlaky(t *testing.T) {
	_, c, err := execute(nil, "unknown_command_baighah6othoo0ikei9Ahngay2geifah")
	if err == nil {
		t.Error("expected error")
	}
	if c != -1 {
		t.Errorf(`expected exit code -1, got %d`, c)
	}
}

func TestIsInstanceAlreadyRunning(t *testing.T) {
	r := func() (io.ReadCloser, io.WriteCloser, error) {
		cmd := exec.Command("test/reader")
		rp, err := cmd.StdoutPipe()
		if err != nil {
			return nil, nil, err
		}
		wp, err := cmd.StdinPipe()
		if err != nil {
			return nil, nil, err
		}
		return rp, wp, cmd.Start()
	}

	rp1, wp1, err := r()
	testNoError(t, err)
	rp1.Read(nil)
	{
		iar, err := isInstanceAlreadyRunning("reader")
		testNoError(t, err)
		if iar {
			t.Error("expected no already running instance")
		}
	}

	rp2, wp2, err := r()
	testNoError(t, err)
	rp2.Read(nil)
	testNoError(t, err)
	{
		iar, err := isInstanceAlreadyRunning("reader")
		testNoError(t, err)
		if !iar {
			t.Error("expected an already running instance")
		}
	}

	wp1.Write([]byte("\n"))
	wp2.Write([]byte("\n"))
}

func TestIsInstanceAlreadyRunningEmptyNameFlaky(t *testing.T) {
	iar, err := isInstanceAlreadyRunning("")
	testNoError(t, err)
	if iar {
		t.Error("expected no already running instance")
	}
}
