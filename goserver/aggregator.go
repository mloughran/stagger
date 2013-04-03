package main

import (
	"fmt"
	"log"
	"math"
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
	Dists    map[string]*Dist
	Counters map[string]float64
}

func NewTimestampedStats() TimestampedStats {
	return TimestampedStats{
		map[string]*Dist{},
		map[string]float64{},
	}
}

func RunAggregator(stats StatsChannels, ts_complete chan (int64), ts_new chan (int64)) {
	aggregates := map[int64]TimestampedStats{}

	for {
		select {
		case ts := <-ts_new:
			aggregates[ts] = NewTimestampedStats()
		case s := <-stats.CounterStats:
			log.Print("[aggregator] Stat: ", s)
			aggregates[s.Timestamp].Counters[s.Name] += s.Count
		case s := <-stats.ValueStats:
			log.Print("[aggregator] Stat: ", s)
			if _, present := aggregates[s.Timestamp].Dists[s.Name]; present {
				dist := aggregates[s.Timestamp].Dists[s.Name]
				dist.AddEntry(s.Value)
			} else {
				aggregates[s.Timestamp].Dists[s.Name] = NewDistFromValue(s.Value)
			}
		case s := <-stats.DistStats:
			log.Print("[aggregator] Stat: ", s)
			dist := ContstructDist(s.Dist)
			if _, present := aggregates[s.Timestamp].Dists[s.Name]; present {
				dist := aggregates[s.Timestamp].Dists[s.Name]
				dist.Add(dist)
			} else {
				aggregates[s.Timestamp].Dists[s.Name] = dist
			}
		case ts := <-ts_complete:
			log.Print("[aggregator] Complete for ts ", ts)
			log.Print("[aggregator] Final data: ", aggregates[ts])
			for key, value := range aggregates[ts].Counters {
				log.Printf("[aggregator] %v: %v", key, value)
			}
			for key, value := range aggregates[ts].Dists {
				log.Printf("[aggregator] %v: %v", key, value)
			}
			// Delete the data, we may now want to do this immediately?
			delete(aggregates, ts)
		}
	}
}

type Dist struct {
	N      float64 // weight
	Min    float64
	Max    float64
	Sum_x  float64
	Sum_x2 float64
}

func NewDistFromValue(v float64) *Dist {
	return &Dist{1, v, v, v, v * v}
}

func ContstructDist(vs []float64) *Dist {
	return &Dist{
		vs[0],
		vs[1],
		vs[2],
		vs[3],
		vs[4],
	}
}

func (self *Dist) AddEntry(v float64) {
	self.Min = math.Min(self.Min, v)
	self.Max = math.Max(self.Max, v)
	self.Sum_x += v
	self.Sum_x2 += v * v
	self.N += 1
}

func (self *Dist) Add(dist *Dist) {
	self.Min = math.Min(self.Min, dist.Min)
	self.Max = math.Max(self.Max, dist.Max)
	self.Sum_x += dist.Sum_x
	self.Sum_x2 += dist.Sum_x2
	self.N += dist.N
}

func (self *Dist) Mean() float64 {
	return self.Sum_x / float64(self.N)
}

func (self *Dist) Sd() float64 {
	mean_x_sq := self.Sum_x2 / self.N
	mean_sq := math.Pow(self.Mean(), 2)
	return math.Sqrt(math.Max(mean_x_sq-mean_sq, 0))
}

func (self *Dist) String() string {
	return fmt.Sprintf("Distribution: mean: %v, sd: %v, min/max: %v/%v (weight %v)", self.Mean(), self.Sd(), self.Min, self.Max, self.N)
}
