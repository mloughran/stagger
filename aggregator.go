// Responsible for receiving stats (on a channel), aggregating them into snapshots per time interval, and outputting data for completed intervals on the output channel.
//
// * At any point in time the aggregator is aggregating stats into 2 snapshots - passed & next. Passed is the last tick timestamp, which is waiting for all survey data to be reported. Stats reported without an associated timestamp go into next. The next snapshot is propagated to passed when the next tick occurs, and will be outputted when all survey data has also been received.
// * If data is received for an already report snapshop, an error is logged and the data is discarded.

package main

type DistMap map[string]*Dist
type CounterMap map[string]float64

type TimestampedStats struct {
	Timestamp int64
	Dists     DistMap
	Counters  CounterMap
	Empty     bool
}

func NewTimestampedStats(ts int64) *TimestampedStats {
	return &TimestampedStats{
		ts,
		DistMap{},
		CounterMap{},
		true,
	}
}

func (self TimestampedStats) AddCount(s StatCount) {
	self.Empty = false
	self.Counters[s.Name] += s.Count
}

func (self TimestampedStats) AddValue(s StatValue) {
	self.Empty = false
	if d, present := self.Dists[s.Name]; present {
		d.AddEntry(s.Value)
	} else {
		self.Dists[s.Name] = NewDistFromValue(s.Value)
	}
}

func (self TimestampedStats) AddDist(s StatDist) {
	self.Empty = false
	dist := ContstructDist(s.Dist)
	if d, present := self.Dists[s.Name]; present {
		d.Add(dist)
	} else {
		self.Dists[s.Name] = dist
	}
}

type Aggregator struct {
	output   chan (*TimestampedStats)
	passed   *TimestampedStats
	passedTs int64
	next     *TimestampedStats
}

func newAggregator(output_chan chan (*TimestampedStats)) *Aggregator {
	return &Aggregator{
		output: output_chan,
		next:   NewTimestampedStats(-1),
	}
}

func RunAggregator(statsc chan (*Stats), ts_complete chan (int64), ts_new chan (int64), output_chan chan (*TimestampedStats)) {
	aggregator := newAggregator(output_chan)

	for {
		select {
		case ts := <-ts_new:
			aggregator.newInterval(ts)
		case stats := <-statsc:
			aggregator.feed(stats)
		case ts := <-ts_complete:
			aggregator.report(ts)
		}
	}
}

func (self *Aggregator) newInterval(ts int64) {
	if self.passed != nil && !self.passed.Empty {
		self.report(self.passed.Timestamp)
	}
	self.passedTs = ts
	self.passed = self.next
	self.passed.Timestamp = ts
	self.next = NewTimestampedStats(-1)
}

func (self *Aggregator) feed(stats *Stats) {
	if stats.Timestamp == self.passedTs {
		for _, s := range stats.Values {
			self.passed.AddValue(s)
		}
		for _, s := range stats.Counts {
			self.passed.AddCount(s)
		}
		for _, s := range stats.Dists {
			self.passed.AddDist(s)
		}
	} else {
		info.Printf("[aggregator] (ts:%v) Stats received for unexpected timestamp, discarding", stats.Timestamp)
	}
}

func (self *Aggregator) report(ts int64) {
	if ts == self.passed.Timestamp {
		debug.Printf("[aggregator] (ts:%v) Finished aggregating data", ts)
		self.output <- self.passed
		self.passed = nil
		self.passedTs = -1
	} else {
		info.Printf("[aggregator] ERROR, impossible timestamp, passedTs is %v, got finish for %v", self.passedTs, ts)
	}
}
