// Exposes a registration channel, creates client objects, keeps track of all existing clients, dedupes existing registrations

package pair

type Server struct {
	reg_addr    string
	on_shutdown chan bool
	ServerDelegate
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

func NewServer(reg_addr string, on_shutdown chan bool, d ServerDelegate) *Server {
	return &Server{reg_addr, on_shutdown, d}
}

func (self *Server) Run() {
	registraton := NewRegistration(self.reg_addr)
	go registraton.Run()

	idIncr := 0
	clients := make(map[int]Pairable)

	// Clients send a message on this channel when they go away
	on_client_gone := make(chan int)

	for {
		select {
		case addr := <-registraton.Registrations:
			debug.Print("Connecting to ", addr)
			pc := NewConn()
			pc.Connect(addr) // TODO: Error checking
			go pc.Run()

			idIncr += 1
			client := self.NewClient(idIncr, pc)
			go client.Run(on_client_gone)

			clients[idIncr] = client
			self.AddClient(client)

		case id := <-on_client_gone:
			self.RemoveClient(clients[id])
			delete(clients, id)

		case <-self.on_shutdown:
			info.Printf("[cm] Sending shutdown message to all clients")
			for _, client := range clients {
				client.Send("pair:shutdown", nil)
			}
		}
	}
}
