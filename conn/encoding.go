package conn

import "io"

type Encoding interface {
	ReadMessage(io.Reader) (Message, error)
	WriteMessage(io.Writer, Message) error
	String() string
}
