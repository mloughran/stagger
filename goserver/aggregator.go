package main

import "log"

type Stat struct {
	Name  string
	Value int
}

func RunAggregator(stats chan (Stat)) {
	for {
		s := <-stats
		log.Print("Stat: ", s.Name, " ", s.Value)
	}
}
