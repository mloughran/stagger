// Exposes a registration channel, creates client objects, keeps track of all existing clients, dedupes existing registrations

package pair

import (
	"github.com/pusher/stagger/conn"
)

type Server struct {
	reg_addr string
	conn.ClientManager
	sigShutdown chan bool
}

func NewServer(reg_addr string, d conn.ClientManager) *Server {
	return &Server{reg_addr, d, make(chan bool)}
}

func (self *Server) Run() {
	registration := NewRegistration(self.reg_addr)
	go registration.Run()
	defer func() {
		registration.Shutdown()
	}()

	for {
		select {
		case addr := <-registration.Registrations:
			pc := NewConn()
			pc.ShouldConnect(addr)
			go pc.Run()

			self.NewClient(pc)

		case <-self.sigShutdown:
			info.Printf("[pair-server] Shutting down registration")
			return
		}
	}
}

// Shutdown is a blocking call which
// * closes the registration channel cleanly,
// * then signals all client connections to shutdown
func (s *Server) Shutdown() {
	debug.Print("[pair-server] willClose")
	s.sigShutdown <- true
	debug.Print("[pair-server] didClose")
}
