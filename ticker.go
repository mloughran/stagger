package main

import "time"

// Like a time.Tick, but anchored at time modulo boundary
func NewTicker(interval int) <-chan (time.Time) {

	period := time.Duration(interval) * time.Second
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
