package main

import "log"
// import "strings"

// import msgpack "github.com/ugorji/go-msgpack"
import zmq "github.com/pebbe/zmq3"

type Client struct {
	Address string
}

func NewClient(address string) *Client {
	return &Client{address}
}

func (self *Client) Connect() {
	sock, _ := zmq.NewSocket(zmq.PAIR)
	defer sock.Close()
	sock.Connect(self.Address)
	
	// self.Log("Connected to ")
	log.Print("Connected to ", self.Address)
}