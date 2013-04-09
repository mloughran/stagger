// Output received aggregated data from the aggregator and exposes it to downstream stagger processes, and to any other process which wishes to subscribe to data

package main

import (
	"log"
)

func RunOutput(complete_chan chan (*TimestampedStats)) {
	for {
		stats := <-complete_chan

		log.Printf("[output] Data for ts %v", stats.Timestamp)
		for key, value := range stats.Counters {
			log.Printf("[output] %v: %v", key, value)
		}
		for key, value := range stats.Dists {
			log.Printf("[output] %v: %v", key, value)
		}
	}
}
