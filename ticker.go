package main

import "time"

// Like a time.Tick, but anchored at time modulo boundary
func NewTicker(period time.Duration) <-chan (time.Time) {
	ticks := make(chan time.Time)
	go func() {
		// Wait till the end of the current period
		elapsed := time.Now().UnixNano() % period.Nanoseconds()
		time.Sleep(time.Duration(period.Nanoseconds() - elapsed))

		// Use Ticker to tick regularly
		tickChan := time.Tick(period)
		for {
			ticks <- <-tickChan
		}
	}()
	return ticks
}
