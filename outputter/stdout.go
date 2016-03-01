package outputter

import (
	"fmt"
	"github.com/pusher/stagger/metric"
	"log"
	"sort"
)

type stdout bool

var Stdout = stdout(true)

func (self stdout) Send(stats *metric.TimestampedStats) {
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
