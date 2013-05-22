package main

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

func (self TimestampedStats) AddCount(s StatCount) {
	self.Counters[s.Name] += s.Count
}

func (self TimestampedStats) AddValue(s StatValue) {
	if d, present := self.Dists[s.Name]; present {
		d.AddEntry(s.Value)
	} else {
		self.Dists[s.Name] = NewDistFromValue(s.Value)
	}
}

func (self TimestampedStats) AddDist(s StatDist) {
	dist := ContstructDist(s.Dist)
	if d, present := self.Dists[s.Name]; present {
		d.Add(dist)
	} else {
		self.Dists[s.Name] = dist
	}
}

type Aggregate map[int64]*TimestampedStats

func (self Aggregate) AddTimestamp(ts int64) {
	self[ts] = NewTimestampedStats(ts)
}

func (self Aggregate) GetTimestamp(ts int64) *TimestampedStats {
	if timestamped_stats, present := self[ts]; present {
		return timestamped_stats
	} else {
		timestamped_stats := NewTimestampedStats(ts)
		self[ts] = timestamped_stats
		return timestamped_stats
	}
	return nil
}

func (self Aggregate) FinishTimestamp(ts int64) *TimestampedStats {
	if agg, ts_exists := self[ts]; ts_exists {
		delete(self, ts)
		return agg
	}
	return nil
}

func RunAggregator(statsc chan (*Stats), ts_complete chan (int64), ts_new chan (int64), output_chan chan (*TimestampedStats)) {
	aggregate := Aggregate{}

	for {
		select {
		case ts := <-ts_new:
			aggregate.AddTimestamp(ts)
		case stats := <-statsc:
			agg := aggregate.GetTimestamp(stats.Timestamp)

			for _, s := range stats.Values {
				agg.AddValue(s)
			}
			for _, s := range stats.Counts {
				agg.AddCount(s)
			}
			for _, s := range stats.Dists {
				agg.AddDist(s)
			}
		case ts := <-ts_complete:
			if agg := aggregate.FinishTimestamp(ts); agg != nil {
				debug.Printf("[aggregator] Finished aggregating data for ts %v", ts)
				output_chan <- agg
			}
		}
	}
}
