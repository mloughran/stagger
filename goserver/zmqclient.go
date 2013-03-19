package main

import "log"

import zmq "github.com/pebbe/zmq3"

type ZmqClientEvents struct {
	OnMessage   chan ZMQMultipart
	OnClose     chan (bool)
	SendMessage chan (string)
}

type ZMQMultipart struct {
	Envelope string
	OnPart   chan string
	OnEnd    chan bool
}

// Creates a Zmq client, runs it's gorouting, and returns the channels on 
// which you should communicate
func NewZmqClient(addr string) ZmqClientEvents {
	events := ZmqClientEvents{make(chan ZMQMultipart), make(chan bool), make(chan string)}

	go RunZmqClient(addr, events)

	return events
}

// 
func RunZmqClient(addr string, events ZmqClientEvents) {
	sock, _ := zmq.NewSocket(zmq.PAIR)

	// Set a hwm of 1 - it's pointless to buffer requests for stats, and this
	// means that we get an error when we try to send if the client has gone
	sock.SetSndhwm(1)

	// Removing because it causes crash... (due to read goroutine)
	// defer sock.Close()

	log.Print("Connecting to ", addr)
	sock.Connect(addr)

	// These are used so that we can stop the reading goroutine - hacky...
	// TODO: Configurable?
	sock.SetRcvtimeo(1e9)
	closed := false

	// TODO: This isn't correct because sock not threadsafe...
	go func() {
		defer sock.Close()

		var multipart_in_progress bool = false
		var multipart ZMQMultipart
		var more bool

		for {
			if closed {
				return
			}

			// We expect an error when the rcv timeout is exceeded
			s, err := sock.Recv(0)
			if err == nil {
				if multipart_in_progress == false {
					multipart_in_progress = true
					multipart = ZMQMultipart{s, make(chan string), make(chan bool)}
					events.OnMessage <- multipart
				} else {
					multipart.OnPart <- s
				}

				if more, _ = sock.GetRcvmore(); more == false {
					multipart.OnEnd <- true
					multipart_in_progress = false
				}
			}
		}
	}()

	defer func() {
		events.OnClose <- true
	}()

	for {
		msg := <-events.SendMessage
		// Set DONTWAIT so that Send doesn't block when HWM is reached
		if _, err := sock.Send(msg, zmq.DONTWAIT); err != nil {
			log.Print("Closing client ", addr, " after ", err)
			closed = true

			return
		}
	}
}
