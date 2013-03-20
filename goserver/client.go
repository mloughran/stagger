package main

import "log"
import "strings"

import msgpack "github.com/ugorji/go-msgpack"

type StatsEnvelope struct {
	Method string
}

// TODO: Return error if fails
func decodeEnv(packed string) StatsEnvelope {
	buf := strings.NewReader(packed)
	var envelope StatsEnvelope
	dec := msgpack.NewDecoder(buf, nil)
	if err := dec.Decode(&envelope); err != nil {
		log.Print(err)
		// return
	}
	return envelope
}

func RunClient(info ClientRef) {
	events := NewZmqClient(info.Address)

	for {
		select {
		case <-info.Mailbox:
			log.Print("[client] Requesting stats from ", info.Address)
			events.SendMessage <- "Stats please!"
		case multipart := <-events.OnMessage:
			log.Print("[client] Received stats")

			envelope := decodeEnv(multipart.Envelope)
			log.Print("envelope: ", envelope)

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
