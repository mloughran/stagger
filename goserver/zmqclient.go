package main

import "log"

import zmq "github.com/pebbe/zmq3"

type ZmqClientEvents struct {
	OnMessage   chan (string)
	OnClose     chan (bool)
	SendMessage chan (string)
}

// Creates a Zmq client, runs it's gorouting, and returns the channels on 
// which you should communicate
func NewZmqClient(addr string) ZmqClientEvents {
	events := ZmqClientEvents{make(chan string), make(chan bool), make(chan string)}

	// on_message := make(chan string)
	// on_close := make(chan bool)
	// send_message := make(chan string)

	go RunZmqClient(addr, events)

	return events
}

// 
func RunZmqClient(addr string, events ZmqClientEvents) {
	sock, _ := zmq.NewSocket(zmq.PAIR)

	// This could help with shutdown since at least read would exit...
	// TODO: Configurable?
	sock.SetRcvtimeo(1e9)

	// Set a hwm of 1 - it's pointless to buffer requests for stats, and this
	// means that we get an error when we try to send if the client has gone
	sock.SetSndhwm(1)

	// Removing because it causes crash... (due to read goroutine)
	// defer sock.Close()

	log.Print("Connecting to ", addr)
	sock.Connect(addr)

	// TODO: This isn't correct because sock not threadsafe...
	on_message := make(chan string)
	closed := false
	go func() {
		defer sock.Close()
		for {
			if closed {
				return
			}

			// We expect an error when the rcv timeout is exceeded
			s, err := sock.Recv(0)
			if err == nil {
				on_message <- s
			}
		}
	}()

	defer func() {
		events.OnClose <- true
	}()

	for {
		select {
		case msg := <-events.SendMessage:
			// Set DONTWAIT so that Send doesn't block when HWM is reached
			if _, err := sock.Send(msg, zmq.DONTWAIT); err != nil {
				log.Print("Closing client ", addr, " after ", err)
				closed = true

				return
			}
		case msg := <-on_message:
			events.OnMessage <- msg
		}
	}
}
