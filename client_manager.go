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

func StartClientManager(ticker chan (time.Time), timeout int, regc chan (*Client), statsc chan (*Stats), ts_complete, ts_new chan (int64), on_shutdown chan (bool)) {
	clients := make(map[int]*Client)

	complete := make(chan CompleteMessage)

	// Clients send a message on this channel when they go away
	on_client_gone := make(chan int)

	outstanding_stats := map[int64]int{}

	// Notification on timeout for receiving data for a timestamp
	on_timeout := make(chan int64)

	// Avoid allocations
	var ts int64
	var now time.Time

	for {
		select {
		case client := <-regc:
			go client.Run(statsc, complete, on_client_gone)

			clients[client.Id] = client
			info.Printf("[cm] Added client %v (count: %v)", client.Id, len(clients))
		case id := <-on_client_gone:
			delete(clients, id)
			info.Printf("[cm] Removed client %v (count: %v)", id, len(clients))
			// TODO: Should also remove this client from the list of unreported
			// clients, but this requires using more than a simple count
		case <-on_shutdown:
			info.Printf("[cm] Sending shutdown message to all clients")
			for _, client := range clients {
				client.Shutdown()
			}
		case now = <-ticker:
			if len(clients) > 0 {
				ts = now.Unix()
				info.Printf("Requesting at %v from %v clients", ts, len(clients))

				// Store number of clients for this stat
				outstanding_stats[ts] = len(clients)

				ts_new <- ts

				for _, client := range clients {
					client.RequestStats(ts)
				}

				// Setup timeout to receive all the data
				go func(ts int64) {
					<-time.After(time.Duration(timeout) * time.Millisecond)
					on_timeout <- ts
				}(ts)
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
