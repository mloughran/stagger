package main

import (
	"fmt"
	"github.com/pusher/stagger/conn"
	"strings"
)

type RegisterProcess struct {
	Tags map[string]string
}

type message struct {
	Method string
	Params map[string]interface{}
}

type Client struct {
	id       int64
	conn     conn.Connection
	name     string
	sendc    chan (message)
	statsc   chan<- (*Stats)
	complete chan<- (CompleteMessage)
}

func NewClient(id int64, c conn.Connection, statsc chan<- (*Stats), complete chan<- (CompleteMessage)) *Client {
	client := &Client{
		id,
		c,
		"",
		make(chan message, 2),
		statsc,
		complete,
	}
	client.setName(nil)
	return client
}

func (c *Client) Id() int64 {
	return c.id
}

func (c *Client) String() string {
	return c.name
}

func (c *Client) Send(m string, p map[string]interface{}) {
	c.sendc <- message{m, p}
}

func (c *Client) RequestStats(ts int64) {
	// TODO: Make Timestamp lowercase
	c.Send("report_all", map[string]interface{}{"Timestamp": ts})
}

func (c *Client) Shutdown() {
	c.conn.Shutdown()
}

func (c *Client) Run(clientDidClose chan<- conn.Client) {
	handleStats := func(data []byte) (ts int64, err error) {
		var stats Stats
		if err = unmarshal(data, &stats); err == nil {
			ts = stats.Timestamp
			c.statsc <- &stats
		} else {
			info.Printf("%s Error decoding msgpack data: %v", c, data)
		}
		return
	}

	var ts int64
	var err error

	for {
		select {
		case message := <-c.sendc:
			if b, err := marshal(message.Params); err == nil {
				c.conn.Send(message.Method, b)
			} else {
				info.Printf("%s Error encoding as msgpack: %v", c, message.Params)
			}
		case m := <-c.conn.OnMethod():
			switch m.Method {
			case "register_process":
				var reg RegisterProcess
				if err = unmarshal(m.Params, &reg); err != nil {
					info.Printf("%s Error registering: %v", c, err)
				} else {
					c.setName(reg.Tags)
					info.Printf("%s Registered", c)
				}
			case "stats_partial":
				if ts, err = handleStats(m.Params); err != nil {
					info.Printf("%s Error decoding stats_partial: %v", c, err)
				}
			case "stats_complete":
				if ts, err = handleStats(m.Params); err != nil {
					info.Printf("%s Error decoding stats_complete: %v", c, err)
				} else {
					c.complete <- CompleteMessage{c.id, ts}
				}
			default:
				info.Printf("%s Received unknown command %v", c, m.Method)
			}
		case <-c.conn.OnClose():
			clientDidClose <- c
			return
		}
	}
}

// Sets the client String() representation based on the list of tags
func (c *Client) setName(tags map[string]string) {
	var tail []string
	if tags != nil && len(tags) > 0 {
		tail = make([]string, len(tags))
		i := 0
		for k, v := range tags {
			tail[i] = k + "=" + v
			i++
		}
	}
	c.name = fmt.Sprintf("[client id=%d %s tags={%s}]", c.id, c.conn, strings.Join(tail, " "))
}
