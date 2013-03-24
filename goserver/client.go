package main

import "log"
import "strings"

import msgpack "github.com/ugorji/go-msgpack"

type StatsEnvelope struct {
	Method string
}

type ProtStat struct {
	N string
	T string
	V int
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

func decodeStat(packed string) ProtStat {
	buf := strings.NewReader(packed)
	var stat ProtStat
	dec := msgpack.NewDecoder(buf, nil)
	if err := dec.Decode(&stat); err != nil {
		log.Print(err)
		// return
	}
	return stat
}

func RunClient(info ClientRef, stat_chan chan (Stat)) {
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
						stat := decodeStat(part)
						log.Print("[client] Stats part ", stat)

						// Send the data on to the stat_channel
						stat_chan <- Stat{stat.N, stat.V}
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
