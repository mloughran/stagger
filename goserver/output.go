// Output received aggregated data from the aggregator and exposes it to downstream stagger processes, and to any other process which wishes to subscribe to data via ZMQ

package main

import (
	"bytes"
	zmq "github.com/pebbe/zmq3"
	msgpack "github.com/ugorji/go-msgpack"
	"log"
)

type OutputStat struct {
	Timestamp int64
	Dist      *Dist
}

func RunOutput(complete_chan chan (*TimestampedStats)) {
	pub, _ := zmq.NewSocket(zmq.PUB)

	pub.Bind("tcp://*:5563")

	for {
		stats := <-complete_chan

		log.Printf("[output] Data for ts %v", stats.Timestamp)
		for key, value := range stats.Counters {
			log.Printf("[output] %v: %v", key, value)
		}
		for key, value := range stats.Dists {
			log.Printf("[output] %v: %v", key, value)
		}

		for key, value := range stats.Dists {
			output_stat := OutputStat{stats.Timestamp, value}

			// Encode the dist as msgpack
			var b bytes.Buffer
			encoder := msgpack.NewEncoder(&b)
			encoder.Encode(output_stat)

			// TODO: Handle errors
			pub.Send(key, zmq.SNDMORE) // Use stat name as channel?
			pub.SendBytes(b.Bytes(), 0)
		}
	}
}