package main

import "log"

type Stat struct {
	Timestamp int64
	Name      string
	Value     int
}

func RunAggregator(stats chan (Stat), ts_complete chan (int64), ts_new chan (int64)) {
	aggregates := map[int64](map[string]int){}

	for {
		select {
		case ts := <-ts_new:
			aggregates[ts] = map[string]int{}
		case s := <-stats:
			log.Print("[aggregator] Stat: ", s)
			aggregates[s.Timestamp][s.Name] += s.Value
		case ts := <-ts_complete:
			log.Print("[aggregator] Complete for ts ", ts)
			log.Print("[aggregator] Final data: ", aggregates[ts])
			// Delete the data, we may now want to do this immediately?
			delete(aggregates, ts)
		}
	}
}
