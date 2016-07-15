package encoding

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/pusher/stagger/conn"
	codec "github.com/ugorji/go/codec"
	"io"
)

type Encoding struct{}

var (
	MAGIC_HEADER   = []byte{MAGIC_BYTE1, MAGIC_BYTE2, MAGIC_VERSION}
	msgpack_handle codec.MsgpackHandle
)

func (e Encoding) ReadMessage(r io.Reader) (msg conn.Message, err error) {
	head := make([]byte, 8)
	if _, err = r.Read(head); err != nil {
		return
	}

	if bytes.Compare(head[0:3], MAGIC_HEADER) != 0 {
		err = fmt.Errorf("Expected %v == %v", head[0:3], MAGIC_HEADER)
		return
	}

	bodySize := binary.BigEndian.Uint32(head[4:])
	params := make([]byte, bodySize)
	if _, err = r.Read(params); err != nil {
		return
	}

	switch head[3] {
	case PAIR_PING:
		msg = &conn.PairPing{}
	case PAIR_PONG:
		msg = &conn.PairPong{}
	case REPORT_ALL:
		var rep conn.ReportAll
		if err = unmarshal(params, &rep); err != nil {
			return
		}
		msg = &rep
	case REGISTER_PROCESS:
		var reg conn.RegisterProcess
		if err = unmarshal(params, &reg); err != nil {
			return
		}
		msg = &reg
	case STATS_PARTIAL:
		var stats conn.Stats
		if err = unmarshal(params, &stats); err != nil {
			return
		}
		msg = &conn.StatsPartial{Stats: stats}
	case STATS_COMPLETE:
		var stats conn.Stats
		if err = unmarshal(params, &stats); err != nil {
			return
		}
		msg = &conn.StatsComplete{Stats: stats}
	default:
		err = fmt.Errorf("Unknown method %v", head[3])
	}

	return
}

func (p Encoding) WriteMessage(w io.Writer, msg conn.Message) (err error) {
	head := make([]byte, 8)
	head[0] = MAGIC_BYTE1
	head[1] = MAGIC_BYTE2
	head[2] = MAGIC_VERSION

	var method byte
	var params []byte

	switch m := msg.(type) {
	case conn.PairPing:
		method = PAIR_PING
	case conn.PairPong:
		method = PAIR_PONG
	case conn.ReportAll:
		method = REPORT_ALL
		if params, err = marshal(m); err != nil {
			return
		}
	case conn.RegisterProcess:
		method = REGISTER_PROCESS
		if params, err = marshal(m); err != nil {
			return
		}
	case conn.StatsPartial:
		method = STATS_PARTIAL
		if params, err = marshal(m.Stats); err != nil {
			return
		}
	case conn.StatsComplete:
		method = STATS_COMPLETE
		if params, err = marshal(m.Stats); err != nil {
			return
		}
	default:
		return fmt.Errorf("Unknown message type %v", msg)
	}

	head[3] = method
	binary.BigEndian.PutUint32(head[4:], uint32(len(params)))
	if _, err = w.Write(head); err != nil {
		return
	}
	_, err = w.Write(params)

	return
}

func (p Encoding) String() string {
	return "V2"
}

func unmarshal(data []byte, v interface{}) (err error) {
	dec := codec.NewDecoderBytes(data, &msgpack_handle)
	err = dec.Decode(&v)
	return
}

func marshal(v interface{}) (b []byte, err error) {
	enc := codec.NewEncoderBytes(&b, &msgpack_handle)
	err = enc.Encode(v)
	return
}
