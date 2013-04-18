// Each client is given a unique reference by the client manager. When stats are requested the client manager must make a record of all the clients for which the request was sent to, since it expects to get a reply from each one of them (which may be a different set from the current set of clients if a new client has just been registered).
//
// A set of clients is created for a given timestamp when the stats are requested. When a client goes away or a client has finished reporting all stats it is removed from this set. When the set is empty, or after a timeout (tbd) the aggregator is notified to say that a given timestamp should be considered complete. Any more stats for that timestamp arriving in the aggregator should then be thrown away.

package main

import (
	"time"
)

type ClientRef struct {
	Id           int
	RequestStats chan (int64) // Request stats from a client for some ts
}

// Sent by clients when they have finished receiving data for a timestamp
type CompleteMessage struct {
	ClientId  int
	Timestamp int64
}

// Like a time.Tick, but anchored at time modulo boundary
func AnchoredTick(period time.Duration) chan (time.Time) {
	ticks := make(chan time.Time)
	go func() {
		// Wait till the end of the current period
		elapsed := time.Now().UnixNano() % period.Nanoseconds()
		now := <-time.After(time.Duration(period.Nanoseconds() - elapsed))

		// Use Ticker to tick regularly
		tick_chan := time.Tick(period)
		ticks <- now
		for {
			ticks <- <-tick_chan
		}
	}()
	return ticks
}

func StartClientManager(interval int, registration chan (Registration), stats_channels StatsChannels, ts_complete, ts_new chan (int64)) {
	clients := make(map[int]ClientRef)

	heartbeat := AnchoredTick(time.Duration(interval) * time.Second)

	complete := make(chan CompleteMessage)

	// Clients send a message on this channel when they go away
	on_client_gone := make(chan int)

	client_id_incr := 0

	outstanding_stats := map[int64]int{}

	// Notification on timeout for receiving data for a timestamp
	on_timeout := make(chan int64)

	// Avoid allocations
	var ts int64

	for {
		select {
		case reg := <-registration:
			client_id_incr += 1

			client := ClientRef{client_id_incr, make(chan int64)}
			go RunClient(reg, client, stats_channels, complete, on_client_gone)
			clients[client_id_incr] = client
			info.Printf("[cm] Added client %v (count: %v)", client_id_incr, len(clients))
		case id := <-on_client_gone:
			delete(clients, id)
			info.Printf("[cm] Removed client %v (count: %v)", id, len(clients))
			// TODO: Should also remove this client from the list of unreported
			// clients, but this requires using more than a simple count
		case now := <-heartbeat:
			if len(clients) > 0 {
				ts = now.Unix()
				debug.Print("[cm] Sending request for stats at ", ts)

				// Store number of clients for this stat
				outstanding_stats[ts] = len(clients)

				ts_new <- ts

				// Send stats request to each client
				for _, client := range clients {
					client.RequestStats <- ts
				}

				// Setup timeout to receive all the data
				go func() {
					<-time.After(3 * time.Second)
					on_timeout <- ts
				}()
			}
		case ts = <-on_timeout:
			if remaining, present := outstanding_stats[ts]; present {
				debug.Printf("[cm] Timeout exceeded for ts %v, %v clients yet to report", ts, remaining)
				ts_complete <- ts // TODO: Notify that it wasn't clean
			}
		case c := <-complete:
			ts = c.Timestamp
			// TODO: Handle possibility that this does not exist
			outstanding_stats[ts] -= 1

			if outstanding_stats[ts] == 0 {
				delete(outstanding_stats, ts)
				if debug {
					t := time.Unix(ts, 0)
					debug.Printf("[cm] Received from all clients for ts %v (%v)", ts, t)
				}
				ts_complete <- ts
			}
		}
	}
}
