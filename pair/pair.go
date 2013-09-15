/*
	Allows clients to register with a server and sets up a ZMQ pair connection
	between the two.
*/
package pair

import "log"

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
