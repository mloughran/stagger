// Package client implements a Stagger client for Go programs. There
// are three different types of metric supported:
//
//   * Callbacks, which are called once every reporting period to
//     gather metric values. These are useful for
//     infrequently-changing or expensive metrics where reporting
//     every change through a Count metric would be inconvenient.
//
//   * Counts, a simple numeric value.
//
//   * Rate Counters, a variant on Count for values which are
//     changing monotonically. When statistics are requested by the
//     server, the difference between the current and prior values is
//     sent, rather than just the current value.
//
//   * Distributions, a simple numeric value, sent to the server as
//     the number of values reported, the min, the max, the sum, and
//     the sum-of-squares.
//
// There is a "callback" variant of each metric type, where the metric
// is only measured when the Stagger server requests a measurement.
// These are useful for metrics which would be inconvenient to meausre
// on change, such as garbage collection pause time.
//
// "Reporter" values are available, which can be used to report new
// values without needing to keep track of the metric name. These may
// be useful for creating a collection of reporters in one place and
// then passing them around to other functions.
//
// Here are some example metrics, and the type of Stagger metric most
// suitable:
//
//   * Allocated memory: a count callback (using
//     runtime.ReadMemStats), as reporting this on every allocation
//     would add a lot of noise to the rest of the code.
//
//   * Number of connected network clients: a count, either reporting
//     the total number of connected clients (if that information is
//     readily available), or a delta of +1 on connect and -1 on
//     disconnect.
//
//   * Change in number of connected network clients every second: a
//     rate counter, with exactly the same reporting as the number of
//     connected network clients.
//
//   * Time some operation takes: a distribution, reporting the
//     duration of each operation as a new entry.
//
//   * Garbage collection pause times: a distribution callback (using
//     runtime.ReadMemStats).
//
// Reporting is atomic, the server is never sent a partially-updated
// state.
package client

import (
	"github.com/pusher/stagger/conn"
	"github.com/pusher/stagger/encoding"
	"github.com/pusher/stagger/metric"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type (
	counts struct {
		values    map[conn.StatKey]float64
		callbacks map[conn.StatKey]func() float64
		mutex     sync.Mutex
	}

	rateCounters struct {
		values    map[conn.StatKey]rateCounter
		callbacks map[conn.StatKey]func() float64
		mutex     sync.Mutex
	}

	// Like a counter, but (current - prior) is reported.
	rateCounter struct {
		current float64
		prior   float64
	}

	distributions struct {
		values    map[conn.StatKey]metric.Dist
		callbacks map[conn.StatKey]func() float64
		mutex     sync.Mutex
	}

	Client struct {
		conn conn.Conn

		counts        counts
		rateCounters  rateCounters
		distributions distributions
	}

	Logger interface {
		// Log an error.
		Printf(string, ...interface{})
	}
)

// Stagger executes the stagger client, reconnecting on error. It
// sends a RegisterProcess message with tags "cmd"=os.Args[0],
// "pid"=strconv.Itoa(os.Getpid()).
//
// If the first connection attempt fails, this terminates with the
// error. This is to prevent an infinite loop in the situation where
// the connection address is wrong.
//
// If an error occurs after starting, and the connection needs to be
// re-established, the error is logged.
func Stagger(addr string, logger Logger) error {
	c, err := Dial(addr)

	if err != nil {
		return err
	}

	for {
		// Run the client until it terminates.
		err = c.Run()
		logger.Printf("Lost connection to Stagger: %s", err.Error())

		// Loop until reconnection is successful.
		for {
			c, err = Dial(addr)
			if err == nil {
				break
			}

			logger.Printf("Failed to reconnect to Stagger: %s", err.Error())

			// Avoid eating CPU if there is some sort of
			// network error.
			time.Sleep(time.Second)
		}
	}
}

// Dial connects to Stagger and send a RegisterProcess message with
// tags "cmd"=os.Args[0], "pid"=strconv.Itoa(os.Getpid()).
func Dial(addr string) (*Client, error) {
	c, err := DialNoRegister(addr)
	if err != nil {
		return nil, err
	}

	args := map[string]string{
		"cmd": os.Args[0],
		"pid": strconv.Itoa(os.Getpid()),
	}
	c.RegisterProcess(args)

	return c, nil
}

// DialNoRegister connects to Stagger, but does not send the
// RegisterProcess message. You will need to call
// 'Client.RegisterProcess' yourself.
func DialNoRegister(addr string) (*Client, error) {
	c1, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, err
	}

	c := Client{conn: conn.NewConn(c1, encoding.Encoding{})}

	return &c, nil
}

