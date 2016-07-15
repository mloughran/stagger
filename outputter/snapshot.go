// Stores the last output and exposes it via http

package outputter

import (
	"encoding/json"
	"github.com/pusher/stagger/metric"
	"net/http"
	"strconv"
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
	var (
		b   []byte
		err error
	)

	s.mutex.Lock()
	if s.last != nil {
		b, err = json.Marshal(SnapshotFormat{s.last.Timestamp.Unix(), s.last})
	} else {
		b, err = json.Marshal(nil)
	}
	s.mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	if err != nil {
		w.WriteHeader(500)
	} else {
		w.WriteHeader(200)
	}

	w.Write(b)
}
