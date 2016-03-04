// Output to librato

package outputter

import (
	"bytes"
	"encoding/json"
	"github.com/pusher/stagger/metric"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Librato struct {
	source     string
	email      string
	token      string
	onStats    chan *metric.TimestampedStats
	httpclient *http.Client
	log        *log.Logger
}

func NewLibrato(source, email, token string, l *log.Logger) *Librato {
	x := &Librato{
		source: source,
		email:  email,
		token:  token,
		// Handle slow posts by combination of buffering channel & timing out
		onStats: make(chan *metric.TimestampedStats, 100),
		httpclient: &http.Client{
			Timeout: 2 * time.Second,
		},
		log: l,
	}
	go x.run()
	return x
}

func (l *Librato) Send(stats *metric.TimestampedStats) error {
	select {
	case l.onStats <- stats:
		return nil
	default:
		return NOT_SENT
	}
}

func (l *Librato) String() string {
	return "librato"
}

func (l *Librato) run() {
	var stats *metric.TimestampedStats
	for stats = range l.onStats {
		// Don't bother posting if there are no metrics (it's an error anyway)
		if len(stats.Counters) == 0 && len(stats.Dists) == 0 {
			// debug.Print("[librato] No stats to report")
			continue
		}

		l.post(stats)
	}
}

func (l *Librato) post(stats *metric.TimestampedStats) {
	gagues := make([]map[string]interface{}, 0)
	for key, value := range stats.Counters {
		if key.HasTags() {
			continue
		}
		gagues = append(gagues, map[string]interface{}{
			"name":  key.Name(),
			"value": value,
		})
	}

	for key, value := range stats.Dists {
		if key.HasTags() {
			continue
		}
		gagues = append(gagues, map[string]interface{}{
			"name":        key.Name(),
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
		"gauges":       gagues,
	}

	json_data, err := json.Marshal(data)
	if nil != err {
		l.log.Printf("[librato] JSON error: %v", err)
		return
	}

	req, err := http.NewRequest(
		"POST",
		"https://metrics-api.librato.com/v1/metrics",
		bytes.NewBuffer(json_data),
	)
	if nil != err {
		l.log.Printf("[librato] Error creating request: %v", err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(l.email, l.token)

	resp, err := l.httpclient.Do(req)
	if err != nil {
		l.log.Printf("[librato] HTTP error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		l.log.Printf("[librato] Invalid response: %v", resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		l.log.Printf("[librato] HTTP body: %v", string(body))
		return
	}
}
