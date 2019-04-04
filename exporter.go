package main

import (
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type bminerExporter struct {
	version *prometheus.GaugeVec
	apiUp   prometheus.Gauge
}

func newExporter() prometheus.Collector {
	return &bminerExporter{
		version: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "bminer",
			Name:      "version",
			Help:      "Version info of bminer",
		}, []string{"version"}),
		apiUp: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "bminer",
			Subsystem: "api",
			Name:      "up",
			Help:      "Whether a scaping error has occurred",
		}),
	}
}

var _ prometheus.Collector = (*bminerExporter)(nil)

func (b *bminerExporter) Describe(ch chan<- *prometheus.Desc) {
	b.version.Describe(ch)
	ch <- b.apiUp.Desc()
}

func (b *bminerExporter) Collect(ch chan<- prometheus.Metric) {
	res, err := http.Get(*bminerScrapeURI + "/api/status")
	if err != nil {
		log.Error(err)
		b.apiUp.Set(0)
		return
	}
	defer res.Body.Close()

	v := struct {
		Version string `json:"version"`
	}{}
	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		log.Error(err)
		b.apiUp.Set(0)
	}

	b.version.WithLabelValues(v.Version).Set(1)
	b.version.Collect(ch)

	b.apiUp.Set(1)
}
