// Output receives aggregated data from the aggregator and exposes it to downstream ws
// Test with wscat --connect ws://127.0.0.1:8991/ --origin ws://127.0.0.1:8990/
// ./stagger -http=127.0.0.1:8991 -http_features=ws

package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"log"
	"sync"
	"time"
)

type Websocketsender struct {
	mutex       sync.Mutex
	connections map[*websocket.Conn]*sync.Cond
}

func NewWebsocketsender() *Websocketsender {
	return &Websocketsender{connections: make(map[*websocket.Conn]*sync.Cond)}
}

func (s *Websocketsender) GetWebsocketSenderHandler() websocket.Handler {
	handlerf := func(ws *websocket.Conn) {
		defer func() {
			log.Println("Web socket connection closing")
			ws.Close()
			s.mutex.Lock()
			delete(s.connections, ws)
			s.mutex.Unlock()
			log.Println("Web socket closed. Total connections", len(s.connections))
		}()
		s.mutex.Lock()
		s.connections[ws] = sync.NewCond(&sync.Mutex{})
		s.connections[ws].L.Lock()
		s.mutex.Unlock()
		if b, err := json.Marshal(NewTimestampedStatsWithTypes(-1)); err == nil {
			err := websocket.Message.Send(ws, string(b))
			if err != nil {
				log.Println("Web socket: Cannot deliver initial message")
			}
		}
		log.Println("Web socket connected:", ws.RemoteAddr(), "Total connections", len(s.connections))
		s.connections[ws].Wait()
	}
	return websocket.Handler(handlerf)
}

func (s *Websocketsender) Send(stats *TimestampedStats) {
	for ws, cond := range s.connections {
		if b, err := json.Marshal(stats); err == nil {
			ws.SetWriteDeadline(time.Now().Add(300 * time.Millisecond))
			err := websocket.Message.Send(ws, string(b))
			if err != nil {
				log.Println("Web socket: Cannot deliver msg")
				cond.Signal()
			}
		}
	}
}
