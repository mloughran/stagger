package main

import (
	"log"
	"math"
)

type Stat struct {
	Timestamp int64
	Name      string
	Value     float64
}

func RunAggregator(stats chan (Stat), ts_complete chan (int64), ts_new chan (int64)) {
	aggregates := map[int64](map[string]*Dist){}

	for {
		select {
		case ts := <-ts_new:
			aggregates[ts] = map[string]*Dist{}
		case s := <-stats:
			log.Print("[aggregator] Stat: ", s)
			if _, present := aggregates[s.Timestamp][s.Name]; present {
				dist := aggregates[s.Timestamp][s.Name]
				dist.AddEntry(s.Value)
			} else {
				aggregates[s.Timestamp][s.Name] = NewDistFromValue(s.Value)
			}
		case ts := <-ts_complete:
			log.Print("[aggregator] Complete for ts ", ts)
			log.Print("[aggregator] Final data: ", aggregates[ts])
			for key, value := range aggregates[ts] {
				log.Printf("[aggregator] %v: %v", key, value)
			}
			// Delete the data, we may now want to do this immediately?
			delete(aggregates, ts)
		}
	}
}

type Dist struct {
	N      int
	Min    float64
	Max    float64
	Sum_x  float64
	Sum_x2 float64
}

func NewDistFromValue(v float64) *Dist {
	return &Dist{1, v, v, v, v * v}
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

func (self *Dist) SD() {

}
