// Each client is given a unique reference by the client manager. When stats are requested the client manager must make a record of all the clients for which the request was sent to, since it expects to get a reply from each one of them (which may be a different set from the current set of clients if a new client has just been registered).
//
// A set of clients is created for a given timestamp when the stats are requested. When a client goes away or a client has finished reporting all stats it is removed from this set. When the set is empty, or after a timeout (tbd) the aggregator is notified to say that a given timestamp should be considered complete. Any more stats for that timestamp arriving in the aggregator should then be thrown away.

package main

import "time"

// Sent by clients when they have finished receiving data for a timestamp
type CompleteMessage struct {
	ClientId  int
	Timestamp int64
}

type ClientManager struct {
	add_client_c chan (*Client)
	rem_client_c chan (*Client)
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		make(chan (*Client)),
		make(chan (*Client)),
	}
}

func (self *ClientManager) Run(ticker <-chan (time.Time), timeout int, ts_complete, ts_new chan<- (int64), complete <-chan (CompleteMessage), aggregator *Aggregator) {
	clients := make(map[int]*Client)

	outstanding_stats := map[int64]int{}

	// Notification on timeout for receiving data for a timestamp
	on_timeout := make(chan int64)

	// Avoid allocations
	var ts int64
	var now time.Time
	var latency float64

	for {
		select {
		case client := <-self.add_client_c:
			info.Printf("[cm] Added client id:%v (count: %v)", client.Id(), len(clients))
			clients[client.Id()] = client

		case client := <-self.rem_client_c:
			info.Printf("[cm] Removed client id:%v (count: %v)", client.Id(), len(clients))
			delete(clients, client.Id())

		case now = <-ticker:
			ts = now.Unix()
			ts_new <- ts
			if len(clients) > 0 {
				info.Printf("[cm] (ts:%v) Surveying %v clients", ts, len(clients))

				// Store number of clients for this stat
				outstanding_stats[ts] = len(clients)

				// Record metric for number registered clients
				aggregator.Count(ts, "stagger.clients", Count(len(clients)))

				for _, client := range clients {
					client.RequestStats(ts)
				}

				// Setup timeout to receive all the data
				go func(ts int64) {
					<-time.After(time.Duration(timeout) * time.Millisecond)
					on_timeout <- ts
				}(ts)
			} else {
				info.Printf("[cm] (ts:%v) No clients connected to survey", ts)
			}

		case ts = <-on_timeout:
			if remaining, ok := outstanding_stats[ts]; ok {
				info.Printf("[cm] (ts:%v) Survey timed out, %v clients yet to report", ts, remaining)
				aggregator.Count(ts, "stagger.timeouts", Count(remaining))
				delete(outstanding_stats, ts)
				ts_complete <- ts // TODO: Notify that it wasn't clean
			}

		case c := <-complete:
			ts = c.Timestamp
			if _, ok := outstanding_stats[ts]; ok {
				outstanding_stats[ts] -= 1

				// Record the time for this client to complete survey in ms
				latency = float64(time.Now().UnixNano()-ts*1000000000) / 1000000
				aggregator.Value(ts, "stagger.survey_latency", latency)

				if outstanding_stats[ts] == 0 {
					delete(outstanding_stats, ts)
					ts_complete <- ts
				}
			}
		}
	}
}

func (self *ClientManager) AddClient(client interface{}) {
	self.add_client_c <- client.(*Client)
}

func (self *ClientManager) RemoveClient(client interface{}) {
	self.rem_client_c <- client.(*Client)
}
