//go:build system

package main

import (
	"os"
	"testing"
	"time"
)

func TestRunnerInitializeFinalize(t *testing.T) {
	tb := func(n string) {
		rn, err := newTestRunner()
		testNoError(t, err)
		rn.configuration.Backend = n
		testNoError(t, rn.initialize())
		testNoError(t, rn.finalize())
	}

	tb("ipset")
	tb("nft")
	tb("test")
}

func TestBackendInitializeInvalid(t *testing.T) {
	tbi := func(n string) {
		rn, err := newTestRunner()
		testNoError(t, err)
		rn.configuration.Backend = n
		if err := rn.initialize(); err == nil {
			t.Error("expected error")
		}
	}

	tbi("")
	tbi("unknownBackend")
}

func TestRunnerExecuteFlaky(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	testNoError(t, rn.initialize())
	go rn.execute(false)
	time.Sleep(100 * time.Millisecond)
	rn.cancelChan <- true
	time.Sleep(100 * time.Millisecond)
	testNoError(t, rn.finalize())
}

func TestRunnerPerformActionFlaky(t *testing.T) {
	pa := func(b string, a []string) {
		rn, err := newTestRunner()
		testNoError(t, err)
		rn.configuration.Backend = b
		rn.configuration.Rules["test"].Action = a
		testNoError(t, rn.initialize())
		go rn.execute(false)
		time.Sleep(200 * time.Millisecond)
		rn.cancelChan <- true
		time.Sleep(100 * time.Millisecond)
		testNoError(t, rn.finalize())
	}

	pa("ipset", []string{"log", "simple"})
	pa("ipset", []string{"ban", "1h"})
	pa("nft", []string{"log", "simple"})
	pa("nft", []string{"ban", "1h"})
	pa("test", []string{"log", "simple"})
	pa("test", []string{"ban", "1h"})
}

func TestRunnerPersistence(t *testing.T) {
	p := func(b string) {
		f, err := os.CreateTemp("", "gerberos-")
		testNoError(t, err)
		n := f.Name()
		defer testNoError(t, os.Remove(n))
		{
			rn, err := newTestRunner()
			testNoError(t, err)
			rn.configuration.Backend = b
			rn.configuration.SaveFilePath = n
			testNoError(t, rn.initialize())
			rn.backend.Ban("123.123.123.123", false, time.Hour)
			rn.backend.Ban("affe::affe", true, time.Hour)
			testNoError(t, rn.backend.Finalize())
		}
		{
			rn, err := newTestRunner()
			testNoError(t, err)
			rn.configuration.Backend = b
			rn.configuration.SaveFilePath = n
			testNoError(t, rn.initialize())
			time.Sleep(100 * time.Millisecond)
			testNoError(t, rn.finalize())
		}
	}

	p("ipset")
	p("nft")
}
