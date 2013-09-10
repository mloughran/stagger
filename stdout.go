package main

import (
	"fmt"
	"log"
	"sort"
)

type StdOut struct{}

func NewStdOut() *StdOut {
	return &StdOut{}
}

func (self *StdOut) Send(stats *TimestampedStats) {
	var heading = fmt.Sprintf("[output] (ts:%v) Aggregated data:\n", stats.Timestamp)

	var output []string
	for key, value := range stats.Counters {
		output = append(output, fmt.Sprintf("%v: %.5g\n", key, value))
	}
	for key, value := range stats.Dists {
		output = append(output, fmt.Sprintf("%v: %v\n", key, value))
	}
	sort.Strings(output)
	log.Print(heading, output)
}
