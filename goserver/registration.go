package main

import "log"
import "strings"

import msgpack "github.com/ugorji/go-msgpack"
import zmq "github.com/pebbe/zmq3"

type RegMsg struct {
	Address string
}

func StartRegistration(reg_chan chan (string)) {
	pull, _ := zmq.NewSocket(zmq.PULL)
	defer pull.Close()
	pull.Bind("tcp://127.0.0.1:2900")
	log.Print("Have bound")

	for {
		// Not sure if it's better to Recv or RecvBytes
		msg, _ := pull.Recv(0)
		log.Print("[registration] Received")

		// Convert into something with a Read method, bit horrible
		// buf := bytes.NewBuffer(msg)
		buf := strings.NewReader(msg)

		// Decode the msgpack data into Response struct
		var resp RegMsg
		dec := msgpack.NewDecoder(buf, nil)
		if err := dec.Decode(&resp); err != nil {
			log.Print(err)
			return
		}

		reg_chan <- resp.Address
	}
}
