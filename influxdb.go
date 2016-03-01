package main

import (
	influxdb "github.com/influxdb/influxdb/client"
	"github.com/pusher/stagger/metric"
	uri "net/url"
	"path"
	"time"
)

type InfluxDB struct {
	tags     map[string]string
	client   *influxdb.Client
	db       string
	on_stats chan *metric.TimestampedStats
}

func NewInfluxDB(tags map[string]string, rawurl string) (client *InfluxDB, err error) {
	url, err := uri.Parse(rawurl)
	if err != nil {
		return
	}

	db := path.Base(url.Path)

	// Client config
	config := influxdb.NewConfig()
	config.Timeout = 2 * time.Second
	config.URL = *url

	if url.User != nil {
		config.Username = url.User.Username()
		config.Password, _ = url.User.Password()
	}

	debug.Printf("[influxdb] Config: %v", config)

	realclient, err := influxdb.NewClient(config)
	if err != nil {
		return
	}

	client = &InfluxDB{tags, realclient, db, make(chan *metric.TimestampedStats, 100)}
	return
}

func (x *InfluxDB) Run() {
	var stats *metric.TimestampedStats
	for stats = range x.on_stats {
		// Don't bother posting if there are no metrics (it's an error anyway)
		if len(stats.Counters) == 0 && len(stats.Dists) == 0 {
			debug.Print("[influxdb] No stats to report")
			continue
		}

		x.post(stats)
	}
}

func (x *InfluxDB) Send(stats *metric.TimestampedStats) {
	x.on_stats <- stats
}

func (x *InfluxDB) post(stats *metric.TimestampedStats) {
	var err error

	points := make([]influxdb.Point, len(stats.Counters)+len(stats.Dists))
	index := 0
	now := stats.Timestamp

	for key, value := range stats.Counters {
		points[index] = influxdb.Point{
			Measurement: key.Name(),
			Tags:        mergeTags(x.tags, key.Tags()),
			Time:        time.Unix(now, 0),
			Fields:      map[string]interface{}{"value": value},
			Precision:   "s",
		}
		index += 1
	}

	for key, value := range stats.Dists {
		points[index] = influxdb.Point{
			Measurement: key.Name(),
			Tags:        mergeTags(x.tags, key.Tags()),
			Time:        time.Unix(now, 0),
			Fields: map[string]interface{}{
				"value":       value.N,
				"sum":         value.Sum_x,
				"sum-squares": value.Sum_x2,
				"min":         value.Min,
				"max":         value.Max,
			},
			Precision: "s",
		}
		index += 1
	}

	debug.Printf("[influxdb] sending %d points for %d", len(points), now)

	bps := influxdb.BatchPoints{
		Points:          points,
		Database:        x.db,
		RetentionPolicy: "default",
	}

	ret, err := x.client.Write(bps)
	if err != nil {
		info.Printf("[influxdb] %v", err)
	} else {
		info.Printf("[influxdb] %v", ret)
	}
}

func mergeTags(t1 map[string]string, t2 map[string]string) map[string]string {
	var ret map[string]string
	if t1 != nil {
		ret = make(map[string]string, len(t1))
		for k, v := range t1 {
			ret[k] = v
		}
	}
	if t2 != nil {
		if ret == nil {
			ret = make(map[string]string)
		}
		for k, v := range t2 {
			ret[k] = v
		}
	}
	return ret
}
