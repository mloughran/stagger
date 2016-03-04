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
		httpAddr     = flag.String("http", "127.0.0.1:8990", "HTTP debugging address (e.g. ':8990')")
		influxdbUrl  = flag.String("influxdb_url", "", "influxdb URL")
		interval     = flag.Int("interval", 5, "stats interval (in seconds)")
		libratoEmail = flag.String("librato_email", "", "librato email")
		libratoToken = flag.String("librato_token", "", "librato token")
		logOutput    = flag.Bool("log_output", true, "log aggregated data")
		showDebug    = flag.Bool("debug", false, "Print debug information")
		source       = flag.String("source", hostname, "source (for reporting)")
		tcpAddr      = flag.String("addr", "tcp://127.0.0.1:5866", "adress for the TCP v1 mode")
		tcpAddr2     = flag.String("addr2", "tcp://127.0.0.1:5865", "adress for the TCP v2 mode")
		timeout      = flag.Int("timeout", 1000, "receive timeout (in ms)")
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

	tsComplete := make(chan int64)
	tsNew := make(chan int64)

	ticker := NewTicker(*interval)

	aggregator := NewAggregator()
	go aggregator.Run(tsComplete, tsNew)

	clientManager := NewClientManager(aggregator)
	go clientManager.Run(ticker, *timeout, tsComplete, tsNew)

	tcpServer, err := tcp.NewServer(*tcpAddr, clientManager, v1.Encoding{}, *interval)
	if err != nil {
		log.Println("invalid address: ", err)
		return
	}
	go tcpServer.Run()

	tcpServer2, err := tcp.NewServer(*tcpAddr2, clientManager, v2.Encoding{}, *interval)
	if err != nil {
		log.Println("invalid address: ", err)
		return
	}
	go tcpServer2.Run()

	group := outputter.NewGroup()

	logger := log.New(os.Stderr, "", log.LstdFlags)
	if len(*libratoEmail) > 0 && len(*libratoToken) > 0 {
		librato := outputter.NewLibrato(*source, *libratoEmail, *libratoToken, logger)
		group.Add(librato)
	}

	if *influxdbUrl != "" {
		influxdb, err := outputter.NewInfluxDB(tags.Value(), *influxdbUrl, logger)
		if err != nil {
			log.Println("InfluxDB error: ", err)
			return
		} else {
			group.Add(influxdb)
		}
	}

	if *logOutput {
		group.Add(outputter.Stdout)
	}
	if *httpAddr != "" {
		snapshot := outputter.NewSnapshot()
		group.Add(snapshot)
		http.Handle("/snapshot.json", snapshot)

		go func() {
			info.Printf("[main] HTTP server running on %v", *httpAddr)
			log.Println(http.ListenAndServe(*httpAddr, nil))
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
