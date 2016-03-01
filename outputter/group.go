// Group receives aggregated data from the aggregator and exposes it to
// all configured outputters.

package outputter

import (
	"github.com/pusher/stagger/metric"
)

type Group struct {
	outputs []Outputter
}

func NewGroup() (g *Group) {
	return &Group{[]Outputter{}}
}

func (o *Group) Send(stats *metric.TimestampedStats) {
	for _, op := range o.outputs {
		op.Send(stats)
	}
}

func (o *Group) Add(op Outputter) {
	o.outputs = append(o.outputs, op)
}
