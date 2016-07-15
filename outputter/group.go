// Group receives aggregated data from the aggregator and exposes it to
// all configured outputters.

package outputter

import (
	"github.com/pusher/stagger/metric"
	"log"
)

type Group struct {
	outputs []Outputter
}

func NewGroup() (g *Group) {
	return &Group{[]Outputter{}}
}

func (o *Group) Send(stats *metric.TimestampedStats) (err error) {
	for _, o := range o.outputs {
		err2 := o.Send(stats)
		if err2 != nil {
			log.Println(o.String(), err2)
			err = err2
		}
	}
	return
}

func (o *Group) Add(op Outputter) {
	o.outputs = append(o.outputs, op)
}
