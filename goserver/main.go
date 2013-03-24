package main

import "log"
import "os"
import "os/signal"
import "net/http"
import _ "net/http/pprof"

func main() {
	reg_chan := make(chan string)
	stat_chan := make(chan Stat)

	go StartRegistration(reg_chan)

	go StartClientManager(reg_chan, stat_chan)

	go RunAggregator(stat_chan)

	// Just used for debugging
	go log.Fatal(http.ListenAndServe(":8080", nil))

	// Exit cleanly
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Print("[main] Exiting cleanly")
}
