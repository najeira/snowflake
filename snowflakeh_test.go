package main

import (
	"testing"
	"time"
)

func TestNow(t *testing.T) {
	u := time.Now().Unix() * 1000
	n := now()
	delta := n - u
	if delta < 0 || (1000*1000) < delta {
		t.Errorf("invalid now(): now=%d, time=%d", n, u)
	}
}
