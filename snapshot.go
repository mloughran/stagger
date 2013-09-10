// Stores the last output and exposes it via http

package main

import (
	"encoding/json"
	"net/http"
	"sync"
)

type Snapshot struct {
	last  *TimestampedStats
	mutex sync.Mutex
}

func NewSnapshot() *Snapshot {
	return &Snapshot{}
}

func (s *Snapshot) Send(stats *TimestampedStats) {
	s.mutex.Lock()
	s.last = stats
	s.mutex.Unlock()
}

func (s *Snapshot) ServeHTTP(w http.ResponseWriter, h *http.Request) {
	s.mutex.Lock()
	if b, err := json.Marshal(s.last); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
	s.mutex.Unlock()
}
