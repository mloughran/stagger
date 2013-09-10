package main

import (
	"fmt"
)

type StatsEnvelope struct {
	Method    string
	Timestamp int64
}

type StatsRequest struct {
	Timestamp int64
}

type message struct {
	Method string
	Params map[string]interface{}
}

type Client struct {
	id       int
	addr     string
	name     string
	sendc    chan (message)
	statsc   chan<- (*Stats)
	complete chan<- (CompleteMessage)
}

func NewClient(id int, addr string, meta string, statsc chan<- (*Stats), complete chan<- (CompleteMessage)) *Client {
	name := fmt.Sprintf("[client:%v-%v]", id, meta)
	sendc := make(chan message)
	return &Client{
		id,
		addr,
		name,
		sendc,
		statsc,
		complete,
	}
}

func (c *Client) Id() int {
	return c.id
}

func (c *Client) Send(m string, p map[string]interface{}) {
	c.sendc <- message{m, p}
}

func (c *Client) RequestStats(ts int64) {
	// TODO: Make Timestamp lowercase
	c.Send("report_all", map[string]interface{}{"Timestamp": ts})
}

func (c *Client) Run(send_gone chan<- (int)) {
	debug.Print(c.name, "Connecting to ", c.addr)
	events := NewZmqClient(c.addr)

	handleStats := func(data []byte) (ts int64, err error) {
		var stats Stats
		if err = unmarshal(data, &stats); err == nil {
			ts = stats.Timestamp
			c.statsc <- &stats
		} else {
			info.Printf("Error decoding msgpack data: %v", data)
		}
		return
	}

	var ts int64
	var err error

	for {
		select {
		case message := <-c.sendc:
			if b, err := marshal(message.Params); err == nil {
				events.SendMessage <- ZMQMessage{message.Method, b}
			} else {
				info.Printf("Error encoding as msgpack: %v", message.Params)
			}
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
					c.complete <- CompleteMessage{c.Id(), ts}
				}
			default:
				info.Printf("Received unknown command %v", m.Method)
			}
		case <-events.OnClose:
			debug.Print(c.name, "Connection to ", c.addr, " closed")
			send_gone <- c.Id()
			return
		}
	}
}
