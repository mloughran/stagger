package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
)

type debugger bool

func (d debugger) Printf(format string, args ...interface{}) {
	if d {
		log.Printf(format, args...)
	}
}

func (d debugger) Print(args ...interface{}) {
	if d {
		log.Print(args...)
	}
}

const debug debugger = false
const info debugger = true

func main() {
	var interval = flag.Int("interval", 10, "stats interval (in seconds)")
	var reg_addr = flag.String("registration", "tcp://127.0.0.1:5867", "address to which clients register")
	flag.Parse()

	reg_chan := make(chan Registration)
	stats_channels := NewStatsChannels()
	ts_complete := make(chan int64)
	ts_new := make(chan int64)

	go StartRegistration(*reg_addr, reg_chan)

	go StartClientManager(*interval, reg_chan, stats_channels, ts_complete, ts_new)

	output_chan := make(chan *TimestampedStats)

	go RunAggregator(stats_channels, ts_complete, ts_new, output_chan)

	go RunOutput(output_chan)

	// Just used for debugging
	go log.Fatal(http.ListenAndServe(":8080", nil))

	// Exit cleanly
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	info.Print("[main] Exiting cleanly")
}
