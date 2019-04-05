package main

import (
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const namespace = "bminer"

type bminerExporter struct {
	apiUp          prometheus.Gauge
	version        *prometheus.GaugeVec
	stratumMetrics map[string]*prometheus.Desc
}

func newExporter() prometheus.Collector {
	return &bminerExporter{
		apiUp: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "up",
			Help:      "Whether a scaping error has occurred",
		}),
		version: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "version",
			Help:      "Version info of bminer",
		}, []string{"version"}),
		stratumMetrics: map[string]*prometheus.Desc{
			"shares": prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "stratum", "shares"),
				"Bminer share submissions", []string{"status"}, nil,
			),
		},
	}
}

var _ prometheus.Collector = (*bminerExporter)(nil)

func (b *bminerExporter) Describe(ch chan<- *prometheus.Desc) {
	b.apiUp.Describe(ch)
	b.version.Describe(ch)

	for _, m := range b.stratumMetrics {
		ch <- m
	}
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
		Stratum struct {
			AcceptedShares uint64 `json:"accepted_shares"`
			RejectedShares uint64 `json:"rejected_shares"`
		} `json:"stratum"`
	}{}
	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		log.Error(err)
		b.apiUp.Set(0)
	}

	ch <- prometheus.MustNewConstMetric(b.stratumMetrics["shares"], prometheus.CounterValue, float64(v.Stratum.AcceptedShares), "accepted")
	ch <- prometheus.MustNewConstMetric(b.stratumMetrics["shares"], prometheus.CounterValue, float64(v.Stratum.RejectedShares), "rejected")

	b.version.WithLabelValues(v.Version).Set(1)
	b.version.Collect(ch)

	b.apiUp.Set(1)
}
