package metric

type DistMap map[StatKey]*Dist
type CounterMap map[StatKey]float64

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
	self.Counters[s.Key()] += s.Count
}

func (self TimestampedStats) AddValue(s StatValue) {
	self.Empty = false
	if d, ok := self.Dists[s.Key()]; ok {
		d.AddEntry(s.Value)
	} else {
		self.Dists[s.Key()] = NewDistFromValue(s.Value)
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

func (self StatValue) Key() StatKey {
	return StatKey(self.Name)
}

type StatCount struct {
	Name  string
	Count float64
}

func (self StatCount) Key() StatKey {
	return StatKey(self.Name)
}

type StatDist struct {
	Name string
	Dist [5]float64
}

func (self StatDist) Key() StatKey {
	return StatKey(self.Name)
}
