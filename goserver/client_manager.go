package main

import "log"
import "time"

type ClientRef struct {
	Address string
	Mailbox chan (string)
}

func StartClientManager(registration chan (string)) {
	clients := make([]ClientRef, 0)

	heartbeat := time.Tick(5 * time.Second)

	for {
		select {
		case client_address := <-registration:
			mailbox := make(chan string)
			client := ClientRef{client_address, mailbox}
			go RunClient(client)
			log.Print("clientmanager", client.Address)
			clients = append(clients, client)

			log.Print("Managing client: ", len(clients))
		case <-heartbeat:
			log.Print("[cm] Sending request for stats")

			// Send stats request to each client
			for _, client := range clients {
				client.Mailbox <- "send me more"
			}
		}
	}
}
