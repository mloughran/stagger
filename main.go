package main

import (
	"./pair"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	//        "net"
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
	http_addr := flag.String("http", "127.0.0.1:8990", "HTTP debugging address (e.g. ':8990')")
	https_addr := flag.String("https", "0.0.0.0:8443", "HTTPS address (e.g. ':8443')")
	http_features_string := flag.String("features", "ws-json,http-json,sparkline", "HTTP features (ws-json,http-json,sparkline)")
	ssl_crt := flag.String("crt", "", "SSL Certificate")
	ssl_key := flag.String("key", "", "SSL Key")
	ssl_ca := flag.String("ca", "", "Client certificate CA (If left unspecified, client certificates are not used)")
	var showBuild = flag.Bool("build", false, "Print build information")
	flag.Parse()
	http_features := make(map[string]bool)
	for _, s := range strings.Split(*http_features_string, ",") {
		http_features[s] = true
	}
	if *ssl_crt == "" || *ssl_key == "" {
		*https_addr = ""
		info.Printf("[main] Either SSL crt (%v) and key (%v) left unspecified, disabling https interface", *ssl_crt, *ssl_key)
	}
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
	if *http_addr != "" || *https_addr != "" {
		if _, ok := http_features["http-json"]; ok {
			log.Println("[main] Http json enabled at http://" + *http_addr + "/snapshot.json")
			snapshot := NewSnapshot()
			output.Add(snapshot)
			go func() {
				http.Handle("/snapshot.json", snapshot)
			}()
		}
		if _, ok := http_features["ws-json"]; ok {
			log.Println("[main] Websocket json enabled at ws://" + *http_addr + "/ws.json")
			websocketsender := NewWebsocketsender()
			output.Add(websocketsender)
			go func() {
				http.Handle("/ws.json", websocketsender.GetWebsocketSenderHandler())
			}()
		}
		if _, ok := http_features["sparkline"]; ok {
			log.Println("[main] Sparkline enabled at http://" + *http_addr + "/spark.html")
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
			http.HandleFunc("/spark",
				func(w http.ResponseWriter, req *http.Request) {
					w.Header().Set("Content-Type", "text/html")
					http.ServeContent(w, req, "spark", time.Time{}, bytes.NewReader(hb))
				})
			http.HandleFunc("/",
				func(w http.ResponseWriter, req *http.Request) {
					if req.URL.Path != "/" {
						http.NotFound(w, req)
						return
					}
					http.Redirect(w, req, "/spark", http.StatusFound)
				})

		}
		if *http_addr != "" {
			go func() {
				info.Printf("[main] HTTP server running on %v", *http_addr)
				log.Println(http.ListenAndServe(*http_addr, nil))
			}()
		}
		if *https_addr != "" {
			cert, err := tls.LoadX509KeyPair(*ssl_crt, *ssl_key)
			if err != nil {
				log.Fatalf("[main] loading key/crt pair: %s", err)
			}
			cp := x509.NewCertPool()
			var request_cert tls.ClientAuthType
			if *ssl_ca != "" {
				ca_data, err := ioutil.ReadFile(*ssl_ca)
				if err != nil {
					log.Fatalf("[main] loading ca: %s", err)
				}
				ca_decoded, _ := pem.Decode(ca_data)
				x509cert, err := x509.ParseCertificate(ca_decoded.Bytes)
				if err != nil {
					log.Fatalf("[main] ca not x509?, %s", err)
				}
				cp.AddCert(x509cert)
				request_cert = tls.RequireAndVerifyClientCert
			} else {
				request_cert = tls.NoClientCert
				log.Println("[main] WARNING: No client certificate specified, disabling authentication")
			}
			config := tls.Config{
				Certificates:       []tls.Certificate{cert},
				ClientAuth:         request_cert,
				RootCAs:            cp,
				ClientCAs:          cp,
				InsecureSkipVerify: true, //Don't check hostname of client certificate
				NextProtos:         []string{"http/1.1"},
				CipherSuites: []uint16{ // Work around chrome ssl certificate issue
					tls.TLS_RSA_WITH_RC4_128_SHA,
					tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
					tls.TLS_RSA_WITH_AES_128_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				},
			}
			srv := http.Server{
				Addr:      *https_addr,
				TLSConfig: &config,
			}
			go func() {
				info.Printf("[main] HTTPS server running on %v", *https_addr)
				log.Println(srv.ListenAndServeTLS(*ssl_crt, *ssl_key))
			}()
		}
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
