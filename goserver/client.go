package main

import (
	"bytes"
	"fmt"
	msgpack "github.com/ugorji/go-msgpack"
	"strings"
)

type StatsEnvelope struct {
	Method    string
	Timestamp int64
}

type ProtStat struct {
	N string
	T string
	V float64
	D [5]float64
}

type StatsRequest struct {
	Timestamp int64
}

type Stats struct {
	Timestamp int64
	Values    []StatValue
	Counts    []StatCount
	Dists     []StatDist
}

// TODO: Needs weight
type StatValue struct {
	Name  string
	Value float64
}

type StatCount struct {
	Name  string
	Count float64
}

type StatDist struct {
	Name string
	Dist [5]float64
}

func decodeStats(packed []byte) Stats {
	buf := strings.NewReader(string(packed))
	var stats Stats
	dec := msgpack.NewDecoder(buf, nil)
	if err := dec.Decode(&stats); err != nil {
		info.Print("Error decoding stats", err)
		// return ?
	}
	return stats
}

// TODO: Return error if fails
func decodeEnv(packed string) StatsEnvelope {
	buf := strings.NewReader(packed)
	var envelope StatsEnvelope
	dec := msgpack.NewDecoder(buf, nil)
	if err := dec.Decode(&envelope); err != nil {
		info.Print(err)
		// return
	}
	return envelope
}

func decodeStat(packed string) ProtStat {
	buf := strings.NewReader(packed)
	var stat ProtStat
	dec := msgpack.NewDecoder(buf, nil)
	if err := dec.Decode(&stat); err != nil {
		info.Print(err)
		// return
	}
	return stat
}

func RunClient(reg Registration, c ClientRef, stats_channels StatsChannels, complete chan (CompleteMessage), send_gone chan (int)) {
	name := fmt.Sprintf("[client:%v-%v]", c.Id, reg.Metadata)

	debug.Print(name, "Connecting to ", reg.Address)
	events := NewZmqClient(reg.Address)

	for {
		select {
		case ts := <-c.RequestStats:
			stats_req := StatsRequest{ts}
			var b bytes.Buffer
			encoder := msgpack.NewEncoder(&b)
			encoder.Encode(stats_req)

			events.SendMessage <- ZMQMessage{"report_all", b.Bytes()}
		case multipart := <-events.OnMessage:
			envelope := decodeEnv(multipart.Envelope)
			ts := envelope.Timestamp
			debug.Printf("%v Receiving for ts:%v [start]", name, ts)

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
							info.Print(name, "Invalid type ", stat.T)
						}

					case <-multipart.OnEnd:
						debug.Printf("%v Receiving for ts:%v [end]", name, ts)
						complete <- CompleteMessage{c.Id, ts}
						return
					}
				}
			}()
		case m := <-events.OnMethod:
			switch m.Method {
			case "stats_complete":
				stats := decodeStats(m.Params)
				stats_channels.Stats <- stats
				complete <- CompleteMessage{c.Id, stats.Timestamp}
			default:
				info.Printf("Received unknown command %v", m.Method)
			}
		case <-events.OnClose:
			debug.Print(name, "Connection to ", reg.Address, " closed")
			send_gone <- c.Id
			return
		}
	}
}
