// Exposes a registration channel, creates client objects, keeps track of all existing clients, dedupes existing registrations

package pair

import (
	"time"
)

type Server struct {
	reg_addr string
	ServerDelegate
	sigShutdown chan bool
	didShutdown chan bool
}

type Pairable interface {
	Run(gone chan<- int)
	Send(m string, p map[string]interface{})
}

type ServerDelegate interface {
	AddClient(interface{})
	RemoveClient(interface{})
	NewClient(id int, pc *Conn) Pairable
}

func NewServer(reg_addr string, d ServerDelegate) *Server {
	return &Server{reg_addr, d, make(chan bool), make(chan bool)}
}

func (self *Server) Run() {
	registration := NewRegistration(self.reg_addr)
	go registration.Run()

	idIncr := 0
	clients := make(map[int]Pairable)

	// Clients send a message on this channel when they go away
	clientDidClose := make(chan int)

	for {
		select {
		case addr := <-registration.Registrations:
			pc := NewConn()
			pc.ShouldConnect(addr)
			go pc.Run()

			idIncr += 1
			client := self.NewClient(idIncr, pc)
			go client.Run(clientDidClose)

			clients[idIncr] = client
			self.AddClient(client)

		case id := <-clientDidClose:
			self.RemoveClient(clients[id])
			delete(clients, id)

		case <-self.sigShutdown:
			info.Printf("[pair-server] Shutting down registration")
			registration.Shutdown()

			info.Printf("[pair-server] Sending shutdown message to all clients")
			for _, client := range clients {
				client.Send("pair:shutdown", nil)
			}
			// Needs time to send shutdown - 1ms was brittle, 50ms is generous
			// TODO: Find a way to do this which doesn't rely on timing
			time.Sleep(50 * time.Millisecond)
			self.didShutdown <- true
		}
	}
}

// Shutdown is a blocking call which
// * closes the registration channel cleanly,
// * then signals all client connections to shutdown
func (s *Server) Shutdown() {
	debug.Print("[pair-server] willClose")
	s.sigShutdown <- true
	<-s.didShutdown
	debug.Print("[pair-server] didClose")
}
