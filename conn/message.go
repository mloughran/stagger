package conn

type Message struct {
	Method string
	Params []byte // Contains a MsgPack object
}
