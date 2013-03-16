package main

// import "bytes"
import "log"
import "strings"
import "net/http"
import _ "net/http/pprof"

import msgpack "github.com/ugorji/go-msgpack"
import zmq "github.com/pebbe/zmq3"

type RegMsg struct {
	Address string
}

func main() {
	// context, _ := zmq.NewContext()
	
	// Registration
	go func() {
		pull, _ := zmq.NewSocket(zmq.PULL)
		defer pull.Close()
		pull.Bind("tcp://127.0.0.1:2900")
		log.Print("Have bound")
		
		for {
			msg, _ := pull.Recv(0)
			log.Print("Have received")
			// Convert into something with a Read method, bit horrible
			// buf := bytes.NewBuffer(msg)
			buf := strings.NewReader(msg)
			
			// Decode the msgpack data into Response struct
			var resp RegMsg
			dec := msgpack.NewDecoder(buf, nil)
			if err := dec.Decode(&resp); err != nil {
				log.Print(err)
				return
			}
			
			log.Print(resp.Address)
		}
	}()
	
	// Pointless, what's the best way to make the process not exit?
	log.Fatal(http.ListenAndServe(":8080", nil))
}