package main

import "log"
import "net/http"
import _ "net/http/pprof"

func main() {
	reg_chan := make(chan string)

	go StartRegistration(reg_chan)

	go StartClientManager(reg_chan)

	// Pointless, what's the best way to make the process not exit?
	log.Fatal(http.ListenAndServe(":8080", nil))
}
