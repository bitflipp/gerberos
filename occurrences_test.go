package main

import (
	"testing"
	"time"
)

func TestOccurrencesFlaky(t *testing.T) {
	h := "123.123.123.123"

	o := newTestOccurrences()
	for i := 0; i < 9; i++ {
		if o.add(h) {
			t.Error("unexpected result")
		}
	}
	if !o.add(h) {
		t.Error("unexpected result")
	}

	o = newTestOccurrences()
	for i := 0; i < 5; i++ {
		if o.add(h) {
			t.Error("unexpected result")
		}
	}
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 9; i++ {
		if o.add(h) {
			t.Error("unexpected result")
		}
	}
	if !o.add(h) {
		t.Error("unexpected result")
	}
}
