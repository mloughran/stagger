package encoding

import (
	"bytes"
	"github.com/pusher/stagger/conn"
	"testing"
)

func TestMirror(t *testing.T) {
	b := new(bytes.Buffer)
	err := (Encoding{}).WriteMessage(b, conn.Message{"report_all", []byte{10}})
	if err != nil {
		t.Error("encode error", err)
	}

	msg, err := (Encoding{}).ReadMessage(b)
	if err != nil {
		t.Error("decode error", err)
	}

	if msg.Method != "report_all" {
		t.Error("bad method", msg.Method)
	}

	if msg.Method != "report_all" {
		t.Error("bad method", msg.Method)
	}

	if bytes.Compare(msg.Params, []byte{10}) != 0 {
		t.Error("bad params", msg.Params)
	}
}
