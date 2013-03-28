package main

import "log"
import "strings"
import "bytes"
import "fmt"

import msgpack "github.com/ugorji/go-msgpack"

type StatsEnvelope struct {
	Method    string
	Timestamp int64
}

type ProtStat struct {
	N string
	T string
	V int
}

type StatsRequest struct {
	Method    string
	Timestamp int64
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

func RunClient(info ClientRef, stat_chan chan (Stat), complete chan (CompleteMessage)) {
	name := fmt.Sprintf("[client %v] ", info.Id)

	log.Print(name, "Connecting to ", info.Address)
	events := NewZmqClient(info.Address)

	for {
		select {
		// case <-info.Mailbox:
		// 	log.Print("[client] Requesting stats from ", info.Address)
		// 	events.SendMessage <- "Stats please!"
		case ts := <-info.RequestStats:
			log.Print(name, "Requesting stats at ", ts)

			stats_req := StatsRequest{"report_all", ts}
			var b bytes.Buffer
			encoder := msgpack.NewEncoder(&b)
			encoder.Encode(stats_req)

			events.SendMessage <- &b
		case multipart := <-events.OnMessage:
			envelope := decodeEnv(multipart.Envelope)
			ts := envelope.Timestamp
			log.Print(name, "Receiving stats at ts ", ts)

			// Handle message parts in a goroutine
			go func() {
				for {
					select {
					case part := <-multipart.OnPart:
						stat := decodeStat(part)
						stat_chan <- Stat{ts, stat.N, stat.V}
					case <-multipart.OnEnd:
						log.Print(name, "End of stats stream")
						complete <- CompleteMessage{info.Id, ts}
						return
					}
				}
			}()
		case <-events.OnClose:
			log.Print(name, "Connection to ", info.Address, " closed")
			return
		}
	}
}
