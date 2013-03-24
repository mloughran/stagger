package main

import "log"
import "time"

type ClientRef struct {
	Address      string
	Mailbox      chan (string)
	RequestStats chan (int64)
}

func StartClientManager(registration chan (string), stat_chan chan (Stat)) {
	clients := make([]ClientRef, 0)

	heartbeat := time.Tick(5 * time.Second)

	for {
		select {
		case client_address := <-registration:
			client := ClientRef{client_address, make(chan string), make(chan int64)}
			go RunClient(client, stat_chan)
			log.Print("clientmanager", client.Address)
			clients = append(clients, client)

			log.Print("Managing client: ", len(clients))
		case ts := <-heartbeat:
			log.Print("[cm] Sending request for stats")

			// Send stats request to each client
			unix := ts.Unix()
			for _, client := range clients {
				client.RequestStats <- unix
			}
		}
	}
}
