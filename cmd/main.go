package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iamhalje/defectdojo-exporter/lib/buildinfo"
	"github.com/iamhalje/defectdojo-exporter/lib/collector"
	"github.com/iamhalje/defectdojo-exporter/lib/defectdojo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ddURL       = flag.String("DD_URL", "", "Base URL of the DefectDojo API (e.g. https://defectdojo.example.com)")
	ddToken     = flag.String("DD_TOKEN", "", "API token used for authenticating requests to DefectDojo")
	port        = flag.Int("port", 8080, "Port number where the exporter HTTP server will listen")
	concurrency = flag.Int("concurrency", 5, "Maximum number of concurrent API requests to DefectDojo")
	interval    = flag.Duration("interval", 5*time.Minute, "Sleep interval duration between metric collection cycles")
)

func main() {
	flag.Parse()
	buildinfo.Init()

	if *ddURL == "" || *ddToken == "" {
		log.Fatalf("Both DD_URL and DD_TOKEN must be set")
	}

	prometheus.MustRegister(defectdojo.VulnActiveGauge)
	prometheus.MustRegister(defectdojo.VulnDuplicateGauge)
	prometheus.MustRegister(defectdojo.VulnUnderReviewGauge)
	prometheus.MustRegister(defectdojo.VulnFalsePositiveGauge)
	prometheus.MustRegister(defectdojo.VulnOutOfScopeGauge)
	prometheus.MustRegister(defectdojo.VulnRiskAcceptedGauge)
	prometheus.MustRegister(defectdojo.VulnVerifiedGauge)
	prometheus.MustRegister(defectdojo.VulnMitigatedGauge)

	go collector.CollectMetrics(*ddURL, *ddToken, *concurrency, *interval)

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting server on :%d", *port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	log.Fatalf("Problem starting HTTP server: %v", err)
}
