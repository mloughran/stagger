package main

import (
	"fmt"
	"github.com/pusher/stagger/conn"
	"github.com/pusher/stagger/tcp"

	"strings"
)

type Client struct {
	id       int64
	conn     tcp.Connection
	name     string
	sendc    chan conn.Message
	statsc   chan<- (*conn.Stats)
	complete chan<- (CompleteMessage)
}

func NewClient(id int64, c tcp.Connection, statsc chan<- (*conn.Stats), complete chan<- (CompleteMessage), clientClosed chan<- tcp.Client) *Client {
	client := &Client{
		id,
		c,
		"",
		make(chan conn.Message, 2),
		statsc,
		complete,
	}
	client.setName(nil)
	go client.run(clientClosed)
	return client
}

func (c *Client) Id() int64 {
	return c.id
}

func (c *Client) String() string {
	return c.name
}

func (c *Client) Send(msg conn.Message) {
	c.sendc <- msg
}

func (c *Client) RequestStats(ts int64) {
	// TODO: Make Timestamp lowercase
	c.Send(conn.ReportAll{Timestamp: ts})
}

func (c *Client) Shutdown() {
	c.conn.Shutdown()
}

func (c *Client) run(clientDidClose chan<- tcp.Client) {
	var ts int64

	for {
		select {
		case message := <-c.sendc:
			if err := c.conn.Send(message); err != nil {
				info.Printf("%s Error sending message: %v", c, err)
			}
		case message := <-c.conn.OnMethod():
			switch m := message.(type) {
			case *conn.RegisterProcess:
				c.setName(m.Tags)
				info.Printf("%s Registered", c)
			case *conn.StatsPartial:
				c.statsc <- &m.Stats
			case *conn.StatsComplete:
				c.statsc <- &m.Stats
				c.complete <- CompleteMessage{c.id, ts}
			default:
				info.Printf("%s Received unknown command %v", c, m)
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
