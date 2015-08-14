package main

import (
	"flag"
	"github.com/pusher/stagger/pair"
	"github.com/pusher/stagger/tcp"
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
		tcp_addr      = flag.String("addr", "tcp://127.0.0.1:5866", "adress for the TCP mode")
		reg_addr      = flag.String("registration", "tcp://127.0.0.1:5867", "address to which clients register")
		log_output    = flag.Bool("log_output", true, "log aggregated data")
		influxdb_url  = flag.String("influxdb_url", "", "influxdb URL")
		librato_email = flag.String("librato_email", "", "librato email")
		librato_token = flag.String("librato_token", "", "librato token")
		showDebug     = flag.Bool("debug", false, "Print debug information")
	)
	tags := NewTagsValue(hostname)
	flag.Var(tags, "tag", "adds key=value to stats (only influxdb)")
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

	tcp_server, err := tcp.NewServer(*tcp_addr, client_manager)
	if err != nil {
		log.Println("invalid address: ", err)
		return
	}
	go tcp_server.Run()

	pair_server := pair.NewServer(*reg_addr, client_manager)
	go pair_server.Run()

	output := NewOutput()

	if len(*librato_email) > 0 && len(*librato_token) > 0 {
		librato := NewLibrato(*source, *librato_email, *librato_token)
		go librato.Run()
		output.Add(librato)
	}

	if *influxdb_url != "" {
		influxdb, err := NewInfluxDB(tags.Value(), *influxdb_url)
		if err != nil {
			log.Println("InfluxDB error: ", err)
			return
		} else {
			go influxdb.Run()
			output.Add(influxdb)
		}
	}

	if *log_output {
		output.Add(StdoutOutputter)
	}

	go output.Run(aggregator.output)

	info.Printf("[main] Stagger running. %s", build)

	// Handle termination
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	tcp_server.Shutdown()
	pair_server.Shutdown()
	client_manager.Shutdown()
	info.Print("[main] Exiting cleanly")
}
