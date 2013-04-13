package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
)

func main() {
	reg_chan := make(chan Registration)
	stats_channels := NewStatsChannels()
	ts_complete := make(chan int64)
	ts_new := make(chan int64)

	go StartRegistration("tcp://127.0.0.1:5867", reg_chan)

	go StartClientManager(reg_chan, stats_channels, ts_complete, ts_new)

	output_chan := make(chan *TimestampedStats)

	go RunAggregator(stats_channels, ts_complete, ts_new, output_chan)

	go RunOutput(output_chan)

	// Just used for debugging
	go log.Fatal(http.ListenAndServe(":8080", nil))

	// Exit cleanly
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Print("[main] Exiting cleanly")
}
