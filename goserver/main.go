package main

import "log"
import "os"
import "os/signal"
import "net/http"
import _ "net/http/pprof"

func main() {
	reg_chan := make(chan Registration)
	stats_channels := NewStatsChannels()
	ts_complete := make(chan int64)
	ts_new := make(chan int64)

	go StartRegistration("tcp://127.0.0.1:2900", reg_chan)

	go StartClientManager(reg_chan, stats_channels, ts_complete, ts_new)

	go RunAggregator(stats_channels, ts_complete, ts_new)

	// Just used for debugging
	go log.Fatal(http.ListenAndServe(":8080", nil))

	// Exit cleanly
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Print("[main] Exiting cleanly")
}
