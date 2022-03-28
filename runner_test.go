//go:build system

package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
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

func TestRunnerBackendInitializeInvalid(t *testing.T) {
	tbi := func(n string) {
		rn, err := newTestRunner()
		testNoError(t, err)
		rn.configuration.Backend = n
		testError(t, rn.initialize())
	}

	tbi("")
	tbi("unknown")
}

func TestRunnerBackendInitializeFaulty(t *testing.T) {
	fi := func(b string, o string, ec int, ferr error, n string, args ...string) {
		rn, err := newTestRunner()
		testNoError(t, err)
		rn.configuration.Backend = b
		rn.executor = newTestFaultyExecutor(o, ec, ferr, n, args...)
		testError(t, rn.initialize())
	}

	c, s4, s6 := "gerberos", "gerberos4", "gerberos6"
	fi("ipset", "", 1, exec.ErrNotFound, "ipset", "list")
	fi("ipset", "", 1, errFault, "ipset", "list")
	fi("ipset", "", 1, exec.ErrNotFound, "iptables", "-L")
	fi("ipset", "", 1, errFault, "iptables", "-L")
	fi("ipset", "", 1, exec.ErrNotFound, "ip6tables", "-L")
	fi("ipset", "", 1, errFault, "ipset", "create", s4, "hash:ip", "timeout", "0")
	fi("ipset", "", 1, errFault, "ipset", "create", s6, "hash:ip", "family", "inet6", "timeout", "0")
	fi("ipset", "", 1, errFault, "ip6tables", "-L")
	fi("ipset", "", 3, errFault, "iptables", "-D", c, "-j", "DROP", "-m", "set", "--match-set", s4, "src")
	fi("ipset", "", 3, errFault, "iptables", "-D", "INPUT", "-j", c)
	fi("ipset", "", 3, errFault, "iptables", "-X", c)
	fi("ipset", "", 3, errFault, "ip6tables", "-D", c, "-j", "DROP", "-m", "set", "--match-set", s6, "src")
	fi("ipset", "", 3, errFault, "ip6tables", "-D", "INPUT", "-j", c)
	fi("ipset", "", 3, errFault, "ip6tables", "-X", c)
	fi("ipset", "", 2, errFault, "ipset", "destroy", s4)
	fi("ipset", "", 2, errFault, "ipset", "destroy", s6)
	fi("ipset", "", 1, errFault, "iptables", "-N", c)
	fi("ipset", "", 1, errFault, "iptables", "-I", c, "-j", "DROP", "-m", "set", "--match-set", s4, "src")
	fi("ipset", "", 1, errFault, "iptables", "-I", "INPUT", "-j", c)
	fi("ipset", "", 1, errFault, "ip6tables", "-N", c)
	fi("ipset", "", 1, errFault, "ip6tables", "-I", c, "-j", "DROP", "-m", "set", "--match-set", s6, "src")
	fi("ipset", "", 1, errFault, "ip6tables", "-I", "INPUT", "-j", c)

	t4, s4, t6, s6 := "gerberos4", "set4", "gerberos6", "set6"
	fi("nft", "", 1, exec.ErrNotFound, "nft", "list", "ruleset")
	fi("nft", "", 1, errFault, "nft", "list", "ruleset")
	fi("nft", "", 1, errFault, "nft", "add", "table", "ip", t4)
	fi("nft", "", 1, errFault, "nft", "add", "set", "ip", t4, s4, "{ type ipv4_addr; flags timeout; }")
	fi("nft", "", 1, errFault, "nft", "add", "chain", "ip", t4, "input", "{ type filter hook input priority 0; policy accept; }")
	fi("nft", "", 1, errFault, "nft", "flush", "chain", "ip", t4, "input")
	fi("nft", "", 1, errFault, "nft", "add", "rule", "ip", t4, "input", "ip", "saddr", "@"+s4, "reject")
	fi("nft", "", 1, errFault, "nft", "add", "table", "ip6", t6)
	fi("nft", "", 1, errFault, "nft", "add", "set", "ip6", t6, s6, "{ type ipv6_addr; flags timeout; }")
	fi("nft", "", 1, errFault, "nft", "add", "chain", "ip6", t6, "input", "{ type filter hook input priority 0; policy accept; }")
	fi("nft", "", 1, errFault, "nft", "flush", "chain", "ip6", t6, "input")
	fi("nft", "", 1, errFault, "nft", "add", "rule", "ip6", t6, "input", "ip6", "saddr", "@"+s6, "reject")
}

