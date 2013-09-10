package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"time"
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
	hostname, _ := os.Hostname()
	var source = flag.String("source", hostname, "source (for reporting)")
	var interval = flag.Int("interval", 10, "stats interval (in seconds)")
	var timeout = flag.Int("timeout", 1000, "receive timeout (in ms)")
	var reg_addr = flag.String("registration", "tcp://127.0.0.1:5867", "address to which clients register")
	var librato_email = flag.String("librato_email", "", "librato email")
	var librato_token = flag.String("librato_token", "", "librato token")
	var http_addr = flag.String("http", "", "HTTP debugging address (e.g. ':8080')")
	flag.Parse()

	ts_complete := make(chan int64)
	ts_new := make(chan int64)
	on_shutdown := make(chan bool)
	on_complete := make(chan CompleteMessage)

	ticker := NewTicker(*interval)

	aggregator := NewAggregator()
	go aggregator.Run(ts_complete, ts_new)

	client_manager := NewClientManager()
	go client_manager.Run(ticker, *timeout, ts_complete, ts_new, on_complete, aggregator)

	gen_client := func(id int, a string) PairClient {
		return PairClient(NewClient(id, a, "", aggregator.stats, on_complete))
	}

	pair_server := NewPairServer(*reg_addr, on_shutdown)
	go pair_server.Run(PairServerDelegate(client_manager), gen_client)

	if len(*librato_email) > 0 && len(*librato_token) > 0 {
		librato := NewLibrato(*source, *librato_email, *librato_token)
		go RunOutput(aggregator.output, librato)
	} else {
		go RunOutput(aggregator.output, nil)
	}

	go func() {
		if *http_addr != "" {
			info.Printf("[main] HTTP debug server running on %v", *http_addr)
			log.Println(http.ListenAndServe(*http_addr, nil))
		}
	}()

	info.Printf("[main] Stagger running")

	// Handle termination
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	on_shutdown <- true
	// TODO: Otherwise on_shutdown doesn't happen - why?
	<-time.After(1 * time.Millisecond)
	info.Print("[main] Exiting cleanly")
}
