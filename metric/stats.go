package metric

import (
	"github.com/pusher/stagger/conn"
	"time"
)

type DistMap map[conn.StatKey]*Dist
type CounterMap map[conn.StatKey]float64

// The object that collects all the stats emitted by clients
type TimestampedStats struct {
	Timestamp time.Time
	Dists     DistMap
	Counters  CounterMap
	Empty     bool
}

func NewTimestampedStats(t time.Time) *TimestampedStats {
	return &TimestampedStats{t, DistMap{}, CounterMap{}, true}
}

func (self TimestampedStats) AddCount(s conn.StatCount) {
	self.Empty = false
	self.Counters[s.Key()] += s.Count
}

func (self TimestampedStats) AddValue(s conn.StatValue) {
	self.Empty = false
	if d, ok := self.Dists[s.Key()]; ok {
		d.AddEntry(s.Value)
	} else {
		self.Dists[s.Key()] = NewDistFromValue(s.Value)
	}
}

func (self TimestampedStats) AddDist(s conn.StatDist) {
	self.Empty = false
	dist := ContstructDist(s.Dist)
	if d, ok := self.Dists[s.Key()]; ok {
		d.Add(dist)
	} else {
		self.Dists[s.Key()] = dist
	}
}
