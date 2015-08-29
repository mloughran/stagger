package v1

import (
	"fmt"
	"github.com/pusher/stagger/conn"
	"io"
)

// Protocol v2
type Encoding struct{}

const HEAD = "%d,%d."
const SEP = ":"

func (e Encoding) ReadMessage(r io.Reader) (msg conn.Message, err error) {
	var (
		headLen  uint64
		paramLen uint64
	)
	msg = conn.Message{}
	sep := make([]byte, 1)

	if _, err = fmt.Fscanf(r, HEAD, &headLen, &paramLen); err != nil {
		return
	}

	method := make([]byte, headLen)
	if _, err = r.Read(method); err != nil {
		return
	}
	msg.Method = string(method)

	if _, err = r.Read(sep); err != nil {
		return
	}
	if string(sep) != SEP {
		err = fmt.Errorf("Invalid separator, should be %s but was %s", SEP, string(sep))
		return
	}

	msg.Params = make([]byte, paramLen)
	if _, err = r.Read(msg.Params); err != nil {
		return
	}

	return
}

func (e Encoding) WriteMessage(w io.Writer, msg conn.Message) (err error) {
	if _, err = fmt.Fprintf(w, HEAD, len(msg.Method), len(msg.Params)); err != nil {
		return
	}

	if _, err = w.Write([]byte(msg.Method)); err != nil {
		return
	}

	if _, err = w.Write([]byte(SEP)); err != nil {
		return
	}

	if _, err = w.Write(msg.Params); err != nil {
		return
	}

	return
}
