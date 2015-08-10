package main

import (
	"strings"
)

type Count float64

// Format: key,tag1=value1,tag2=value2
//
// We assume the tag keys are sorted alphabetically
// We assume the tag keys and values don't contain commas
// We assume the tag keys and values are separated by '='
type StatKey string

func (self StatKey) Name() string {
	return strings.SplitN(self.String(), ",", 2)[0]
}

func (self StatKey) Tags() (tags map[string]string) {
	pairs := strings.Split(self.String(), ",")
	if len(pairs) <= 1 {
		return
	}
	pairs = pairs[1:] // Remove name
	tags = make(map[string]string, len(pairs))
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			tags[kv[0]] = kv[1]
		} // Ignore malformed data otherwise
	}
	return
}

func (self StatKey) HasTags() bool {
	return strings.ContainsRune(self.String(), ',')
}

func (self StatKey) String() string {
	return string(self)
}

type DistMap map[StatKey]*Dist
type CounterMap map[StatKey]float64
type TypeMap map[StatKey]*string

type TimestampedStats struct {
	Timestamp int64
	Dists     DistMap
	Counters  CounterMap
	Types     TypeMap
	Empty     bool
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

func (self TimestampedStats) AddCount(s StatCount) {
	self.Empty = false
	self.Counters[s.Key()] += s.Count
	if s.Type != nil {
		self.Types[s.Key()] = s.Type
	}
}

func (self TimestampedStats) AddValue(s StatValue) {
	self.Empty = false
	if d, ok := self.Dists[s.Key()]; ok {
		d.AddEntry(s.Value)
	} else {
		self.Dists[s.Key()] = NewDistFromValue(s.Value)
	}
	if s.Type != nil {
		self.Types[s.Key()] = s.Type
	}
}

func (self TimestampedStats) AddDist(s StatDist) {
	self.Empty = false
	dist := ContstructDist(s.Dist)
	if d, ok := self.Dists[s.Key()]; ok {
		d.Add(dist)
	} else {
		self.Dists[s.Key()] = dist
	}
	if s.Type != nil {
		self.Types[s.Key()] = s.Type
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

func (self StatValue) Key() StatKey {
	return StatKey(self.Name)
}

type StatCount struct {
	Name  string
	Count float64
	Type  *string
}

func (self StatCount) Key() StatKey {
	return StatKey(self.Name)
}

type StatDist struct {
	Name string
	Dist [5]float64
	Type *string
}

func (self StatDist) Key() StatKey {
	return StatKey(self.Name)
}
