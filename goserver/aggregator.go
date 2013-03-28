package main

import "log"

type Stat struct {
	Timestamp int64
	Name      string
	Value     int
}

func RunAggregator(stats chan (Stat)) {
	for {
		s := <-stats
		log.Print("[aggregator] Stat: ", s)
	}
}
