package outputter

import (
	"github.com/pusher/stagger/metric"
)

type Outputter interface {
	Send(*metric.TimestampedStats)
}
