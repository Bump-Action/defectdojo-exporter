package collector

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/iamhalje/defectdojo-exporter/lib/defectdojo"
	"github.com/prometheus/client_golang/prometheus"
)

// CollectMetrics main collector
func CollectMetrics(link, token string, concurrency int, interval time.Duration, timeout time.Duration, useEngagementUpdate bool) {
	limiter := make(chan struct{}, concurrency)

	for {
		products, err := defectdojo.FetchProducts(link, token, timeout)
		if err != nil {
			log.Printf("Error fetching products: %v", err)
			return
		}

		var wg sync.WaitGroup

		for _, p := range products {
			wg.Add(1)
			limiter <- struct{}{}

			go func(product string, productID int, productTypeID int) {
				defer wg.Done()
				defer func() { <-limiter }()

				if useEngagementUpdate {
					latestEngagementUpdate, err := defectdojo.FetchEngagementUpdatedTimestamp(productID, link, token, timeout)
					if err != nil {
						log.Printf("Error fetching engagement update time for product %s: %v", product, err)
						return
					}

					defectdojo.MU.Lock()
					prevUpdate, exists := defectdojo.PrevEngagementUpdateTimes[product]
					if exists && !latestEngagementUpdate.After(prevUpdate) {
						defectdojo.MU.Unlock()
						return
					}
					defectdojo.PrevEngagementUpdateTimes[product] = latestEngagementUpdate
					defectdojo.MU.Unlock()
				}

				productType, err := defectdojo.FetchProductType(productTypeID, link, token, timeout)
				if err != nil {
					log.Printf("Error fetching product type for product %s: %v", product, err)
					return
				}

				vulnerabilities, err := defectdojo.FetchVulnerabilities(product, link, token, timeout)
				if err != nil {
					log.Printf("Error fetching vulnerabilities for product %s: %v", product, err)
					return
				}

				type statusCountMap map[string]map[string]float64 // severity -> cwe -> count
				statusMaps := map[string]statusCountMap{
					"active":         {},
					"duplicate":      {},
					"under_review":   {},
					"false_positive": {},
					"out_of_scope":   {},
					"risk_accepted":  {},
					"verified":       {},
					"mitigated":      {},
				}

				// Aggregate the number of vulnerabilities by severity and CWE
				for _, vuln := range vulnerabilities {
					severity := strings.ToLower(vuln.Severity)

					if vuln.Active {
						increment(statusMaps["active"], severity, fmt.Sprintf("%d", vuln.CWE))
					}
					if vuln.Duplicate {
						increment(statusMaps["duplicate"], severity, fmt.Sprintf("%d", vuln.CWE))
					}
					if vuln.UnderReview {
						increment(statusMaps["under_review"], severity, fmt.Sprintf("%d", vuln.CWE))
					}
					if vuln.FalseP {
						increment(statusMaps["false_positive"], severity, fmt.Sprintf("%d", vuln.CWE))
					}
					if vuln.OutOfScope {
						increment(statusMaps["out_of_scope"], severity, fmt.Sprintf("%d", vuln.CWE))
					}
					if vuln.RiskAccepted {
						increment(statusMaps["risk_accepted"], severity, fmt.Sprintf("%d", vuln.CWE))
					}
					if vuln.Verified {
						increment(statusMaps["verified"], severity, fmt.Sprintf("%d", vuln.CWE))
					}
					if vuln.Mitigated {
						increment(statusMaps["mitigated"], severity, fmt.Sprintf("%d", vuln.CWE))
					}
				}

				update := func(metric *prometheus.GaugeVec, prevMap map[string]map[string]float64, current statusCountMap) {
					defectdojo.MU.Lock()
					defer defectdojo.MU.Unlock()

					if prevMap[product] == nil {
						prevMap[product] = make(map[string]float64)
					}

					// Mark seen entries
					seen := make(map[string]bool)

					for severity, cweMap := range current {
						for cwe, count := range cweMap {
							labels := []string{product, productType, severity, cwe}
							metric.WithLabelValues(labels...).Set(count)
							prevMap[product][fmt.Sprintf("%s|%s", severity, cwe)] = count
							seen[fmt.Sprintf("%s|%s", severity, cwe)] = true
						}
					}

					// Set to 0 those that were present before but now not seen
					for key, prevVal := range prevMap[product] {
						if !seen[key] && prevVal != 0 {
							parts := strings.Split(key, "|")
							if len(parts) != 2 {
								continue
							}
							severity, cwe := parts[0], parts[1]
							labels := []string{product, productType, severity, cwe}
							metric.WithLabelValues(labels...).Set(0)
							prevMap[product][key] = 0
						}
					}
				}

				update(defectdojo.VulnActiveGauge, defectdojo.PrevActive, statusMaps["active"])
				update(defectdojo.VulnDuplicateGauge, defectdojo.PrevDuplicate, statusMaps["duplicate"])
				update(defectdojo.VulnUnderReviewGauge, defectdojo.PrevUnderReview, statusMaps["under_review"])
				update(defectdojo.VulnFalsePositiveGauge, defectdojo.PrevFalsePositive, statusMaps["false_positive"])
				update(defectdojo.VulnOutOfScopeGauge, defectdojo.PrevOutOfScope, statusMaps["out_of_scope"])
				update(defectdojo.VulnRiskAcceptedGauge, defectdojo.PrevRiskAccepted, statusMaps["risk_accepted"])
				update(defectdojo.VulnVerifiedGauge, defectdojo.PrevVerified, statusMaps["verified"])
				update(defectdojo.VulnMitigatedGauge, defectdojo.PrevMitigated, statusMaps["mitigated"])

			}(p.Name, p.ID, p.Type)
		}
		wg.Wait()
		time.Sleep(interval)
	}
}

func increment(m map[string]map[string]float64, severity, cwe string) {
	if m[severity] == nil {
		m[severity] = make(map[string]float64)
	}
	m[severity][cwe]++
}
