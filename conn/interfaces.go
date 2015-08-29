package conn

type (
	Connection interface {
		Send(method string, params []byte)
		OnMethod() <-chan Message
		OnClose() <-chan bool
		Shutdown()
		String() string
	}

	Client interface {
		Id() int64
		RequestStats(ts int64)
		Run(clientDidClose chan<- Client)
		Send(m string, p map[string]interface{})
		Shutdown()
	}

	ClientManager interface {
		NewClient(Connection)
	}
)
