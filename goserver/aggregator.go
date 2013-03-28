package main

import "log"

type Stat struct {
	Timestamp int64
	Name      string
	Value     int
}

func RunAggregator(stats chan (Stat), ts_complete chan (int64)) {
	for {
		select {
		case s := <-stats:
			log.Print("[aggregator] Stat: ", s)
		case ts := <-ts_complete:
			log.Print("[aggregator] Complete for ts ", ts)
		}
	}
}
