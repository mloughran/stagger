package tcp

import (
	"github.com/pusher/stagger/conn"
)

type (
	Connection interface {
		Send(message conn.Message) error
		OnMethod() <-chan conn.Message
		OnClose() <-chan bool
		Shutdown()
		String() string
	}

	Client interface {
		Id() int64
		RequestStats(ts int64)
		Send(message conn.Message)
		Shutdown()
	}

	ClientManager interface {
		NewClient(Connection)
	}
)