// RegisterProcess identifies the client to the server.
func (c *Client) RegisterProcess(tags map[string]string) error {
	return c.conn.WriteMessage(conn.RegisterProcess{Tags: tags})
}

// ReportCount reports the current value of a counter.
func (c *Client) ReportCount(k conn.StatKey, v float64) {
	c.counts.lock()
	c.counts.unsafeReport(k, v)
	c.counts.unlock()
}

// ReportCountDelta adds the delta to the current value of a counter.
// If the metric does not exist, it is created with a current value of
// 0.
func (c *Client) ReportCountDelta(k conn.StatKey, d float64) {
	c.counts.lock()
	v := c.counts.values[k]
	c.counts.unsafeReport(k, v+d)
	c.counts.unlock()
}

// ReportCounts reports a collection of counter values. This is
// atomic: if a ReportAll request arrives as this function executes,
// the server will not be given a partially-updated state.
func (c *Client) ReportCounts(counters map[conn.StatKey]float64) {
	c.counts.lock()
	for k, v := range counters {
		c.counts.unsafeReport(k, v)
	}
	c.counts.unlock()
}

// RegisterCountCallback adds a callback which will be used once every
// reporting period to report the value of a count.
func (c *Client) RegisterCountCallback(k conn.StatKey, cb func() float64) {
	c.counts.lock()
	c.counts.unsafeRegister(k, cb)
	c.counts.unlock()
}

// ReportRateCounter reports the current value of a rate counter. When
// statistics are sent to the server, the value reported is (current -
// prior), to allow reporting metrics which change monotonically over
// time.
func (c *Client) ReportRateCounter(k conn.StatKey, v float64) {
	c.rateCounters.lock()
	c.rateCounters.unsafeReport(k, v)
	c.rateCounters.unlock()
}

// ReportRateCounterDelta adds the delta to the current value of a
// rate counter. If the rate counter does not exist, it is created
// with a current (and prior) value of 0.
func (c *Client) ReportRateCounterDelta(k conn.StatKey, d float64) {
	c.rateCounters.lock()
	v := c.rateCounters.values[k]
	c.rateCounters.unsafeReport(k, v.current+d)
	c.rateCounters.unlock()
}

// ReportRateCounters reports a collection of rate counter values.
// Like 'ReportCounts', this is atomic.
func (c *Client) ReportRateCounters(counters map[conn.StatKey]float64) {
	c.rateCounters.lock()
	for k, v := range counters {
		c.rateCounters.unsafeReport(k, v)
	}
	c.rateCounters.unlock()
}

// RegisterRateCounterCallback adds a callback which will be used once
// every reporting period to report the current value of a rate
// counter. The initial prior value is 0.
func (c *Client) RegisterRateCounterCallback(k conn.StatKey, cb func() float64) {
	c.rateCounters.lock()
	c.rateCounters.unsafeRegister(k, cb)
	c.rateCounters.unlock()
}

// ReportDistribution adds a measurement to a distribution. If the
// distribution did not exist, it is created with this being the first
// measurement.
func (c *Client) ReportDistribution(k conn.StatKey, v float64) {
	c.distributions.lock()
	c.distributions.unsafeReport(k, v)
	c.distributions.unlock()
}

// ReportDistributions adds a collection of distribution values. Like
// 'ReportCounts', this is atomic.
func (c *Client) ReportDistributions(distributions map[conn.StatKey][]float64) {
	c.distributions.lock()
	for k, vs := range distributions {
		for _, v := range vs {
			c.distributions.unsafeReport(k, v)
		}
	}
	c.distributions.unlock()
}

