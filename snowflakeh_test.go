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

func TestWorkerNextUnique(t *testing.T) {
	w := NewWorker(0)
	var n int64 = 0
	for i := 0; i < 10000; i++ {
		m, err := w.Next()
		if err != nil {
			t.Error(err)
		}
		if m <= n {
			t.Errorf("invalid: n=%d, m=%d", n, m)
		}
	}
}

func TestWorkerNextTimestamp(t *testing.T) {
	w := NewWorker(0)
	for i := 0; i < 10000; i++ {
		w.Next()
		lt1 := w.lastTimestamp
		sq1 := w.sequence
		w.Next()
		lt2 := w.lastTimestamp
		sq2 := w.sequence
		if lt1 == lt2 {
			if sq1 >= sq2 {
				t.Errorf("invalid sequence: %d, %d", sq1, sq2)
			}
			if sq1 > SequenceMask {
				t.Errorf("invalid sequence 1: %d", sq1)
			}
			if sq2 > SequenceMask {
				t.Errorf("invalid sequence 2: %d", sq2)
			}
		} else {
			if lt2 <= lt1 {
				t.Errorf("invalid lastTimestamp: %d", lt2)
			}
			if sq1 > SequenceMask {
				t.Errorf("invalid sequence 1: %d", sq1)
			}
			if sq2 != 0 {
				t.Errorf("invalid sequence 2: %d", sq2)
			}
		}
	}
}

func TestServerNew(t *testing.T) {
	s := NewServer(8181, 0, 1)
	if s.port != 8181 {
		t.Errorf("invalid port: %s", s.port)
	}
	if len(s.workers) != 1 {
		t.Errorf("invalid workers: %s", s.workers)
	}
	s = NewServer(8181, 0, 8)
	if s.port != 8181 {
		t.Errorf("invalid port: %s", s.port)
	}
	if len(s.workers) != 8 {
		t.Errorf("invalid workers: %s", s.workers)
	}
}

func TestServerWorkers(t *testing.T) {
	s := NewServer(8181, 0, 8)
	var lastWorker *Worker
	var lastServerId int = -1
	for i := 0; i < len(s.workers); i++ {
		w := <-s.workers
		if w == lastWorker {
			t.Errorf("invalid worker: %v", w)
		}
		lastWorker = w

		if w.serverId != lastServerId+1 {
			t.Errorf("invalid serverId: %d", w.serverId)
		}
		lastServerId = w.serverId

		if w.lastTimestamp != 0 {
			t.Errorf("invalid lastTimestamp: %d", w.lastTimestamp)
		}

		if w.sequence != 0 {
			t.Errorf("invalid serverId: %d", w.sequence)
		}
	}
}

func TestServerNextUnique(t *testing.T) {
	ids := make([]int64, 10000)
	s := NewServer(8181, 0, 2)
	for i := 0; i < 10000; i++ {
		m, err := s.Next()
		if err != nil {
			t.Error(err)
		}
		for _, n := range ids {
			if n == m {
				t.Errorf("duplicate id: %d", n)
			}
		}
		ids[i] = m
	}
}
