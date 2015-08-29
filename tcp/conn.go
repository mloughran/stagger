package tcp

import (
	"github.com/pusher/stagger/conn"
	"io"
	"net"
	"time"
)

type Message struct {
	conn.Message
	Err error
}

type Conn struct {
	c           net.Conn
	encoding    Encoding
	onMethod    chan conn.Message
	onClose     chan bool
	sendMessage chan conn.Message
}

// NewConn creates a Connection. You must select on OnMethod and OnClose
func NewConn(c net.Conn, e Encoding) *Conn {
	return &Conn{c, e, make(chan conn.Message), make(chan bool), make(chan conn.Message, 1)}
}

// Send sends a message to the Connection
func (c *Conn) Send(method string, params []byte) {
	c.sendMessage <- conn.Message{method, params}
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
		for {
			msg := Message{}
			msg.Message, msg.Err = c.encoding.ReadMessage(c.c)
			recvMessage <- msg
			if msg.Err != nil {
				break
			}
		}
	}()

	// send method for use by this goroutine
	send := func(msg conn.Message) (err error) {
		debug.Printf("Sending %v", msg)
		err = c.encoding.WriteMessage(c.c, msg)
		return
	}

	for {
		select {
		case msg := <-c.sendMessage:
			if err := send(msg); err != nil {
				info.Printf("[tcp-conn] Closing %v after err: %v", c.c.RemoteAddr(), err)
				return
			}
		case msg := <-recvMessage:
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
				send(conn.Message{"pair:pong", []byte{}})
			case "pair:pong":
				debug.Print("[tcp-conn] Received pong")
				// Do nothing
			default:
				debug.Printf("[tcp-conn] Received message %v", msg.Method)
				c.onMethod <- msg.Message
			}
		case <-time.After(30 * time.Second):
			if err := send(conn.Message{"pair:ping", []byte{}}); err != nil {
				info.Printf("ping error", err)
			}
		}
	}
}
