package main

import (
	"bytes"
	"testing"
	"time"
)

type mockClock struct {
	time time.Time
}

func (c *mockClock) now() time.Time {
	return c.time
}

func TestLogWriter(t *testing.T) {
	tm := time.Date(1984, 1, 15, 21, 34, 15, 0, time.UTC)
	b := bytes.Buffer{}
	lw := logWriter{
		clock:  &mockClock{time: tm},
		writer: &b,
	}
	_, err := lw.Write([]byte("message"))
	testNoError(t, err)
	if b.String() != "1984-01-15T21:34:15Z message" {
		t.Error("unexpected result")
	}
}