func TestRunnerBackendFinalizeFaulty(t *testing.T) {
	ff := func(b string, o string, ec int, ferr error, n string, args ...string) {
		f, err := os.CreateTemp("", "gerberos-")
		testNoError(t, err)
		tn := f.Name()
		defer testNoError(t, os.Remove(tn))
		rn, err := newTestRunner()
		testNoError(t, err)
		rn.configuration.Backend = b
		rn.configuration.SaveFilePath = tn
		rn.executor = newTestFaultyExecutor(o, ec, ferr, n, args...)
		testNoError(t, rn.initialize())
		testError(t, rn.finalize())
	}

	t4, s4, t6, s6 := "gerberos4", "set4", "gerberos6", "set6"
	ff("ipset", "", 1, errFault, "ipset", "save")
	ff("nft", "", 1, errFault, "nft", "delete", "table", "ip", t4)
	ff("nft", "", 1, errFault, "nft", "delete", "table", "ip6", t6)
	ff("nft", "", 1, errFault, "nft", "list", "set", "ip", t4, s4)
	ff("nft", "", 1, errFault, "nft", "list", "set", "ip6", t6, s6)
}

func TestRunnerExecute(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	testNoError(t, rn.initialize())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		rn.run(false)
		testNoError(t, rn.finalize())
		wg.Done()
	}()
	time.Sleep(5 * time.Second)
	rn.stop()
	wg.Wait()
}

func TestRunnerPerformAction(t *testing.T) {
	pa := func(b string, a []string) {
		rn, err := newTestRunner()
		testNoError(t, err)
		rn.configuration.Backend = b
		rn.configuration.Rules["test"].Action = a
		testNoError(t, rn.initialize())
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			rn.run(false)
			testNoError(t, rn.finalize())
			wg.Done()
		}()
		time.Sleep(5 * time.Second)
		rn.stop()
		wg.Wait()
	}

	pa("ipset", []string{"log", "simple"})
	pa("ipset", []string{"log", "extended"})
	pa("ipset", []string{"ban", "1h"})
	pa("nft", []string{"log", "simple"})
	pa("nft", []string{"log", "extended"})
	pa("nft", []string{"ban", "1h"})
	pa("test", []string{"log", "simple"})
	pa("nft", []string{"log", "extended"})
	pa("test", []string{"ban", "1h"})
}

func TestRunnerPersistence(t *testing.T) {
	p := func(b string) {
		f, err := os.CreateTemp("", "gerberos-")
		testNoError(t, err)
		tn := f.Name()
		defer testNoError(t, os.Remove(tn))
		{
			rn, err := newTestRunner()
			testNoError(t, err)
			rn.configuration.Backend = b
			rn.configuration.SaveFilePath = tn
			testNoError(t, rn.initialize())
			rn.backend.Ban("123.123.123.123", false, time.Hour)
			rn.backend.Ban("affe::affe", true, time.Hour)
			testNoError(t, rn.finalize())
		}
		{
			rn, err := newTestRunner()
			testNoError(t, err)
			rn.configuration.Backend = b
			rn.configuration.SaveFilePath = tn
			testNoError(t, rn.initialize())
			testNoError(t, rn.finalize())
		}
	}

	p("ipset")
	p("nft")
}

func TestRunnerMissingConfiguration(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	rn.configuration = nil
	testError(t, rn.initialize())
}

func TestRunnerMissingBackend(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	rn.configuration.Backend = ""
	testError(t, rn.initialize())
}

func TestRunnerBanFaulty(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	r := rn.configuration.Rules["test"]
	r.Action = []string{"ban", "1h"}
	testNoError(t, rn.initialize())
	rn.backend.(*testBackend).banErr = errFault
	testError(t, r.action.perform(&match{}))
}

func TestRunnerIpsetBackendRestoreFaulty(t *testing.T) {
	f, err := os.CreateTemp("", "gerberos-")
	testNoError(t, err)
	tn := f.Name()
	defer testNoError(t, os.Remove(tn))
	{
		rn, err := newTestRunner()
		testNoError(t, err)
		rn.configuration.Backend = "ipset"
		rn.configuration.SaveFilePath = tn
		testNoError(t, rn.initialize())
		testNoError(t, rn.finalize())
	}
	{
		rn, err := newTestRunner()
		testNoError(t, err)
		rn.configuration.Backend = "ipset"
		rn.configuration.SaveFilePath = tn
		rn.executor = newTestFaultyExecutor("", 1, errFault, "ipset", "restore")
		testNoError(t, rn.initialize())
		rn.executor = newTestFaultyExecutor("", 1, errFault, "ipset", "create", "gerberos4", "hash:ip", "timeout", "0")
		testError(t, rn.initialize())
	}
}

func TestRunnerIpsetBackendBanFaulty(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	rn.configuration.Backend = "ipset"
	testNoError(t, rn.initialize())
	rn.executor = newTestFaultyExecutor("", 1, errFault, "ipset", "add", "gerberos4", "123.123.123.123", "timeout", "3600")
	testError(t, rn.backend.Ban("123.123.123.123", false, time.Hour))
}

