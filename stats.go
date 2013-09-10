package main

type Count float64

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
	if d, ok := self.Dists[s.Name]; ok {
		d.AddEntry(s.Value)
	} else {
		self.Dists[s.Name] = NewDistFromValue(s.Value)
	}
}

func (self TimestampedStats) AddDist(s StatDist) {
	self.Empty = false
	dist := ContstructDist(s.Dist)
	if d, ok := self.Dists[s.Name]; ok {
		d.Add(dist)
	} else {
		self.Dists[s.Name] = dist
	}
}

type Stats struct {
	Timestamp int64
	Values    []StatValue
	Counts    []StatCount
	Dists     []StatDist
}

// TODO: Needs weight
type StatValue struct {
	Name  string
	Value float64
}

type StatCount struct {
	Name  string
	Count float64
}

type StatDist struct {
	Name string
	Dist [5]float64
}
