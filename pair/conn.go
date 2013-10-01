package pair

import (
	zmq "github.com/pebbe/zmq3"
)

type Conn struct {
	OnMethod      chan zmqMessage
	OnClose       chan bool
	sendMessage   chan zmqMessage
	addr          string
	shouldConnect bool
}

type zmqMessage struct {
	Method string
	Params []byte
}

// NewConn creates a Connection. You must select on OnMethod and OnClose
func NewConn() *Conn {
	return &Conn{make(chan zmqMessage), make(chan bool), make(chan zmqMessage), "", false}
}

// ShouldConnect notifies the Connection that it should Connect when Run called
func (c *Conn) ShouldConnect(addr string) {
	c.addr = addr
	c.shouldConnect = true
}

// Send sends a message to the Connection
func (c *Conn) Send(method string, params []byte) {
	c.sendMessage <- zmqMessage{method, params}
}

func (c *Conn) Run() {
	recvMessage := make(chan ([][]byte))
	shouldClose := false

	pair, _ := zmq.NewSocket(zmq.PAIR)

	// Set a HWM to force an error when sending, thereby detecting that the
	// client has gone away
	if err := pair.SetSndhwm(1); err != nil {
		info.Printf("[pair] Error in SetSndhwm: %v", err)
		return
	}

	// Timeout reads after 61s so that this goroutine can close
	if err := pair.SetRcvtimeo(9e9); err != nil {
		info.Printf("[pair] Error in SetRcvtimeo: %v", err)
		return
	}

	if c.shouldConnect {
		if err := pair.Connect(c.addr); err != nil {
			info.Printf("[pair] Error connecting to client: %v", err)
			return
		}
	}

	// Goroutine handles reading from zmq
	go func() {
		for {
			parts, err := pair.RecvMessageBytes(0)

			if shouldClose {
				debug.Printf("[pair] Closing pair socket")
				pair.Close()
				return
			}

			if err == nil {
				recvMessage <- parts
			}
		}
	}()

	defer func() {
		shouldClose = true
		c.OnClose <- true
	}()

	for {
		select {
		case msg := <-c.sendMessage:
			// Note: I considered using the cleaner pair.SendMessage but
			// unfortunately that doesn't currently support arbitrary flags
			// Using DONTWAIT to avoid Send blocking when HWM is reached
			if _, err := pair.Send(msg.Method, zmq.DONTWAIT|zmq.SNDMORE); err != nil {
				info.Printf("[pair] Closing %v after err: %v", c.addr, err)
				return
			}
			if _, err := pair.SendBytes(msg.Params, zmq.DONTWAIT); err != nil {
				info.Printf("[pair] Closing %v after err: %v", c.addr, err)
				return
			}
		case parts := <-recvMessage:
			s := string(parts[0])

			if s == "pair:ping" {
				debug.Print("[pair] Received ping, sending pong")
				c.Send("pair:pong", []byte(""))
			} else if s == "pair:pong" {
				debug.Print("[pair] Received pong")
				// Do nothing
			} else if s == "pair:shutdown" {
				info.Print("[pair] Remote peer sent shutdown message")
				return
			} else {
				if len(parts) != 2 {
					info.Printf("[pair] Unexpected message %v with %v parts", s, len(parts))
				} else {
					debug.Printf("[pair] Received message %s", s)
					c.OnMethod <- zmqMessage{s, parts[1]}
				}
			}
		}
	}
}
