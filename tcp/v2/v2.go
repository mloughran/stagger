package v2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/pusher/stagger/conn"
	"io"
)

type Encoding struct{}

type method byte

const (
	PAIR_PING        = method(0x28) // Both
	PAIR_PONG        = method(0x29) // Both
	REPORT_ALL       = method(0x30) // Server -> Client
	REGISTER_PROCESS = method(0x41) // Client -> Server
	STATS_PARTIAL    = method(0x42) // Client -> Server
	STATS_COMPLETE   = method(0x43) // Client -> Server
)

const (
	MAGIC_BYTE1   = 0x83 // 'S'
	MAGIC_BYTE2   = 0x84 // 'T'
	MAGIC_VERSION = 0x00
)

var (
	MAGIC_HEADER = []byte{MAGIC_BYTE1, MAGIC_BYTE2, MAGIC_VERSION}
	method2str   map[method]string
	str2method   map[string]method
)

func init() {
	x := []struct {
		m method
		s string
	}{
		{PAIR_PING, "pair:ping"},
		{PAIR_PONG, "pair:pong"},
		{REPORT_ALL, "report_all"},
		{REGISTER_PROCESS, "register_process"},
		{STATS_PARTIAL, "stats_partial"},
		{STATS_COMPLETE, "stats_complete"},
	}

	method2str = make(map[method]string)
	str2method = make(map[string]method)
	for _, kv := range x {
		method2str[kv.m] = kv.s
		str2method[kv.s] = kv.m
	}
}

func (p Encoding) ReadMessage(r io.Reader) (msg conn.Message, err error) {
	head := make([]byte, 8)
	if _, err = r.Read(head); err != nil {
		return
	}

	if bytes.Compare(head[0:3], MAGIC_HEADER) != 0 {
		err = fmt.Errorf("Expected %v == %v", head[0:3], MAGIC_HEADER)
		return
	}

	m, ok := method2str[method(head[3])]
	if !ok {
		err = fmt.Errorf("Unknown method %v", head[3])
		return
	}
	msg = conn.Message{m, nil}

	bodySize := binary.BigEndian.Uint32(head[4:])
	msg.Params = make([]byte, bodySize)
	_, err = r.Read(msg.Params)

	return
}

func (p Encoding) WriteMessage(w io.Writer, msg conn.Message) (err error) {
	head := make([]byte, 8)
	head[0] = MAGIC_BYTE1
	head[1] = MAGIC_BYTE2
	head[2] = MAGIC_VERSION
	if method, ok := str2method[msg.Method]; ok {
		head[3] = byte(method)
	} else {
		return fmt.Errorf("No mapping for method %s", msg.Method)
	}

	binary.BigEndian.PutUint32(head[4:], uint32(len(msg.Params)))
	if _, err = w.Write(head); err != nil {
		return
	}

	_, err = w.Write(msg.Params)

	return
}

func (p Encoding) String() string {
	return "V2"
}
