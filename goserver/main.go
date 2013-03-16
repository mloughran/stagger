package main

import "log"
import "net/http"
import _ "net/http/pprof"

func main() {
	go StartRegistration()

	// Pointless, what's the best way to make the process not exit?
	log.Fatal(http.ListenAndServe(":8080", nil))
}