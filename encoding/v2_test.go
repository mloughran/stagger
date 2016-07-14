package encoding

import (
	"bytes"
	"github.com/pusher/stagger/conn"
	"testing"
)

func TestMirror(t *testing.T) {
	b := new(bytes.Buffer)
	when := int64(1337)
	err := (Encoding{}).WriteMessage(b, conn.ReportAll{when})
	if err != nil {
		t.Error("encode error", err)
	}

	msg, err := (Encoding{}).ReadMessage(b)
	if err != nil {
		t.Error("decode error", err)
	}

	switch m := msg.(type) {
	case *conn.ReportAll:
		if m.Timestamp != when {
			t.Error("bad timestamp", m.Timestamp)
		}
	default:
		t.Error("bad method", msg)
	}
}
