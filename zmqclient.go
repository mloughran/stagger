package main

import (
	zmq "github.com/pebbe/zmq3"
)

type ZmqClientEvents struct {
	OnMethod    chan ZMQMessage
	OnClose     chan (bool)
	SendMessage chan ZMQMessage
}

type ZMQMessage struct {
	Method string
	Params []byte // TODO: Can we change to interface{} and pack here?
}

// Creates a Zmq client, runs its goroutine, and returns the channels on
// which you should communicate
func NewZmqClient(addr string) ZmqClientEvents {
	events := ZmqClientEvents{make(chan ZMQMessage), make(chan bool), make(chan ZMQMessage)}

	go RunZmqClient(addr, events)

	return events
}

func RunZmqClient(addr string, events ZmqClientEvents) {
	sock, _ := zmq.NewSocket(zmq.PAIR)

	// Set a hwm of 1 - it's pointless to buffer requests for stats, and this
	// means that we get an error when we try to send if the client has gone
	sock.SetSndhwm(1)

	// Removing because it causes crash... (due to read goroutine)
	// defer sock.Close()

	sock.Connect(addr)

	// These are used so that we can stop the reading goroutine - hacky...
	sock.SetRcvtimeo(61e9) // 61s
	closed := false

	// TODO: This isn't correct because sock not threadsafe...
	go func() {
		defer func() {
			debug.Printf("[zmqclient] Closing socket")
			sock.Close()
		}()

		for {
			parts, err := sock.RecvMessageBytes(0)

			if closed {
				return
			}

			if err != nil {
				debug.Printf("[zmqclient] Recv err: %v", err)
			} else {
				s := string(parts[0])

				if s == "ping" {
					debug.Print("[zmqclient] Received ping, sending pong")
					events.SendMessage <- ZMQMessage{"pong", []byte("")}
				} else if s == "pong" {
					debug.Print("[zmqclient] Received pong")
					// Do nothing
				} else {
					if len(parts) != 2 {
						info.Printf("[zmqclient] Unepected message %v with %v parts", s, len(parts))
					} else {
						debug.Printf("[zmqclient] Received message %s", s)
						events.OnMethod <- ZMQMessage{s, parts[1]}
					}
				}
			}
		}
	}()

	defer func() {
		closed = true
		events.OnClose <- true
	}()

	for {
		msg := <-events.SendMessage

		// Note: I considered using sock.SendMessage but unfortunately that
		// doesn't support arbitrary flags, and it basically does this anway.

		// Set DONTWAIT so that Send doesn't block when HWM is reached

		// In case of an error and return, an OnClose will be sent by defer

		if _, err := sock.Send(msg.Method, zmq.DONTWAIT|zmq.SNDMORE); err != nil {
			info.Printf("[zmqclient] Closing %v after err: %v", addr, err)
			return
		}
		if _, err := sock.SendBytes(msg.Params, zmq.DONTWAIT); err != nil {
			info.Printf("[zmqclient] Closing %v after err: %v", addr, err)
			return
		}
	}
}
