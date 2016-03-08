// Responsible for receiving stats (on a channel), aggregating them into
// snapshots per time interval, and outputting data for completed intervals on
// the output channel.
//
// If data is received for an already report snapshop, an error is logged and
// the data is discarded.

package main

import (
	"github.com/pusher/stagger/metric"
	"time"
)

type Count float64

type Aggregator struct {
	output  chan (*metric.TimestampedStats)
	current *metric.TimestampedStats
	Stats   chan (*metric.Stats)
}

func NewAggregator() *Aggregator {
	return &Aggregator{
		output: make(chan *metric.TimestampedStats),
		Stats:  make(chan *metric.Stats),
	}
}

func (self *Aggregator) Run(tsComplete <-chan time.Time, tsNew <-chan time.Time) {
	for {
		select {
		case t := <-tsNew:
			self.newInterval(t)
		case stats := <-self.Stats:
			self.feed(stats)
		case t := <-tsComplete:
			self.report(t)
		}
	}
}

func (self *Aggregator) newInterval(t time.Time) {
	if self.current != nil && !self.current.Empty {
		self.report(self.current.Timestamp)
	}
	self.current = metric.NewTimestampedStats(t)
}

func (self *Aggregator) feed(stats *metric.Stats) {
	if self.current == nil || stats.Timestamp != self.current.Timestamp.Unix() {
		info.Printf(
			"[aggregator] (ts:%v) Stats received for unexpected timestamp %v, discarding",
			self.current.Timestamp,
			stats.Timestamp,
		)
		return
	}

	for _, s := range stats.Values {
		self.current.AddValue(s)
	}
	for _, s := range stats.Counts {
		self.current.AddCount(s)
	}
	for _, s := range stats.Dists {
		self.current.AddDist(s)
	}
}

func (self *Aggregator) report(t time.Time) {
	if self.current == nil {
		panic("Missing timestamped stats to report")
	}
	if t != self.current.Timestamp {
		info.Printf("[aggregator] ERROR, impossible timestamp, current is %v, got finish for %v", self.current.Timestamp, t)
		return
	}
	debug.Printf("[aggregator] (ts:%v) Finished aggregating data", t)
	self.output <- self.current
	self.current = nil
}

// These functions are for internal reporting

// TODO: Not sure about these functions
func (self *Aggregator) Count(ts int64, name string, value Count) {
	self.Stats <- &metric.Stats{
		Timestamp: ts,
		Counts:    []metric.StatCount{metric.StatCount{name, float64(value)}},
	}
}

func (self *Aggregator) Value(ts int64, name string, value float64) {
	self.Stats <- &metric.Stats{
		Timestamp: ts,
		Values:    []metric.StatValue{metric.StatValue{name, value}},
	}
}
