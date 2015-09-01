package tcp

import (
	"github.com/pusher/stagger/conn"
	"io"
)

type Encoding interface {
	ReadMessage(io.Reader) (conn.Message, error)
	WriteMessage(io.Writer, conn.Message) error
	String() string
}
