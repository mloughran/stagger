package main

import (
	"flag"
	"github.com/pusher/stagger/pair"
	"log"
	"os"
	"os/signal"
	"syscall"
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

var (
	debug debugger = false
	info  debugger = true
	build string
)

func init() {
	if len(build) == 0 {
		build = "Build with `go build -ldflags \"-X main.build <info-string>\"` to include build information in binary"
	}
}

func main() {
	hostname, _ := os.Hostname()
	var (
		source        = flag.String("source", hostname, "source (for reporting)")
		interval      = flag.Int("interval", 10, "stats interval (in seconds)")
		timeout       = flag.Int("timeout", 1000, "receive timeout (in ms)")
		reg_addr      = flag.String("registration", "tcp://127.0.0.1:5867", "address to which clients register")
		log_output    = flag.Bool("log_output", true, "log aggregated data")
		librato_email = flag.String("librato_email", "", "librato email")
		librato_token = flag.String("librato_token", "", "librato token")
		showDebug     = flag.Bool("debug", false, "Print debug information")
	)
	flag.Parse()

	if showDebug != nil && *showDebug {
		debug = true
	}

	ts_complete := make(chan int64)
	ts_new := make(chan int64)

	ticker := NewTicker(*interval)

	aggregator := NewAggregator()
	go aggregator.Run(ts_complete, ts_new)

	client_manager := NewClientManager(aggregator)
	go client_manager.Run(ticker, *timeout, ts_complete, ts_new)

	pair_server := pair.NewServer(*reg_addr, pair.ServerDelegate(client_manager))
	go pair_server.Run()

	output := NewOutput()

	if len(*librato_email) > 0 && len(*librato_token) > 0 {
		librato := NewLibrato(*source, *librato_email, *librato_token)
		go librato.Run()
		output.Add(librato)
	}

	if *log_output {
		stdout := NewStdOut()
		output.Add(stdout)
	}

	go output.Run(aggregator.output)

	info.Printf("[main] Stagger running. %s", build)

	// Handle termination
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	pair_server.Shutdown()
	info.Print("[main] Exiting cleanly")
}
