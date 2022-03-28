package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type runner struct {
	configuration      *configuration
	backend            backend
	respawnWorkerDelay time.Duration
	respawnWorkerChan  chan *rule
	executor           executor
	stop               context.CancelFunc
	stopped            context.Context
}

func (rn *runner) initialize() error {
	if rn.configuration == nil {
		return errors.New("configuration has not been set")
	}

	// Backend
	switch rn.configuration.Backend {
	case "":
		return errors.New("missing configuration value for backend")
	case "ipset":
		rn.backend = &ipsetBackend{runner: rn}
	case "nft":
		rn.backend = &nftBackend{runner: rn}
	case "test":
		rn.backend = &testBackend{runner: rn}
	default:
		return fmt.Errorf("unknown backend: %s", rn.configuration.Backend)
	}
	if err := rn.backend.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize backend: %w", err)
	}

	// Rules
	for n, r := range rn.configuration.Rules {
		r.name = n
		if err := r.initialize(rn); err != nil {
			return fmt.Errorf(`failed to initialize rule "%s": %s`, n, err)
		}
	}

	return nil
}

func (rn *runner) finalize() error {
	if err := rn.backend.Finalize(); err != nil {
		return fmt.Errorf("failed to finalize backend: %w", err)
	}

	return nil
}

func (rn *runner) spawnWorker(r *rule, requeue bool) {
	go func() {
		select {
		case <-rn.stopped.Done():
			return
		default:
			r.worker(requeue)
		}
	}()
	log.Printf("%s: spawned worker", r.name)
}

func (rn *runner) run(requeueWorkers bool) {
	for _, r := range rn.configuration.Rules {
		rn.spawnWorker(r, requeueWorkers)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signalChan)

	go func() {
		for {
			select {
			case r := <-rn.respawnWorkerChan:
				time.Sleep(rn.respawnWorkerDelay)
				rn.spawnWorker(r, requeueWorkers)
			case <-rn.stopped.Done():
				return
			}
		}
	}()

	select {
	case <-rn.stopped.Done():
		return
	case <-signalChan:
		rn.stop()
	}
}

func newRunner(c *configuration) *runner {
	ctx, cancel := context.WithCancel(context.Background())
	return &runner{
		configuration:      c,
		respawnWorkerDelay: 5 * time.Second,
		respawnWorkerChan:  make(chan *rule),
		executor:           &defaultExecutor{},
		stop:               cancel,
		stopped:            ctx,
	}
}
