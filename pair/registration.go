package pair

import (
	zmq "github.com/pebbe/zmq3"
	"os"
	"time"
)

type Registration struct {
	address       string
	Registrations chan string
	sigClose      chan bool
	didClose      chan bool
}

func NewRegistration(a string) *Registration {
	return &Registration{a, make(chan string), make(chan bool), make(chan bool)}
}

func (r *Registration) Run() {
	recvMessage := make(chan ([]string))
	shouldClose := false

	// Goroutine handles all zmq interactions
	go func() {
		pull, _ := zmq.NewSocket(zmq.PULL)

		// It's necesssary to timeout so that we can close the registration
		// channel cleanly _before_ disconnecting clients. This ensures that
		// re-registrations aren't received by the dying process.
		if err := pull.SetRcvtimeo(500 * time.Millisecond); err != nil {
			info.Printf("[pair-reg] Error in SetRcvtimeo: %v", err)
			return
		}

		if err := pull.Bind(r.address); err != nil {
			info.Printf("[pair-reg] Error binding to %v: %v", r.address, err)
			os.Exit(1)
		}

		for {
			if parts, err := pull.RecvMessage(0); err == nil {
				recvMessage <- parts
			}
			if shouldClose {
				pull.Close() // blocks until closed
				r.didClose <- true
				return
			}
		}
	}()

	for {
		select {
		case parts := <-recvMessage:
			if len(parts) != 2 {
				info.Printf("[pair-reg] Invalid reg, should have 2 parts")
			} else {
				// TODO: Use the 2nd part
				r.Registrations <- parts[0]
			}
		case <-r.sigClose:
			shouldClose = true
		}
	}
}

// Shutdown is a blocking call which closes the registration channel cleanly
// (which takes up to 500ms)
func (r *Registration) Shutdown() {
	debug.Print("[pair-reg] willClose")
	r.sigClose <- true
	<-r.didClose
	debug.Print("[pair-reg] didClose")
}
