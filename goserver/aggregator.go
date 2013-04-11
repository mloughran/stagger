package main

import (
	"log"
)

type StatIdentifier struct {
	Timestamp int64
	Name      string
}

type ValueStat struct {
	*StatIdentifier
	Value float64
}

type CounterStat struct {
	*StatIdentifier
	Count float64
}

type DistStat struct {
	*StatIdentifier
	Dist []float64
}

type StatsChannels struct {
	ValueStats   chan (ValueStat)
	CounterStats chan (CounterStat)
	DistStats    chan (DistStat)
}

func NewStatsChannels() StatsChannels {
	return StatsChannels{
		make(chan ValueStat),
		make(chan CounterStat),
		make(chan DistStat),
	}
}

type TimestampedStats struct {
	Timestamp int64
	Dists     map[string]*Dist
	Counters  map[string]float64
}

func NewTimestampedStats(ts int64) *TimestampedStats {
	return &TimestampedStats{
		ts,
		map[string]*Dist{},
		map[string]float64{},
	}
}

type Aggregate map[int64]*TimestampedStats

func (self Aggregate) AddTimestamp(ts int64) {
	self[ts] = NewTimestampedStats(ts)
}

func (self Aggregate) FinishTimestamp(ts int64) *TimestampedStats {
	data := self[ts]
	delete(self, ts)
	return data
}

func (self Aggregate) AddCounter(s CounterStat) {
	self[s.Timestamp].Counters[s.Name] += s.Count
}

func (self Aggregate) AddValue(s ValueStat) {
	if _, present := self[s.Timestamp].Dists[s.Name]; present {
		dist := self[s.Timestamp].Dists[s.Name]
		dist.AddEntry(s.Value)
	} else {
		self[s.Timestamp].Dists[s.Name] = NewDistFromValue(s.Value)
	}
}

func (self Aggregate) AddDist(s DistStat) {
	dist := ContstructDist(s.Dist)
	if _, present := self[s.Timestamp].Dists[s.Name]; present {
		dist := self[s.Timestamp].Dists[s.Name]
		dist.Add(dist)
	} else {
		self[s.Timestamp].Dists[s.Name] = dist
	}
}

func RunAggregator(stats StatsChannels, ts_complete chan (int64), ts_new chan (int64), output_chan chan (*TimestampedStats)) {
	aggregate := Aggregate{}

	for {
		select {
		case ts := <-ts_new:
			aggregate.AddTimestamp(ts)
		case s := <-stats.CounterStats:
			aggregate.AddCounter(s)
		case s := <-stats.ValueStats:
			aggregate.AddValue(s)
		case s := <-stats.DistStats:
			aggregate.AddDist(s)
		case ts := <-ts_complete:
			data := aggregate.FinishTimestamp(ts)
			log.Print("[aggregator] Finished aggregating data for timestamp: ", ts)
			output_chan <- data
		}
	}
}
