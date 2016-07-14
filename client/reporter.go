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

// A CumulativeCountReporter tracks the current value of a Cumulative
// Count metric.
type CumulativeCountReporter struct {
	client *Client
	metric conn.StatKey
}

func (c *Client) NewCumulativeCountReporter(metric conn.StatKey) CumulativeCountReporter {
	return CumulativeCountReporter{c, metric}
}

func (r *CumulativeCountReporter) Metric() conn.StatKey {
	return r.metric
}

func (r *CumulativeCountReporter) Report(v float64) {
	r.client.ReportCumulativeCount(r.metric, v)
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
