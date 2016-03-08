// Stores the last output and exposes it via http

package outputter

import (
	"encoding/json"
	"github.com/pusher/stagger/metric"
	"net/http"
	"sync"
)

type Snapshot struct {
	last  *metric.TimestampedStats
	mutex sync.Mutex
}

type SnapshotFormat struct {
	Timestamp int64
	*metric.TimestampedStats
}

func NewSnapshot() *Snapshot {
	return &Snapshot{}
}

func (s *Snapshot) Send(stats *metric.TimestampedStats) error {
	s.mutex.Lock()
	s.last = stats
	s.mutex.Unlock()
	return nil
}

func (s *Snapshot) String() string {
	return "snapshot"
}

func (s *Snapshot) ServeHTTP(w http.ResponseWriter, h *http.Request) {
	s.mutex.Lock()
	if b, err := json.Marshal(SnapshotFormat{s.last.Timestamp.Unix(), s.last}); err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
	s.mutex.Unlock()
}