// Run executes the Stagger client, responding to pings and requests
// for reports. This closes the connection on error.
func (c *Client) Run() error {
	defer c.conn.Close()

	for {
		msg, err := c.conn.ReadMessage()
		if err != nil {
			return err
		}

		switch msg.(type) {
		case *conn.PairPing:
			if err := c.conn.WriteMessage(conn.PairPong{}); err != nil {
				return err
			}
		case *conn.ReportAll:
			stats := c.report()

			if err := c.conn.WriteMessage(conn.StatsComplete{Stats: stats}); err != nil {
				return err
			}
		}
	}
}

/// INTERNAL

func (cs *counts) lock() {
	cs.mutex.Lock()
}

func (cs *counts) unlock() {
	cs.mutex.Unlock()
}

// Not safe for concurrent use!
func (cs *counts) unsafeReport(k conn.StatKey, v float64) {
	cs.values[k] = v
}

// Not safe for concurrent use!
func (cs *counts) unsafeRegister(k conn.StatKey, cb func() float64) {
	cs.callbacks[k] = cb
}

func (rs *rateCounters) lock() {
	rs.mutex.Lock()
}

func (rs *rateCounters) unlock() {
	rs.mutex.Unlock()
}

// Not safe for concurrent use!
func (rs *rateCounters) unsafeReport(k conn.StatKey, v float64) {
	r := rs.values[k]
	r.current = v
	rs.values[k] = r
}

// Not safe for concurrent use!
func (rs *rateCounters) unsafeRegister(k conn.StatKey, cb func() float64) {
	rs.callbacks[k] = cb
}

func (ds *distributions) lock() {
	ds.mutex.Lock()
}

func (ds *distributions) unlock() {
	ds.mutex.Unlock()
}

// Not safe for concurrent use!
func (ds *distributions) unsafeReport(k conn.StatKey, v float64) {
	d, ok := ds.values[k]

	if ok {
		d.AddEntry(v)
	} else {
		d = *metric.NewDistFromValue(v)
	}

	ds.values[k] = d
}

// Not safe for concurrent use!
func (ds *distributions) unsafeRegister(k conn.StatKey, cb func() float64) {
	ds.callbacks[k] = cb
}

// Returns the current statistics, is atomic.
func (c *Client) report() conn.Stats {
	// Prevent concurrent modification of the stats as we report.
	c.counts.lock()
	c.rateCounters.lock()
	c.distributions.lock()
	defer c.counts.unlock()
	defer c.rateCounters.unlock()
	defer c.distributions.unlock()

	// Run all the callbacks, to populate the values.
	for k, vf := range c.counts.callbacks {
		c.counts.values[k] = vf()
	}
	for k, vf := range c.rateCounters.callbacks {
		rateCounter := c.rateCounters.values[k]
		rateCounter.prior = rateCounter.current
		rateCounter.current = vf()
	}
	for k, vf := range c.distributions.callbacks {
		value := vf()
		dist, ok := c.distributions.values[k]

		if ok {
			dist.AddEntry(value)
		} else {
			dist = *metric.NewDistFromValue(value)
		}

		c.distributions.values[k] = dist
	}

	// Then construct a conn.Stats struct with all the metric values in.
	var (
		numCounts     = len(c.counts.values)
		numRateCounts = len(c.rateCounters.values)
		numDists      = len(c.distributions.values)
	)

	stats := conn.Stats{
		Timestamp: time.Now().Unix(),
		Counts:    make([]conn.StatCount, numCounts+numRateCounts),
		Dists:     make([]conn.StatDist, numDists),
	}

	i := 0
	for k, v := range c.counts.values {
		stats.Counts[i] = conn.StatCount{Name: k.String(), Count: v}
		i++
	}

	for k, v := range c.rateCounters.values {
		stats.Counts[i] = conn.StatCount{Name: k.String(), Count: v.current - v.prior}
		v.prior = v.current
		i++
	}

	i = 0
	for k, v := range c.distributions.values {
		stats.Dists[i] = conn.StatDist{Name: k.String(), Dist: [5]float64{v.N, v.Min, v.Max, v.Sum_x, v.Sum_x2}}
		i++
	}

	return stats
}
