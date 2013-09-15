// Exposes a registration channel, creates client objects, keeps track of all existing clients, dedupes existing registrations

package pair

type PairServer struct {
	reg_addr    string
	on_shutdown chan bool
}

type Pairable interface {
	Run(gone chan<- int)
	Send(m string, p map[string]interface{})
}

type PairServerDelegate interface {
	AddClient(interface{})
	RemoveClient(interface{})
}

func NewPairServer(reg_addr string, on_shutdown chan bool) *PairServer {
	return &PairServer{reg_addr, on_shutdown}
}

type clientGen func(id int, pc *PairConn) Pairable

func (self *PairServer) Run(d PairServerDelegate, g clientGen) {
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
			pc := NewPairConn()
			pc.Connect(addr) // TODO: Error checking
			go pc.Run()

			idIncr += 1
			client := g(idIncr, pc)
			go client.Run(on_client_gone)

			clients[idIncr] = client
			d.AddClient(client)

		case id := <-on_client_gone:
			d.RemoveClient(clients[id])
			delete(clients, id)

		case <-self.on_shutdown:
			info.Printf("[cm] Sending shutdown message to all clients")
			for _, client := range clients {
				client.Send("pair:shutdown", nil)
			}
		}
	}
}
