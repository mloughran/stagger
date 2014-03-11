package main

import (
	"./pair"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
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

const debug debugger = false
const info debugger = true

var build string

func main() {
	hostname, _ := os.Hostname()
	var source = flag.String("source", hostname, "source (for reporting)")
	var interval = flag.Int("interval", 10, "stats interval (in seconds)")
	var timeout = flag.Int("timeout", 1000, "receive timeout (in ms)")
	var reg_addr = flag.String("registration", "tcp://127.0.0.1:5867", "address to which clients register")
	var log_output = flag.Bool("log_output", true, "log aggregated data")
	var librato_email = flag.String("librato_email", "", "librato email")
	var librato_token = flag.String("librato_token", "", "librato token")
        var http_addr = flag.String("http", "", "HTTP debugging address (e.g. ':8080')")
        var ws_addr = flag.String("ws", "", "WS stream address (e.g. ':8090')")
	var showBuild = flag.Bool("build", false, "Print build information")
	flag.Parse()

	if *showBuild {
		if len(build) > 0 {
			fmt.Println(build)
		} else {
			fmt.Println("Build with `go build -ldflags \"-X main.build <info-string>\"` to include build information in binary")
		}
		os.Exit(0)
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

	if *http_addr != "" {
		snapshot := NewSnapshot()
		output.Add(snapshot)

		go func() {
			http.Handle("/snapshot.json", snapshot)

			info.Printf("[main] HTTP debug server running on %v", *http_addr)
			log.Println(http.ListenAndServe(*http_addr, nil))
		}()
	}

        if *ws_addr != "" {
		websocketsender := NewWebsocketsender()
		output.Add(websocketsender)

                go func() {
                        http.Handle("/",websocketsender.GetWebsocketSenderHandler())
			info.Printf("[main] WS stream server running on %v", *ws_addr)
			log.Println(http.ListenAndServe(*ws_addr, nil))
		}()

	}

	go output.Run(aggregator.output)

	info.Printf("[main] Stagger running")

	// Handle termination
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	pair_server.Shutdown()
	info.Print("[main] Exiting cleanly")
}
