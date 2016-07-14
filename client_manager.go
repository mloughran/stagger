// Each client is given a unique reference by the client manager. When stats are requested the client manager must make a record of all the clients for which the request was sent to, since it expects to get a reply from each one of them (which may be a different set from the current set of clients if a new client has just been registered).
//
// A set of clients is created for a given timestamp when the stats are requested. When a client goes away or a client has finished reporting all stats it is removed from this set. When the set is empty, or after a timeout (tbd) the aggregator is notified to say that a given timestamp should be considered complete. Any more stats for that timestamp arriving in the aggregator should then be thrown away.

package main

import (
	"github.com/pusher/stagger/tcp"
	"time"
)

// Sent by clients when they have finished receiving data for a timestamp
type CompleteMessage struct {
	ClientId  int64
	Timestamp int64
}

type ClientManager struct {
	addClientC  chan tcp.Connection
	remClient   chan tcp.Client
	sigShutdown chan bool
	didShutdown chan bool
	onComplete  chan (CompleteMessage)
	agg         *Aggregator
	index       int64
}

func NewClientManager(a *Aggregator) *ClientManager {
	return &ClientManager{
		make(chan tcp.Connection),
		make(chan tcp.Client),
		make(chan bool),
		make(chan bool),
		make(chan CompleteMessage),
		a,
		0,
	}
}

func (self *ClientManager) Run(interval time.Duration, timeout int, tsComplete, tsNew chan<- time.Time) {
	clients := make(map[int64]tcp.Client)
	ticker := NewTicker(interval)

	outstandingStats := map[time.Time]int{}

	// Notification on timeout for receiving data for a timestamp
	onTimeout := make(chan time.Time)

	// Avoid allocations
	var (
		t       time.Time
		latency float64
	)

	closeTimestamp := func(t time.Time) {
		delete(outstandingStats, t)
		tsComplete <- t // TODO: Notify that it wasn't clean
	}

	for {
		select {
		case conn := <-self.addClientC:
			self.index += 1
			client := NewClient(self.index, conn, self.agg.Stats, self.onComplete, self.remClient)
			clients[self.index] = client

			info.Printf("[cm] Added %s (count: %v)", client, len(clients))

		case client := <-self.remClient:
			delete(clients, client.Id())
			info.Printf("[cm] Removed %s (count: %v)", client, len(clients))

		case t = <-ticker:
			tsNew <- t
			if len(outstandingStats) > 1 {
				info.Printf("Too many outstanding stats: %v", outstandingStats)
			}
			if t.UnixNano()%int64(interval) != 0 {
				info.Printf("[cm] timestamp (t:%v) out of sync", t)
				continue // skip this beat
			}
			if len(clients) > 0 {
				info.Printf("[cm] (t:%v) Surveying %v clients", t, len(clients))

				// Store number of clients for this stat
				outstandingStats[t] = len(clients)

				// Record metric for number registered clients
				self.agg.Count(t.Unix(), "stagger.clients", Count(len(clients)))

				for _, client := range clients {
					client.RequestStats(t.Unix())
				}

				// Setup timeout to receive all the data
				go func(t time.Time) {
					time.Sleep(time.Duration(timeout) * time.Millisecond)
					onTimeout <- t
				}(t)
			} else {
				info.Printf("[cm] (t:%v) No clients connected to survey", t)
			}

		case t = <-onTimeout:
			if remaining, ok := outstandingStats[t]; ok {
				info.Printf("[cm] (t:%v) Survey timed out, %v clients yet to report", t, remaining)
				self.agg.Count(t.Unix(), "stagger.timeouts", Count(remaining))
				closeTimestamp(t)
			}

		case c := <-self.onComplete:
			t = time.Unix(c.Timestamp, 0)
			if left, ok := outstandingStats[t]; ok {
				if left <= 0 {
					info.Printf("[cm] Bad outstanding stats: %d", left)
				}

				outstandingStats[t] -= 1
				// Record the time for this client to complete survey in ms
				latency = float64(time.Now().Sub(t)) / float64(time.Millisecond)

				self.agg.Value(t.Unix(), "stagger.survey_latency", latency)

				if outstandingStats[t] == 0 {
					closeTimestamp(t)
				}
			}
		case <-self.sigShutdown:
			info.Printf("[cm] Requesting shutdown from clients")
			for _, client := range clients {
				client.Shutdown()
			}

			// TODO: wait for all of the clients to disappear
			time.Sleep(1 * time.Second)

			self.didShutdown <- true
		}
	}
}

func (self *ClientManager) NewClient(c tcp.Connection) {
	self.addClientC <- c
}

func (self *ClientManager) Shutdown() {
	self.sigShutdown <- true
	<-self.didShutdown
}
