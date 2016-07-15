package tcp

import (
	"fmt"
	"github.com/pusher/stagger/conn"
	"io"
	"net"
	"time"
)

var QUEUE_FULL = fmt.Errorf("Send queue full")

type Message struct {
	conn.Message
	Err error
}

type Conn struct {
	c           conn.Conn
	encoding    conn.Encoding
	onMethod    chan conn.Message
	onClose     chan bool
	sendMessage chan conn.Message
	interval    time.Duration
}

// NewConn creates a Connection. You must select on OnMethod and OnClose
func NewConn(c net.Conn, e conn.Encoding, interval time.Duration) *Conn {
	return &Conn{conn.NewConn(c, e), e, make(chan conn.Message, 1), make(chan bool), make(chan conn.Message, 1), interval}
}

func (c *Conn) String() string {
	return fmt.Sprintf("mode=tcp encoding=%s", c.encoding)
}

// Send sends a message to the Connection
func (c *Conn) Send(msg conn.Message) error {
	select {
	case c.sendMessage <- msg:
		return nil
	default:
		return QUEUE_FULL
	}
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
			// The next message should arrive in max 2 the update interval
			c.c.SetReadDeadline(
				time.Now().Add(2 * c.interval),
			)

			msg := Message{}
			msg.Message, msg.Err = c.c.ReadMessage()
			recvMessage <- msg
			if msg.Err != nil {
				break
			}
		}
	}()

	// send method for use by this goroutine
	send := func(msg conn.Message) (err error) {
		debug.Printf("Sending %v", msg)
		err = c.c.WriteMessage(msg)
		return
	}

	for {
		select {
		case msg := <-c.sendMessage:
			// Writes should be fast
			c.c.SetWriteDeadline(
				time.Now().Add(1 * time.Second),
			)
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

			switch msg.Message.(type) {
			case *conn.PairPing:
				debug.Print("[tcp-conn] Received ping, sending pong")
				send(conn.PairPong{})
			case *conn.PairPong:
				debug.Print("[tcp-conn] Received pong")
				// Do nothing
			default:
				debug.Printf("[tcp-conn] Received message %v", msg)
				c.onMethod <- msg.Message
			}
		}
	}
}
