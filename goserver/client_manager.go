// Each client is given a unique reference by the client manager. When stats are requested the client manager must make a record of all the clients for which the request was sent to, since it expects to get a reply from each one of them (which may be a different set from the current set of clients if a new client has just been registered).
//
// A set of clients is created for a given timestamp when the stats are requested. When a client goes away or a client has finished reporting all stats it is removed from this set. When the set is empty, or after a timeout (tbd) the aggregator is notified to say that a given timestamp should be considered complete. Any more stats for that timestamp arriving in the aggregator should then be thrown away.

package main

import "log"
import "time"

type ClientRef struct {
	Id           int
	RequestStats chan (int64) // Request stats from a client for some ts
}

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

func StartClientManager(registration chan (Registration), stats_channels StatsChannels, ts_complete chan (int64), ts_new chan (int64)) {
	clients := make([]ClientRef, 0)

	heartbeat := AnchoredTick(5 * time.Second)

	complete := make(chan CompleteMessage)

	client_id_incr := 1

	outstanding_stats := map[int64]int{}

	for {
		select {
		case reg := <-registration:
			client := ClientRef{client_id_incr, make(chan int64)}
			client_id_incr += 1
			go RunClient(reg, client, stats_channels, complete)
			clients = append(clients, client)

			log.Print("[cm] Managing clients: ", len(clients))
		case time := <-heartbeat:
			if len(clients) > 0 {
				unix_ts := time.Unix()
				log.Print("[cm] Sending request for stats at ", unix_ts)

				// Store number of clients for this stat
				outstanding_stats[unix_ts] = len(clients)

				ts_new <- unix_ts

				// Send stats request to each client
				for _, client := range clients {
					client.RequestStats <- unix_ts
				}
			}
		case c := <-complete:
			// TODO: Handle possibility that this does not exist
			outstanding_stats[c.Timestamp] -= 1

			if outstanding_stats[c.Timestamp] == 0 {
				delete(outstanding_stats, c.Timestamp)
				log.Print("[cm] Stats done for ts ", c.Timestamp)
				ts_complete <- c.Timestamp
			}
		}
	}
}
