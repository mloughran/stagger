package main

import (
	zmq "github.com/pebbe/zmq3"
)

type ZmqClientEvents struct {
	OnMessage   chan ZMQMultipart
	OnClose     chan (bool)
	SendMessage chan ZMQMessage
}

type ZMQMultipart struct {
	Envelope string
	OnPart   chan string
	OnEnd    chan bool
}

type ZMQMessage struct {
	Method string
	Params []byte // TODO: Can we change to interface{} and pack here?
}

// Creates a Zmq client, runs it's gorouting, and returns the channels on 
// which you should communicate
func NewZmqClient(addr string) ZmqClientEvents {
	events := ZmqClientEvents{make(chan ZMQMultipart), make(chan bool), make(chan ZMQMessage)}

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
					// Convenience so you can send an empty last message to close (the
					// empty message will be ignored), not entirely sure about this
					if len(s) > 0 {
						multipart.OnPart <- s
					}
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

		// Note: I considered using sock.SendMessage but unfortunately that
		// doesn't support arbitrary flags, and it basically does this anway.

		// Set DONTWAIT so that Send doesn't block when HWM is reached

		// In case of an error and return, an OnClose will be sent by defer

		if _, err := sock.Send(msg.Method, zmq.DONTWAIT|zmq.SNDMORE); err != nil {
			info.Print("Closing client ", addr, " after ", err)
			closed = true
			return
		}
		if _, err := sock.SendBytes(msg.Params, zmq.DONTWAIT); err != nil {
			info.Print("Closing client ", addr, " after ", err)
			closed = true
			return
		}
	}
}
