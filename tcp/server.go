package tcp

import (
	"github.com/pusher/stagger/conn"
	"net"
	"net/url"
)

type Server struct {
	tcp_net   string
	tcp_laddr string
	conn.ClientManager
	sigShutdown chan bool
	didShutdown chan bool
	encoding    Encoding
}

func NewServer(tcp_addr string, d conn.ClientManager, e Encoding) (*Server, error) {
	url_, err := url.Parse(tcp_addr)
	if err != nil {
		return nil, err
	}
	net := url_.Scheme
	laddr := url_.Host

	return &Server{net, laddr, d, make(chan bool), make(chan bool), e}, nil
}

func (self *Server) Run() {
	conns := make(chan *Conn)

	l, err := net.Listen(self.tcp_net, self.tcp_laddr)
	if err != nil {
		return
	}
	defer l.Close()

	go func() {
		var (
			err  error
			conn net.Conn
		)
		for {
			conn, err = l.Accept()
			if err != nil {
				break
			}
			info.Printf("[tcp-server] new conn")
			conns <- NewConn(conn, self.encoding)
		}
		self.didShutdown <- true
	}()

	for {
		select {
		case c := <-conns:
			go c.Run()
			self.NewClient(c)
		case <-self.sigShutdown:
			info.Printf("[tcp-server] Shutting down listener")
			return
		}
	}
}

// Shutdown is a blocking call which
func (self *Server) Shutdown() {
	debug.Print("[tcp-server] willClose")
	self.sigShutdown <- true
	<-self.didShutdown
	debug.Print("[tcp-server] didClose")
}
