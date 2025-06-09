package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/envflag"
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
	envflag.Parse()
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, err := fmt.Fprint(w, "<h2>DefectDojo Exporter</h2>")
		if err != nil {
			log.Fatalf("Error writing response: %v", err)
		}
		_, err = fmt.Fprintf(w, "<p><a href='/metrics'>/metrics</a> -  available service metrics</p>")
		if err != nil {
			log.Fatalf("Error writing reponse: %v", err)
		}
	})

	http.Handle("/metrics", promhttp.Handler())

	log.Printf("Starting server on :%d", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	log.Fatalf("Problem starting HTTP server: %v", err)
}
