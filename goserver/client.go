package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
)

import msgpack "github.com/ugorji/go-msgpack"

type StatsEnvelope struct {
	Method    string
	Timestamp int64
}

type ProtStat struct {
	N string
	T string
	V float64
	D []float64
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

func RunClient(reg Registration, info ClientRef, stats_channels StatsChannels, complete chan (CompleteMessage), send_gone chan (int)) {
	name := fmt.Sprintf("[client:%v-%v]", info.Id, reg.Name)

	log.Print(name, "Connecting to ", reg.Address)
	events := NewZmqClient(reg.Address)

	for {
		select {
		case ts := <-info.RequestStats:
			stats_req := StatsRequest{"report_all", ts}
			var b bytes.Buffer
			encoder := msgpack.NewEncoder(&b)
			encoder.Encode(stats_req)

			events.SendMessage <- b.Bytes()
		case multipart := <-events.OnMessage:
			envelope := decodeEnv(multipart.Envelope)
			ts := envelope.Timestamp
			log.Printf("%v Receiving for ts:%v [start]", name, ts)

			// Handle message parts in a goroutine
			go func() {
				for {
					select {
					case part := <-multipart.OnPart:
						stat := decodeStat(part)
						id := StatIdentifier{ts, stat.N}
						switch stat.T {
						case "c":
							stats_channels.CounterStats <- CounterStat{&id, stat.V}
						case "v":
							stats_channels.ValueStats <- ValueStat{&id, stat.V}
						case "vd":
							stats_channels.DistStats <- DistStat{&id, stat.D}
						default:
							log.Print(name, "Invalid type ", stat.T)
						}

					case <-multipart.OnEnd:
						log.Printf("%v Receiving for ts:%v [end]", name, ts)
						complete <- CompleteMessage{info.Id, ts}
						return
					}
				}
			}()
		case <-events.OnClose:
			log.Print(name, "Connection to ", reg.Address, " closed")
			send_gone <- info.Id
			return
		}
	}
}
