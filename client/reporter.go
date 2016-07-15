package client

import "github.com/pusher/stagger/conn"

// A Reporter sends metrics to the stagger client.
type Reporter interface {
	// The name of the metric this reports for.
	Metric() conn.StatKey

	// Report a value.
	Report(float64)
}

// A CountReporter tracks the current value of a Count metric.
type CountReporter struct {
	client *Client
	metric conn.StatKey
}

func (c *Client) NewCountReporter(metric conn.StatKey) CountReporter {
	return CountReporter{c, metric}
}

func (r *CountReporter) Metric() conn.StatKey {
	return r.metric
}

func (r *CountReporter) Report(v float64) {
	r.client.ReportCount(r.metric, v)
}

// ReportDelta adds the delta to the current value.
func (r *CountReporter) ReportDelta(d float64) {
	r.client.ReportCountDelta(r.metric, d)
}

// A RateCounterReporter tracks the current value of a Rate Counter
// metric.
type RateCounterReporter struct {
	client *Client
	metric conn.StatKey
}

func (c *Client) NewRateCounterReporter(metric conn.StatKey) RateCounterReporter {
	return RateCounterReporter{c, metric}
}

func (r *RateCounterReporter) Metric() conn.StatKey {
	return r.metric
}

func (r *RateCounterReporter) Report(v float64) {
	r.client.ReportRateCounter(r.metric, v)
}

// ReportDelta adds the delta to the current value.
func (r *RateCounterReporter) ReportDelta(d float64) {
	r.client.ReportRateCounterDelta(r.metric, d)
}

// A DistributionReporter tracks the current value of a Distribution
// metric.
type DistributionReporter struct {
	client *Client
	metric conn.StatKey
}

func (c *Client) NewDistributionReporter(metric conn.StatKey) DistributionReporter {
	return DistributionReporter{c, metric}
}

func (r *DistributionReporter) Metric() conn.StatKey {
	return r.metric
}

func (r *DistributionReporter) Report(v float64) {
	r.client.ReportDistribution(r.metric, v)
}
