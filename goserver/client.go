package main

import "log"

func RunClient(info ClientRef) {
	events := NewZmqClient(info.Address)

	for {
		select {
		case <-info.Mailbox:
			log.Print("[client] Requesting stats from ", info.Address)
			events.SendMessage <- "Stats please!"
		case multipart := <-events.OnMessage:
			log.Print("[client] Received stats")
			log.Print("envelope: ", multipart.Envelope)

			// Handle message parts in a goroutine
			go func() {
				for {
					select {
					case part := <-multipart.OnPart:
						log.Print("[client] Stats part ", part)
					case <-multipart.OnEnd:
						log.Print("[client] End of stats stream")
						return
					}
				}
			}()
		case <-events.OnClose:
			log.Print("[client] Connection to ", info.Address, " closed")
			return
		}
	}
}
