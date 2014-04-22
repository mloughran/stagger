package main

import (
	"./pair"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
	http_features_string := flag.String("features", "ws-json,http-json,sparkline", "HTTP features (ws-json,http-json,sparkline)")
	http_features := make(map[string]bool)
	for _, s := range strings.Split(*http_features_string, ",") {
		http_features[s] = true
	}
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
		if _, ok := http_features["http-json"]; ok {
			log.Println("Http json enabled at http://" + *http_addr + "/snapshot.json")
			snapshot := NewSnapshot()
			output.Add(snapshot)
			go func() {
				http.Handle("/snapshot.json", snapshot)
			}()
		}
		if _, ok := http_features["ws-json"]; ok {
			log.Println("Websocket json enabled at ws://" + *http_addr + "/ws.json")
			websocketsender := NewWebsocketsender()
			output.Add(websocketsender)
			go func() {
				http.Handle("/ws.json", websocketsender.GetWebsocketSenderHandler())
			}()
		}
		if _, ok := http_features["sparkline"]; ok {
			log.Println("Sparkline enabled at http://" + *http_addr + "/spark.html")
			js_data := []string{"/jquery.js", "/jquery.sparkline.js", "/jquery.appear.js", "/reconnecting-websocket.js"}

			for _, n := range js_data {
				js, _ := Asset("sparkline" + n)
				http.HandleFunc(n,
					func(w http.ResponseWriter, req *http.Request) {
						w.Header().Set("Content-Type", "application/javascript")
						http.ServeContent(w, req, n, time.Time{}, bytes.NewReader(js))
					})
			}
			hb, _ := Asset("sparkline/spark.html")
			html := strings.Replace(string(hb), "@@@@", *http_addr, -1)
			http.HandleFunc("/spark",
				func(w http.ResponseWriter, req *http.Request) {
					w.Header().Set("Content-Type", "text/html")
					http.ServeContent(w, req, "spark.html", time.Time{}, strings.NewReader(html))
				})
		}
		go func() {
			info.Printf("[main] HTTP debug server running on %v", *http_addr)
			log.Println(http.ListenAndServe(*http_addr, nil))
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
