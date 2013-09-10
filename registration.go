package main

import (
	zmq "github.com/pebbe/zmq3"
)

type Registration struct {
	address       string
	Registrations chan string
}

func NewRegistration(address string) *Registration {
	return &Registration{address, make(chan string)}
}

func (self *Registration) Run() {
	pull, _ := zmq.NewSocket(zmq.PULL)
	defer pull.Close()
	pull.Bind(self.address)
	debug.Printf("[registration] Bound to %v", self.address)

	for {
		if parts, err := pull.RecvMessage(0); err != nil {
			info.Printf("[registration] Recv error: %v", err)
		} else {
			if len(parts) != 2 {
				info.Printf("[registration] Invalid registration, should have 2 parts")
			} else {
				// TODO: Use the 2nd part
				self.Registrations <- parts[0]
			}
		}
	}
}
