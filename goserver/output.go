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

func RunOutput(complete_chan chan (*TimestampedStats), librato *Librato) {
	pub, _ := zmq.NewSocket(zmq.PUB)

	pub.Bind("tcp://*:5563")

	// Using a buffered channel to isolate slowness posting to librato
	librato_chan := make(chan *TimestampedStats, 100)
	if librato != nil {
		go librato.Run(librato_chan)
	}

	for stats := range complete_chan {
		// Logging (temp)

		log.Printf("[output] Data for ts %v", stats.Timestamp)
		for key, value := range stats.Counters {
			log.Printf("[output] %v: %v", key, value)
		}
		for key, value := range stats.Dists {
			log.Printf("[output] %v: %v", key, value)
		}

		// ZMQ pubsub

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

		// LIBRATO
		if librato != nil {
			librato_chan <- stats
		}
	}
}
