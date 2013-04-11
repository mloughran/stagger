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

func RunAggregator(stats StatsChannels, ts_complete chan (int64), ts_new chan (int64), output_chan chan (*TimestampedStats)) {
	aggregates := map[int64]*TimestampedStats{}

	for {
		select {
		case ts := <-ts_new:
			aggregates[ts] = NewTimestampedStats(ts)
		case s := <-stats.CounterStats:
			aggregates[s.Timestamp].Counters[s.Name] += s.Count
		case s := <-stats.ValueStats:
			if _, present := aggregates[s.Timestamp].Dists[s.Name]; present {
				dist := aggregates[s.Timestamp].Dists[s.Name]
				dist.AddEntry(s.Value)
			} else {
				aggregates[s.Timestamp].Dists[s.Name] = NewDistFromValue(s.Value)
			}
		case s := <-stats.DistStats:
			dist := ContstructDist(s.Dist)
			if _, present := aggregates[s.Timestamp].Dists[s.Name]; present {
				dist := aggregates[s.Timestamp].Dists[s.Name]
				dist.Add(dist)
			} else {
				aggregates[s.Timestamp].Dists[s.Name] = dist
			}
		case ts := <-ts_complete:
			log.Print("[aggregator] Finished aggregating data for timestamp: ", ts)
			output_chan <- aggregates[ts]
			delete(aggregates, ts)
		}
	}
}
