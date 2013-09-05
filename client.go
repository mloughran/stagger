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

type message struct {
	Method string
	Params map[string]interface{}
}

type Client struct {
	Id    int
	addr  string
	name  string
	sendc chan (message)
}

func NewClient(id int, addr string, meta string) *Client {
	name := fmt.Sprintf("[client:%v-%v]", id, meta)
	sendc := make(chan message)
	return &Client{
		id,
		addr,
		name,
		sendc,
	}
}

func (c *Client) RequestStats(ts int64) {
	// TODO: Make Timestamp lowercase
	c.sendc <- message{"report_all", map[string]interface{}{"Timestamp": ts}}
}

func (c *Client) Shutdown() {
	c.sendc <- message{Method: "pair:shutdown"}
}

func (c *Client) Run(statsc chan<- (*Stats), complete chan<- (CompleteMessage), send_gone chan<- (int)) {
	debug.Print(c.name, "Connecting to ", c.addr)
	events := NewZmqClient(c.addr)

	handleStats := func(data []byte) (ts int64, err error) {
		var stats Stats
		if err = unmarshal(data, &stats); err == nil {
			ts = stats.Timestamp
			statsc <- &stats
		}
		return
	}

	var ts int64
	var err error

	for {
		select {
		case message := <-c.sendc:
			b, _ := msgpack.Marshal(message.Params)
			events.SendMessage <- ZMQMessage{message.Method, b}
		case m := <-events.OnMethod:
			switch m.Method {
			case "stats_partial":
				if _, err = handleStats(m.Params); err != nil {
					info.Printf("Error decoding stats_partial: %v", err)
				}
			case "stats_complete":
				if ts, err = handleStats(m.Params); err != nil {
					info.Printf("Error decoding stats_complete: %v", err)
				} else {
					complete <- CompleteMessage{c.Id, ts}
				}
			default:
				info.Printf("Received unknown command %v", m.Method)
			}
		case <-events.OnClose:
			debug.Print(c.name, "Connection to ", c.addr, " closed")
			send_gone <- c.Id
			return
		}
	}
}
