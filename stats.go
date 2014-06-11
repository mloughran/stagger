package main

type Count float64

type DistMap map[string]*Dist
type CounterMap map[string]float64
type TypeMap map[string]*string

type TimestampedStats struct {
	Timestamp int64
	Dists     DistMap
	Counters  CounterMap
	Types     TypeMap
	Empty     bool
}

var TypeCache TypeMap

func init() {
	TypeCache = make(TypeMap)
}

func NewTimestampedStats(ts int64) *TimestampedStats {
	return &TimestampedStats{
		ts,
		DistMap{},
		CounterMap{},
		TypeMap{},
		true,
	}
}

func NewTimestampedStatsWithTypes(ts int64) *TimestampedStats {
	return &TimestampedStats{
		ts,
		DistMap{},
		CounterMap{},
		TypeCache,
		true,
	}
}

func (self TimestampedStats) AddCount(s StatCount) {
	self.Empty = false
	self.Counters[s.Name] += s.Count
	if s.Type != nil {
		self.Types[s.Name] = s.Type
		TypeCache[s.Name] = s.Type
	}
}

func (self TimestampedStats) AddValue(s StatValue) {
	self.Empty = false
	if d, ok := self.Dists[s.Name]; ok {
		d.AddEntry(s.Value)
	} else {
		self.Dists[s.Name] = NewDistFromValue(s.Value)
	}
	if s.Type != nil {
		self.Types[s.Name] = s.Type
		TypeCache[s.Name] = s.Type
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
	if s.Type != nil {
		self.Types[s.Name] = s.Type
		TypeCache[s.Name] = s.Type
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
	Type  *string
}

type StatCount struct {
	Name  string
	Count float64
	Type  *string
}

type StatDist struct {
	Name string
	Dist [5]float64
	Type *string
}
