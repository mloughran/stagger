package main

import (
	"./pair"
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
	pc       *pair.Conn
	name     string
	sendc    chan (message)
	statsc   chan<- (*Stats)
	complete chan<- (CompleteMessage)
}

func NewClient(id int, pc *pair.Conn, meta string, statsc chan<- (*Stats), complete chan<- (CompleteMessage)) *Client {
	name := fmt.Sprintf("[client:%v-%v]", id, meta)
	sendc := make(chan message, 1)
	return &Client{
		id,
		pc,
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

func (c *Client) Run(clientDidClose chan<- int) {
	handleStats := func(data []byte) (ts int64, err error) {
		var stats Stats
		if err = unmarshal(data, &stats); err == nil {
			ts = stats.Timestamp
			c.statsc <- &stats
		} else {
			info.Printf("[%v] Error decoding msgpack data: %v", c.id, data)
		}
		return
	}

	var ts int64
	var err error

	for {
		select {
		case message := <-c.sendc:
			if b, err := marshal(message.Params); err == nil {
				c.pc.Send(message.Method, b)
			} else {
				info.Printf("[%v] Error encoding as msgpack: %v", c.id, message.Params)
			}
		case m := <-c.pc.OnMethod:
			switch m.Method {
			case "stats_partial":
				if _, err = handleStats(m.Params); err != nil {
					info.Printf("[%v] Error decoding stats_partial: %v", c.id, err)
				}
			case "stats_complete":
				if ts, err = handleStats(m.Params); err != nil {
					info.Printf("[%v] Error decoding stats_complete: %v", c.id, err)
				} else {
					c.complete <- CompleteMessage{c.Id(), ts}
				}
			default:
				info.Printf("[%v] Received unknown command %v", c.id, m.Method)
			}
		case <-c.pc.OnClose:
			clientDidClose <- c.Id()
			return
		}
	}
}