func TestRunnerIpsetBackendFinalizeFaulty(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	rn.configuration.Backend = "ipset"
	testNoError(t, rn.initialize())
	rn.executor = newTestFaultyExecutor("", 3, errFault, "ipset", "destroy", "gerberos4")
	testError(t, rn.finalize())
}

func TestRunnerNftBackendBanFaulty(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	rn.configuration.Backend = "nft"
	testNoError(t, rn.initialize())
	rn.executor = newTestFaultyExecutor("", 1, errFault, "nft", "add", "element", "ip6", "gerberos6", "set6", "{ affe::affe timeout 3600s }")
	testNoError(t, rn.backend.Ban("affe::affe", true, time.Hour))
	rn.executor = newTestFaultyExecutor("", 2, errFault, "nft", "add", "element", "ip6", "gerberos6", "set6", "{ affe::affe timeout 3600s }")
	testError(t, rn.backend.Ban("affe::affe", true, time.Hour))
	rn.executor = newTestFaultyExecutor("", 1, errFault, "nft", "add", "element", "ip", "gerberos4", "set4", "{ 123.123.123.123 timeout 3600s }")
	testNoError(t, rn.backend.Ban("123.123.123.123", false, time.Hour))
	rn.executor = newTestFaultyExecutor("", 2, errFault, "nft", "add", "element", "ip", "gerberos4", "set4", "{ 123.123.123.123 timeout 3600s }")
	testError(t, rn.backend.Ban("123.123.123.123", false, time.Hour))
}

func TestRunnerRulesWorker(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	testNoError(t, rn.initialize())
	r := rn.configuration.Rules["test"]
	r.worker(false)
}

func TestRunnerRulesWorkerRequeue(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	testNoError(t, rn.initialize())
	r := rn.configuration.Rules["test"]
	rn.respawnWorkerDelay = 0
	s := r.source.(*testSource)
	s.processPath = "test/quitter"
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		rn.run(true)
		wg.Done()
	}()
	r.worker(true)
	time.Sleep(5 * time.Second)
	rn.stop()
	wg.Wait()
}

func TestRunnerRulesWorkerMatchesFaulty(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	testNoError(t, rn.initialize())
	r := rn.configuration.Rules["test"]
	r.source.(*testSource).matchesErr = errFault
	r.worker(false)
}

func TestRunnerRulesInvalid(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	rn.configuration.Rules["test"].Source = []string{"unknown"}
	testError(t, rn.initialize())
}

func TestRunnerRulesWorkerInvalidProcess(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	testNoError(t, rn.initialize())
	r := rn.configuration.Rules["test"]
	s := r.source.(*testSource)
	s.processPath = "test/unknown"
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		rn.run(false)
		wg.Done()
	}()
	time.Sleep(5 * time.Second)
	rn.stop()
	wg.Wait()
}

func TestRunnerSources(t *testing.T) {
	ts := func(s []string) {
		rn, err := newTestRunner()
		testNoError(t, err)
		r := rn.configuration.Rules["test"]
		r.Source = s
		testNoError(t, rn.initialize())
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			rn.run(false)
			wg.Done()
		}()
		time.Sleep(5 * time.Second)
		rn.stop()
		wg.Wait()
	}

	ts([]string{"file", "test/empty"})
	ts([]string{"systemd", "service"})
	ts([]string{"kernel"})
	ts([]string{"process", "test/quitter"})
}

func TestRunnerWorkerActionFaulty(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	r := rn.configuration.Rules["test"]
	r.Action = []string{"test"}
	testNoError(t, rn.initialize())
	testNoError(t, r.worker(false))
}

func TestRunnerDanglingProcessFlaky(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	r := newTestValidRule()
	r.Source = []string{"process", "test/trapper_forever"}
	rn.configuration.Rules["test"] = r
	testNoError(t, rn.initialize())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	before, err := testCountChildren()
	testNoError(t, err)
	go func() {
		rn.run(true)
		wg.Done()
	}()
	time.Sleep(1 * time.Second)
	rn.stop()
	time.Sleep(6 * time.Second)
	wg.Wait()
	after, err := testCountChildren()
	testNoError(t, err)
	if after != before {
		t.Errorf("Children not cleaned up. Before: %d; After: %d", before, after)
	}
}

func TestRunnerManyRulesFlaky(t *testing.T) {
	rn, err := newTestRunner()
	testNoError(t, err)
	delete(rn.configuration.Rules, "test")
	cn := 100
	for i := 0; i < cn; i++ {
		r := newTestValidRule()
		r.Source = []string{"process", "test/trapper_random"}
		rn.configuration.Rules[fmt.Sprintf("test-%d", i)] = r
	}
	rn.respawnWorkerDelay = 2 * time.Millisecond
	testNoError(t, rn.initialize())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		rn.run(true)
		wg.Done()
	}()
	time.Sleep(4 * time.Second)
	rn.stop()
	wg.Wait()
}
