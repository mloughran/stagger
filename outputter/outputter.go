package outputter

import (
	"errors"
	"github.com/pusher/stagger/metric"
)

type Outputter interface {
	Send(*metric.TimestampedStats) error
	String() string
}

var NOT_SENT = errors.New("Not sent")
