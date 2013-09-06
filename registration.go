package main

import (
	zmq "github.com/pebbe/zmq3"
)

type Registration struct {
	address       string
	Registrations chan *Client
}

func NewRegistration(address string) *Registration {
	return &Registration{address, make(chan *Client)}
}

func (self *Registration) Run() {
	pull, _ := zmq.NewSocket(zmq.PULL)
	defer pull.Close()
	pull.Bind(self.address)
	debug.Printf("[registration] Bound to %v", self.address)

	client_id_incr := 0

	for {
		parts, err := pull.RecvMessage(0)

		if err != nil {
			info.Printf("[registration] Recv error: %v", err)
			continue
		}

		if len(parts) != 2 {
			info.Printf("[registration] Invalid registration, should have 2 parts")
			continue
		}

		client_id_incr += 1

		self.Registrations <- NewClient(client_id_incr, parts[0], parts[1])
	}
}
