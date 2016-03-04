package main

import "time"

// Like a time.Tick, but anchored at time modulo boundary
func NewTicker(interval int) <-chan (time.Time) {
	period := time.Duration(interval) * time.Second
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
