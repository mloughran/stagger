package main

import (
	"fmt"
	msgpack "github.com/ugorji/go-msgpack"
)

type StatsEnvelope struct {
	Method    string
	Timestamp int64
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

func unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v, msgpack.DefaultDecoderContainerResolver)
}

func RunClient(reg Registration, c ClientRef, statsc chan (Stats), complete chan (CompleteMessage), send_gone chan (int)) {
	name := fmt.Sprintf("[client:%v-%v]", c.Id, reg.Metadata)

	debug.Print(name, "Connecting to ", reg.Address)
	events := NewZmqClient(reg.Address)

	for {
		select {
		case message := <-c.SendMessage:
			b, _ := msgpack.Marshal(message.Params)
			events.SendMessage <- ZMQMessage{message.Method, b}
		case m := <-events.OnMethod:
			switch m.Method {
			case "stats_complete":
				var stats Stats
				if err := unmarshal(m.Params, &stats); err != nil {
					info.Printf("Error decoding stats_complete: %v", err)
				} else {
					statsc <- stats
					complete <- CompleteMessage{c.Id, stats.Timestamp}
				}
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
