// Output to librato - this is temporary

package main

import (
	"bytes"
	"encoding/json"
	"github.com/pusher/stagger/httpclient"
	"io/ioutil"
	"net/http"
	"time"
)

type Librato struct {
	source     string
	email      string
	token      string
	on_stats   chan *TimestampedStats
	httpclient *http.Client
}

func NewLibrato(source, email, token string) *Librato {
	return &Librato{
		source: source,
		email:  email,
		token:  token,
		// Handle slow posts by combination of buffering channel & timing out
		on_stats:   make(chan *TimestampedStats, 100),
		httpclient: httpclient.NewTimeoutClient(2 * time.Second),
	}
}

func (l *Librato) Run() {
	var stats *TimestampedStats
	for stats = range l.on_stats {
		// Don't bother posting if there are no metrics (it's an error anyway)
		if len(stats.Counters) == 0 && len(stats.Dists) == 0 {
			debug.Print("[librato] No stats to report")
			continue
		}

		l.post(stats)
	}
}

func (l *Librato) Send(stats *TimestampedStats) {
	l.on_stats <- stats
}

func (l *Librato) post(stats *TimestampedStats) {
	gauges := make([]map[string]interface{}, 0)
	for key, value := range stats.Counters {
		gauges = append(gauges, map[string]interface{}{
			"name":  key,
			"value": value,
		})
	}

	for key, value := range stats.Dists {
		gauges = append(gauges, map[string]interface{}{
			"name":        key,
			"count":       value.N,
			"sum":         value.Sum_x,
			"sum_squares": value.Sum_x2,
			"min":         value.Min,
			"max":         value.Max,
		})
	}

	data := map[string]interface{}{
		"source":       l.source,
		"measure_time": stats.Timestamp,
		"gauges":       gauges,
	}

	json_data, err := json.Marshal(data)
	if nil != err {
		info.Printf("[librato] JSON error: %v", err)
		return
	}

	req, err := http.NewRequest(
		"POST",
		"https://metrics-api.librato.com/v1/metrics",
		bytes.NewBuffer(json_data),
	)
	if nil != err {
		info.Printf("[librato] Error creating request: %v", err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(l.email, l.token)

	resp, err := l.httpclient.Do(req)
	if err != nil {
		info.Printf("[librato] HTTP error: %v", err)
		return
	}

	if resp.StatusCode != 200 {
		info.Printf("[librato] Invalid response: %v", resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		info.Printf("[librato] HTTP body: %v", string(body))
		return
	}

	// Close the body as requested by the docs
	resp.Body.Close()
}
