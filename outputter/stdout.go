package outputter

import (
	"fmt"
	"github.com/pusher/stagger/metric"
	"io"
	"os"
	"sort"
)

type Writer struct {
	io.Writer
	name    string
	onStats chan *metric.TimestampedStats
}

func NewWriter(w io.Writer, name string) (x *Writer) {
	x = &Writer{
		Writer:  w,
		name:    name,
		onStats: make(chan *metric.TimestampedStats, 100),
	}
	go x.run()
	return x
}

func NewStdout() *Writer {
	return NewWriter(os.Stdout, "stdout")
}

func (w *Writer) Send(stats *metric.TimestampedStats) error {
	select {
	case w.onStats <- stats:
		return nil
	default:
		return NOT_SENT
	}
}

func (w *Writer) String() string {
	return w.name
}

func (w *Writer) run() {
	for stats := range w.onStats {
		var heading = fmt.Sprintf("[output] (ts:%v) Aggregated data:\n", stats.Timestamp)

		var output []string
		for key, value := range stats.Counters {
			output = append(output, fmt.Sprintf("%v: %.5g\n", key, value))
		}
		for key, value := range stats.Dists {
			output = append(output, fmt.Sprintf("%v: %v\n", key, value))
		}
		sort.Strings(output)
		w.Write([]byte(heading))
		for _, line := range output {
			w.Write([]byte(line))
		}
	}
}
