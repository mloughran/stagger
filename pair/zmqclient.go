package pair

import (
	zmq "github.com/pebbe/zmq3"
)

type PairConn struct {
	OnMethod    chan zmqMessage
	OnClose     chan (bool)
	sendMessage chan zmqMessage
	pair        *zmq.Socket
	addr        string
}

type zmqMessage struct {
	Method string
	Params []byte // TODO: Can we change to interface{} and pack here?
}

// Creates a Zmq client, runs its goroutine, and returns the channels on
// which you should communicate
func NewPairConn() *PairConn {
	sock, _ := zmq.NewSocket(zmq.PAIR)
	return &PairConn{make(chan zmqMessage), make(chan bool), make(chan zmqMessage), sock, ""}
}

func (z *PairConn) Connect(addr string) {
	z.addr = addr
	z.pair.Connect(addr)
}

func (z *PairConn) Send(method string, params []byte) {
	z.sendMessage <- zmqMessage{method, params}
}

func (z *PairConn) Run() {
	sock := z.pair

	// Set a hwm of 1 - it's pointless to buffer requests for stats, and this
	// means that we get an error when we try to send if the client has gone
	sock.SetSndhwm(1)

	// Removing because it causes crash... (due to read goroutine)
	// defer sock.Close()

	// These are used so that we can stop the reading goroutine - hacky...
	sock.SetRcvtimeo(61e9) // 61s
	closed := false

	// Used to signal shutdown to the main goroutine
	sig_shutdown := make(chan bool)

	// TODO: This isn't correct because sock not threadsafe...
	go func() {
		defer func() {
			debug.Printf("[zmqclient] Closing socket")
			sock.Close()
		}()

		for {
			parts, err := sock.RecvMessageBytes(0)

			if closed {
				// Returns from this (zmq reading) goroutine
				return
			}

			if err != nil {
				debug.Printf("[zmqclient] Recv err: %v", err)
			} else {
				s := string(parts[0])

				// Non-prefixed ping & ping are deprecated. TODO: remove
				if s == "ping" || s == "pair:ping" {
					debug.Print("[zmqclient] Received ping, sending pong")
					z.Send("pair:pong", []byte(""))
				} else if s == "pong" || s == "pair:pong" {
					debug.Print("[zmqclient] Received pong")
					// Do nothing
				} else if s == "pair:shutdown" {
					debug.Print("[zmqclient] Received shutdown")
					sig_shutdown <- true
				} else {
					if len(parts) != 2 {
						info.Printf("[zmqclient] Unepected message %v with %v parts", s, len(parts))
					} else {
						debug.Printf("[zmqclient] Received message %s", s)
						z.OnMethod <- zmqMessage{s, parts[1]}
					}
				}
			}
		}
	}()

	defer func() {
		closed = true
		z.OnClose <- true
	}()

	for {
		select {
		case <-sig_shutdown:
			return
		case msg := <-z.sendMessage:
			// Note: I considered using sock.SendMessage but unfortunately that
			// doesn't support arbitrary flags, and it basically does this anway.

			// Set DONTWAIT so that Send doesn't block when HWM is reached

			// In case of an error and return, an OnClose will be sent by defer

			if _, err := sock.Send(msg.Method, zmq.DONTWAIT|zmq.SNDMORE); err != nil {
				info.Printf("[zmqclient] Closing %v after err: %v", z.addr, err)
				return
			}
			if _, err := sock.SendBytes(msg.Params, zmq.DONTWAIT); err != nil {
				info.Printf("[zmqclient] Closing %v after err: %v", z.addr, err)
				return
			}
		}
	}
}
