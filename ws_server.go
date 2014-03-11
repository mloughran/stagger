// Output receives aggregated data from the aggregator and exposes it to downstream ws
// Test with wscat --connect ws://127.0.0.1:8991/ --origin ws://127.0.0.1:8990/
// ./stagger -ws=127.0.0.1:8991

package main

import (
    "code.google.com/p/go.net/websocket"
    "sync"
    "encoding/json"
    "log"
    "time"
)

type Websocketsender struct {
       mutex sync.Mutex;
       connections map [*websocket.Conn] *sync.Cond;
}

func NewWebsocketsender() *Websocketsender {
        return &Websocketsender{connections: make(map[*websocket.Conn]*sync.Cond)}
}

func (s *Websocketsender) GetWebsocketSenderHandler() (websocket.Handler){
    var handlerf func(ws *websocket.Conn);
    handlerf = func(ws *websocket.Conn) {
       defer func() {
           log.Println("Ws connection closing")
           ws.Close()
           s.mutex.Lock()
           delete(s.connections,ws)
           s.mutex.Unlock()
           log.Println("Total connections", len(s.connections))
       }()
       locker := &sync.Mutex{}
       locker.Lock()
       s.mutex.Lock()
       s.connections[ws] = sync.NewCond(locker);
       s.mutex.Unlock()
       log.Println("Web socket connected:", ws.RemoteAddr(), "Total connections", len(s.connections))
       s.connections[ws].Wait()
       log.Println("Web socket shutting down:", ws.RemoteAddr())
   }
   return websocket.Handler(handlerf)
}

func (s *Websocketsender) Send(stats *TimestampedStats) {
     for ws,cond := range s.connections {
         if b, err := json.Marshal(stats); err == nil {
              ws.SetWriteDeadline(time.Now().Add(300*time.Millisecond))
              err := websocket.Message.Send(ws, b)
              if err != nil {
                        log.Println("Cannot deliver connection checker msg")
                        cond.Signal()
              }
         }
     }
}
