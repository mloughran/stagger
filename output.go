// Output receives aggregated data from the aggregator and exposes it to
// all configured outputters.

package main

type OutputStat struct {
	Timestamp int64
	Dist      *Dist
}

type Outputter interface {
	Send(*TimestampedStats)
}

type Output struct {
	outputs []Outputter
}

func NewOutput() *Output {
	return &Output{[]Outputter{}}
}

func (o *Output) Run(complete_chan <-chan (*TimestampedStats)) {
	// Forward all stats to outputters
	for stats := range complete_chan {
		for _, op := range o.outputs {
			op.Send(stats)
		}
	}
}

func (o *Output) Add(op Outputter) {
	o.outputs = append(o.outputs, op)
}
