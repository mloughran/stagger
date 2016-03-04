// Each client is given a unique reference by the client manager. When stats are requested the client manager must make a record of all the clients for which the request was sent to, since it expects to get a reply from each one of them (which may be a different set from the current set of clients if a new client has just been registered).
//
// A set of clients is created for a given timestamp when the stats are requested. When a client goes away or a client has finished reporting all stats it is removed from this set. When the set is empty, or after a timeout (tbd) the aggregator is notified to say that a given timestamp should be considered complete. Any more stats for that timestamp arriving in the aggregator should then be thrown away.

package main

import (
	"github.com/pusher/stagger/conn"
	"time"
)

// Sent by clients when they have finished receiving data for a timestamp
type CompleteMessage struct {
	ClientId  int64
	Timestamp int64
}

type ClientManager struct {
	addClientC  chan conn.Connection
	remClient   chan conn.Client
	sigShutdown chan bool
	didShutdown chan bool
	onComplete  chan (CompleteMessage)
	agg         *Aggregator
	index       int64
}

func NewClientManager(a *Aggregator) *ClientManager {
	return &ClientManager{
		make(chan conn.Connection),
		make(chan conn.Client),
		make(chan bool),
		make(chan bool),
		make(chan CompleteMessage),
		a,
		0,
	}
}

func (self *ClientManager) Run(ticker <-chan (time.Time), timeout int, tsComplete, tsNew chan<- int64) {
	clients := make(map[int64]conn.Client)

	outstandingStats := map[int64]int{}

	// Stores the nanosecond time at which a timestamp was emitted, which may
	// be a few ms after the second. This is used to calculate a more precise
	// survey_latency
	nanoTs := map[int64]int64{}

	// Notification on timeout for receiving data for a timestamp
	onTimeout := make(chan int64)

	// Avoid allocations
	var (
		ts, tsn int64
		now     time.Time
		latency float64
	)

	closeTimestamp := func(ts int64) {
		delete(outstandingStats, ts)
		delete(nanoTs, ts)
		tsComplete <- ts // TODO: Notify that it wasn't clean
	}

	for {
		select {
		case conn := <-self.addClientC:
			self.index += 1
			client := NewClient(self.index, conn, self.agg.Stats, self.onComplete)
			clients[self.index] = client
			go client.Run(self.remClient)

			info.Printf("[cm] Added %s (count: %v)", client, len(clients))

		case client := <-self.remClient:
			delete(clients, client.Id())
			info.Printf("[cm] Removed %s (count: %v)", client, len(clients))

		case now = <-ticker:
			ts = now.Unix()
			nanoTs[ts] = now.UnixNano()
			tsNew <- ts
			if len(outstandingStats) > 1 {
				info.Printf("Too many outstanding stats: %v", outstandingStats)
			}
			if len(clients) > 0 {
				info.Printf("[cm] (ts:%v) Surveying %v clients", ts, len(clients))

				// Store number of clients for this stat
				outstandingStats[ts] = len(clients)

				// Record metric for number registered clients
				self.agg.Count(ts, "stagger.clients", Count(len(clients)))

				for _, client := range clients {
					client.RequestStats(ts)
				}

				// Setup timeout to receive all the data
				go func(ts int64) {
					time.Sleep(time.Duration(timeout) * time.Millisecond)
					onTimeout <- ts
				}(ts)
			} else {
				info.Printf("[cm] (ts:%v) No clients connected to survey", ts)
			}

		case ts = <-onTimeout:
			if remaining, ok := outstandingStats[ts]; ok {
				info.Printf("[cm] (ts:%v) Survey timed out, %v clients yet to report", ts, remaining)
				self.agg.Count(ts, "stagger.timeouts", Count(remaining))
				closeTimestamp(ts)
			}

		case c := <-self.onComplete:
			ts = c.Timestamp
			if left, ok := outstandingStats[ts]; ok {
				if left <= 0 {
					info.Printf("[cm] Bad outstanding stats: %d", left)
				}

				tsn = nanoTs[ts]
				outstandingStats[ts] -= 1
				// Record the time for this client to complete survey in ms
				latency = float64(time.Now().UnixNano()-tsn) / 1000000

				self.agg.Value(ts, "stagger.survey_latency", latency)

				if outstandingStats[ts] == 0 {
					closeTimestamp(ts)
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

func (self *ClientManager) NewClient(c conn.Connection) {
	self.addClientC <- c
}

func (self *ClientManager) Shutdown() {
	self.sigShutdown <- true
	<-self.didShutdown
}
