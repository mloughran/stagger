package conn

import (
	"net"
)

type Conn interface {
	net.Conn
	ReadMessage() (Message, error)
	WriteMessage(Message) error
}

func NewConn(c net.Conn, e Encoding) Conn {
	return &Wrapper{c, e}
}

type Wrapper struct {
	net.Conn
	enc Encoding
}

func (c *Wrapper) ReadMessage() (Message, error) {
	return c.enc.ReadMessage(c.Conn)
}

func (c *Wrapper) WriteMessage(m Message) error {
	return c.enc.WriteMessage(c.Conn, m)
}
