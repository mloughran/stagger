// Output received aggregated data from the aggregator and exposes it to downstream stagger processes, and to any other process which wishes to subscribe to data via ZMQ

package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq3"
	"log"
	"sort"
)

type OutputStat struct {
	Timestamp int64
	Dist      *Dist
}

type Outputter interface {
	Send(*TimestampedStats)
}

type Output struct {
	outputs []Outputter
}

func NewOutput() *Output {
	return &Output{[]Outputter{}}
}

func (o *Output) Run(complete_chan <-chan (*TimestampedStats)) {
	pub, _ := zmq.NewSocket(zmq.PUB)

	pub.Bind("tcp://*:5563")

	for stats := range complete_chan {
		// Logging (temp)
		// TODO: Extract into outputter
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

		// ZMQ pubsub
		// TODO: Extract as outputter
		for key, value := range stats.Dists {
			output_stat := OutputStat{stats.Timestamp, value}

			if b, err := marshal(output_stat); err == nil {
				pub.Send(key, zmq.SNDMORE) // Use stat name as channel?
				pub.SendBytes(b, 0)
			} else {
				info.Printf("Error encoding as msgpack: %v", output_stat)
			}
		}

		for _, op := range o.outputs {
			op.Send(stats)
		}
	}
}

func (o *Output) Add(op Outputter) {
	o.outputs = append(o.outputs, op)
}
