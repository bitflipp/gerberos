package main

import (
	"io"
	"time"
)

type clock interface {
	now() time.Time
}

type realTimeClock struct{}

func (c *realTimeClock) now() time.Time {
	return time.Now()
}

type rfcPrefixWriter struct {
	clock  clock
	writer io.Writer
}

func (l rfcPrefixWriter) Write(p []byte) (n int, err error) {
	ts := []byte(l.clock.now().Format(time.RFC3339) + " ")
	return l.writer.Write(append(ts, p...))
}
