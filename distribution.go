package main

import (
	"fmt"
	"math"
)

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

func ContstructDist(vs [5]float64) *Dist {
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
	return fmt.Sprintf("Distribution: mean: %.5g, sd: %.5g, min/max: %.5g/%.5g (weight %.5g)", self.Mean(), self.Sd(), self.Min, self.Max, self.N)
}
