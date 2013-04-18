package main

import (
	zmq "github.com/pebbe/zmq3"
	msgpack "github.com/ugorji/go-msgpack"
	"strings"
)

type RegMsg struct {
	Name    string
	Address string
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
		msg, _ := pull.Recv(0)

		// Convert into something with a Read method, bit horrible
		// buf := bytes.NewBuffer(msg)
		buf := strings.NewReader(msg)

		// Decode the msgpack data into Response struct
		var resp RegMsg
		dec := msgpack.NewDecoder(buf, nil)
		if err := dec.Decode(&resp); err != nil {
			info.Print(err)
			return
		}

		reg_chan <- Registration(resp)
	}
}
