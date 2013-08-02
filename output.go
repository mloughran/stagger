// Output received aggregated data from the aggregator and exposes it to downstream stagger processes, and to any other process which wishes to subscribe to data via ZMQ

package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq3"
	msgpack "github.com/ugorji/go-msgpack"
	"log"
	"sort"
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

		var output []string
		for key, value := range stats.Counters {
			output = append(output, fmt.Sprintf("%v: %v\n", key, value))
		}
		for key, value := range stats.Dists {
			output = append(output, fmt.Sprintf("%v: %v\n", key, value))
		}
		sort.Strings(output)
		log.Print(output)

		// ZMQ pubsub
		for key, value := range stats.Dists {
			output_stat := OutputStat{stats.Timestamp, value}
			b, _ := msgpack.Marshal(output_stat)

			// TODO: Handle errors
			pub.Send(key, zmq.SNDMORE) // Use stat name as channel?
			pub.SendBytes(b, 0)
		}

		// LIBRATO
		if librato != nil {
			librato_chan <- stats
		}
	}
}
