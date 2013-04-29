package main

import (
	zmq "github.com/pebbe/zmq3"
)

type RegMsg struct {
	Address  string
	Metadata string
}

// Alias the type to decouple internals from the ZMQ message
type Registration RegMsg

func StartRegistration(address string, reg_chan chan (Registration)) {
	pull, _ := zmq.NewSocket(zmq.PULL)
	defer pull.Close()
	pull.Bind(address)
	debug.Printf("[registration] Bound to %v", address)

	for {
		// Not sure if it's better to Recv or RecvBytes
		parts, err := pull.RecvMessage(0)

		if err != nil {
			info.Printf("[registration] Recv error: %v", err)
			continue
		}

		if len(parts) != 2 {
			info.Printf("[registration] Invalid registration, should have 2 parts")
			continue
		}

		reg_chan <- Registration(RegMsg{parts[0], parts[1]})
	}
}
