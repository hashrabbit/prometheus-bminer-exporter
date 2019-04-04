package main

import (
	"net/http"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

var (
	// TODO(ivy): reserve port on Prometheus dev wiki
	listenAddress = kingpin.Flag("web.listen-address",
		"Address on which to expose metrics and web interface.",
	).Default(":9999").String()
	metricsPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	bminerScrapeURI = kingpin.Flag(
		"bminer.scrape-uri",
		"URI on which to scrape Bminer metrics.",
	).Default("http://localhost:1880").String()
)

func init() {
	prometheus.MustRegister(newExporter())
	prometheus.MustRegister(version.NewCollector("bminer_exporter"))
}

func main() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("bminer_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting bminer exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head>
            <title>Bminer Exporter</title>
            </head>
            <body>
            <h1>Bminer Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
            </body>
            </html>`))
	})

	log.Infof("Listening on %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
