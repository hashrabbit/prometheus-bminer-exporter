package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const namespace = "bminer"

type bminerExporter struct {
	apiUp          prometheus.Gauge
	version        *prometheus.GaugeVec
	uptime         *prometheus.Desc
	deviceMetrics  map[string]*prometheus.Desc
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
		uptime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "uptime_seconds"),
			"Uptime of Bminer in seconds", nil, nil,
		),
		deviceMetrics: map[string]*prometheus.Desc{
			"solution_rate": prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "device", "solutions_per_second"),
				"Bminer solutions per second", []string{"index", "algorithm"}, nil,
			),
		},
		stratumMetrics: map[string]*prometheus.Desc{
			"shares": prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "stratum", "shares"),
				"Bminer share submissions", []string{"status"}, nil,
			),
			"share_rate": prometheus.NewDesc(
				prometheus.BuildFQName(namespace, "stratum", "share_rate"),
				"Bminer share submission rate", []string{"status"}, nil,
			),
		},
	}
}

var _ prometheus.Collector = (*bminerExporter)(nil)

func (b *bminerExporter) Describe(ch chan<- *prometheus.Desc) {
	b.apiUp.Describe(ch)
	b.version.Describe(ch)

	ch <- b.uptime

	for _, m := range b.deviceMetrics {
		ch <- m
	}

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
		Algorithm string `json:"algorithm"`
		Stratum   struct {
			AcceptedShares    uint64  `json:"accepted_shares"`
			RejectedShares    uint64  `json:"rejected_shares"`
			AcceptedShareRate float64 `json:"accepted_share_rate"`
			RejectedShareRate float64 `json:"rejected_share_rate"`
		} `json:"stratum"`
		Miners map[string]struct {
			Solver struct {
				SolutionRate float64 `json:"solution_rate"`
			} `json:"solver"`
		} `json:"miners"`
		Version   string `json:"version"`
		StartTime int64  `json:"start_time"`
	}{}

	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		log.Error(err)
		b.apiUp.Set(0)
	}

	ch <- prometheus.MustNewConstMetric(b.stratumMetrics["shares"], prometheus.CounterValue, float64(v.Stratum.AcceptedShares), "accepted")
	ch <- prometheus.MustNewConstMetric(b.stratumMetrics["shares"], prometheus.CounterValue, float64(v.Stratum.RejectedShares), "rejected")
	ch <- prometheus.MustNewConstMetric(b.stratumMetrics["share_rate"], prometheus.CounterValue, v.Stratum.AcceptedShareRate, "accepted")
	ch <- prometheus.MustNewConstMetric(b.stratumMetrics["share_rate"], prometheus.CounterValue, v.Stratum.RejectedShareRate, "rejected")

	for deviceIndex, miner := range v.Miners {
		ch <- prometheus.MustNewConstMetric(b.deviceMetrics["solution_rate"], prometheus.GaugeValue, miner.Solver.SolutionRate, deviceIndex, v.Algorithm)
	}

	b.version.WithLabelValues(v.Version).Set(1)
	b.version.Collect(ch)

	ch <- prometheus.MustNewConstMetric(b.uptime, prometheus.CounterValue, float64(time.Now().Unix()-v.StartTime))

	b.apiUp.Set(1)
}
