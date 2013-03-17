package main

import "log"

import zmq "github.com/pebbe/zmq3"

func RunClient(info ClientRef) {
	sock, _ := zmq.NewSocket(zmq.PAIR)
	defer sock.Close()
	log.Print("Connecting to ", info.Address)
	sock.Connect(info.Address)

	on_message := make(chan string)
	go func() {
		var s string
		for {
			// TODO: Handle error
			s, _ = sock.Recv(0)
			on_message <- s
		}
	}()
	
	for {
		select {
		case <- info.Mailbox:
			log.Print("Requesting stats from ", info.Address)
			sock.Send("Stats please!", zmq.DONTWAIT)
		case msg := <- on_message:
			log.Print(msg)
		}
	}
}
