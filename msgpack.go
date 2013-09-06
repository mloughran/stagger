package main

import (
	codec "github.com/ugorji/go/codec"
)

var msgpack_handle codec.MsgpackHandle

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
