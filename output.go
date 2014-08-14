// Output receives aggregated data from the aggregator and exposes it to downstream stagger processes, to any other process which wishes to subscribe to data via ZMQ, and to all configured outputters

package main

import (
	zmq "github.com/pebbe/zmq4"
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
