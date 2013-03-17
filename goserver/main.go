package main

import "log"
import "net/http"
import _ "net/http/pprof"

func main() {
	reg_chan := make(chan *Client)

	go StartRegistration(reg_chan)

	go func() {
		for {
			client := <- reg_chan
			log.Print(client.Address)
		}
	}()

	// Pointless, what's the best way to make the process not exit?
	log.Fatal(http.ListenAndServe(":8080", nil))
}
