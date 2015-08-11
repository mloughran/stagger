package tcp

import (
	"fmt"
	"github.com/pusher/stagger/conn"
	"io"
	"net"
	"time"
)

const HEAD = "%d,%d."
const SEP = ":"

type Message struct {
	conn.Message
	Err error
}

type Conn struct {
	c           net.Conn
	onMethod    chan conn.Message
	onClose     chan bool
	sendMessage chan Message
}

// NewConn creates a Connection. You must select on OnMethod and OnClose
func NewConn(c net.Conn) *Conn {
	return &Conn{c, make(chan conn.Message), make(chan bool), make(chan Message, 1)}
}

// Send sends a message to the Connection
func (c *Conn) Send(method string, params []byte) {
	c.sendMessage <- Message{conn.Message{method, params}, nil}
}

func (c *Conn) OnMethod() <-chan conn.Message {
	return c.onMethod
}

func (c *Conn) OnClose() <-chan bool {
	return c.onClose
}

func (c *Conn) Shutdown() {
	c.c.Close()
}

func (c *Conn) Run() {
	defer func() {
		c.c.Close()
		c.onClose <- true
	}()

	recvMessage := make(chan Message, 2)

	// Goroutine handles reading from tcp
	go func() {
		var (
			msg      Message
			headLen  uint64
			paramLen uint64
		)
		sep := make([]byte, 1)
		for {
			msg = Message{}

			if _, msg.Err = fmt.Fscanf(c.c, HEAD, &headLen, &paramLen); msg.Err != nil {
				break
			}

			method := make([]byte, headLen)
			if _, msg.Err = c.c.Read(method); msg.Err != nil {
				break
			}
			msg.Method = string(method)

			if _, msg.Err = c.c.Read(sep); msg.Err != nil {
				break
			}
			if string(sep) != SEP {
				msg.Err = fmt.Errorf("Invalid separator, should be %s but was %s", SEP, string(sep))
				break
			}

			msg.Params = make([]byte, paramLen)
			if _, msg.Err = c.c.Read(msg.Params); msg.Err != nil {
				break
			}

			recvMessage <- msg
		}
		if msg.Err != nil {
			recvMessage <- msg
		}
	}()

	// send method for use by this goroutine
	send := func(method string, params []byte) (err error) {
		debug.Printf("Sending %s %s", method, params)
		if _, err = fmt.Fprintf(c.c, HEAD, len(method), len(params)); err != nil {
			return
		}

		if _, err = c.c.Write([]byte(method)); err != nil {
			return
		}

		if _, err = c.c.Write([]byte(SEP)); err != nil {
			return
		}

		if _, err = c.c.Write(params); err != nil {
			return
		}

		return
	}

	var msg Message
	for {
		select {
		case msg = <-c.sendMessage:
			if err := send(msg.Method, msg.Params); err != nil {
				info.Printf("[tcp-conn] Closing %v after err: %v", c.c.RemoteAddr(), err)
				return
			}
		case msg = <-recvMessage:
			if msg.Err != nil {
				if msg.Err != io.EOF {
					info.Printf("[tcp-conn] Closing(2) %v after err: %v", c.c.RemoteAddr(), msg.Err)
				}
				return
			}
			// The next message should arrive in max 61 seconds
			c.c.SetDeadline(time.Now().Add(61 * time.Second))

			switch msg.Method {
			case "pair:ping":
				debug.Print("[tcp-conn] Received ping, sending pong")
				send("pair:pong", []byte{})
			case "pair:pong":
				debug.Print("[tcp-conn] Received pong")
				// Do nothing
			default:
				debug.Printf("[tcp-conn] Received message %v", msg.Method)
				c.onMethod <- msg.Message
			}
		case <-time.After(30 * time.Second):
			if err := send("pair:ping", []byte{}); err != nil {
				info.Printf("ping error", err)
			}
		}
	}
}
