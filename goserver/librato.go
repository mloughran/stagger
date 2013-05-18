// Output to librato - this is temporary

package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Librato struct {
	Source string
	Email  string
	Token  string
}

func NewLibrato(source, email, token string) *Librato {
	return &Librato{source, email, token}
}

func (l *Librato) Run(on_stats chan (*TimestampedStats)) {
	var stats *TimestampedStats
	for stats = range on_stats {
		// Don't bother posting if there are no metrics (it's an error anyway)
		if len(stats.Counters) == 0 && len(stats.Dists) == 0 {
			debug.Print("[librato] No stats to report")
			continue
		}

		l.Post(stats)
	}
}

func (l *Librato) Post(stats *TimestampedStats) {
	gagues := make([]map[string]interface{}, 0)
	for key, value := range stats.Counters {
		gagues = append(gagues, map[string]interface{}{
			"name":  key,
			"value": value,
		})
	}

	for key, value := range stats.Dists {
		gagues = append(gagues, map[string]interface{}{
			"name":        key,
			"count":       value.N,
			"sum":         value.Sum_x,
			"sum_squares": value.Sum_x2,
			"min":         value.Min,
			"max":         value.Max,
		})
	}

	data := map[string]interface{}{
		"source":       l.Source,
		"measure_time": stats.Timestamp,
		"gauges":       gagues,
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
	req.SetBasicAuth(l.Email, l.Token)

	debug.Print("[librato] POSTING")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.Printf("[librato] HTTP error: %v", err)
		return
	}

	if resp.StatusCode != 200 {
		info.Printf("[librato] Invalid status: %v", resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		info.Printf("[librato] HTTP body: %v", string(body))
		return
	}

	// Close the body as requested by the docs
	resp.Body.Close()

	debug.Print("[librato] DONE")
}
