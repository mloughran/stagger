package main

import (
	"flag"
	"github.com/pusher/stagger/outputter"
	"github.com/pusher/stagger/tcp"
	"github.com/pusher/stagger/tcp/v1"
	"github.com/pusher/stagger/tcp/v2"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
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
	debug     debugger = false
	info      debugger = true
	buildSha  string   = "<unknown>"
	buildDate string   = "<unknown>"
)

func main() {
	hostname, _ := os.Hostname()
	var (
		http_addr     = flag.String("http", "127.0.0.1:8990", "HTTP debugging address (e.g. ':8990')")
		influxdb_url  = flag.String("influxdb_url", "", "influxdb URL")
		interval      = flag.Int("interval", 5, "stats interval (in seconds)")
		librato_email = flag.String("librato_email", "", "librato email")
		librato_token = flag.String("librato_token", "", "librato token")
		log_output    = flag.Bool("log_output", true, "log aggregated data")
		showDebug     = flag.Bool("debug", false, "Print debug information")
		source        = flag.String("source", hostname, "source (for reporting)")
		tcp_addr      = flag.String("addr", "tcp://127.0.0.1:5866", "adress for the TCP v1 mode")
		tcp_addr2     = flag.String("addr2", "tcp://127.0.0.1:5865", "adress for the TCP v2 mode")
		timeout       = flag.Int("timeout", 1000, "receive timeout (in ms)")
	)
	tags := NewTagsValue(hostname)
	flag.Var(tags, "tag", "adds key=value to stats (only influxdb)")
	flag.Parse()

	if showDebug != nil && *showDebug {
		debug = true
	}

	// Make sure clients have time to report
	if *interval <= 2 {
		log.Println("Bad interval %d, changing to 2", interval)
		*interval = 2
	}

	ts_complete := make(chan int64)
	ts_new := make(chan int64)

	ticker := NewTicker(*interval)

	aggregator := NewAggregator()
	go aggregator.Run(ts_complete, ts_new)

	client_manager := NewClientManager(aggregator)
	go client_manager.Run(ticker, *timeout, ts_complete, ts_new)

	tcp_server, err := tcp.NewServer(*tcp_addr, client_manager, v1.Encoding{}, *interval)
	if err != nil {
		log.Println("invalid address: ", err)
		return
	}
	go tcp_server.Run()

	tcp_server2, err := tcp.NewServer(*tcp_addr2, client_manager, v2.Encoding{}, *interval)
	if err != nil {
		log.Println("invalid address: ", err)
		return
	}
	go tcp_server2.Run()

	group := outputter.NewGroup()

	logger := log.New(os.Stderr, "", log.LstdFlags)
	if len(*librato_email) > 0 && len(*librato_token) > 0 {
		librato := outputter.NewLibrato(*source, *librato_email, *librato_token, logger)
		group.Add(librato)
	}

	if *influxdb_url != "" {
		influxdb, err := outputter.NewInfluxDB(tags.Value(), *influxdb_url, logger)
		if err != nil {
			log.Println("InfluxDB error: ", err)
			return
		} else {
			group.Add(influxdb)
		}
	}

	if *log_output {
		group.Add(outputter.Stdout)
	}
	if *http_addr != "" {
		snapshot := outputter.NewSnapshot()
		group.Add(snapshot)
		http.Handle("/snapshot.json", snapshot)

		go func() {
			info.Printf("[main] HTTP server running on %v", *http_addr)
			log.Println(http.ListenAndServe(*http_addr, nil))
		}()
	}

	go func() {
		for msg := range aggregator.output {
			group.Send(msg)
		}
	}()

	info.Printf("[main] Stagger running. sha=%s date=%s go=%s", buildSha, buildDate, runtime.Version())

	// Wait forever
	<-make(chan os.Signal, 1)
}
