package defectdojo

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	VulnActiveGauge        = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "dojo_vulnerabilities_active", Help: "Number of active vulnerabilities in DefectDojo"}, []string{"product", "product_type", "severity", "cwe"})
	VulnDuplicateGauge     = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "dojo_vulnerabilities_duplicate", Help: "Number of duplicate vulnerabilities in DefectDojo"}, []string{"product", "product_type", "severity", "cwe"})
	VulnUnderReviewGauge   = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "dojo_vulnerabilities_under_review", Help: "Number of vulnerabilities under review in DefectDojo"}, []string{"product", "product_type", "severity", "cwe"})
	VulnFalsePositiveGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "dojo_vulnerabilities_false_positive", Help: "Number of false positive vulnerabilities in DefectDojo"}, []string{"product", "product_type", "severity", "cwe"})
	VulnOutOfScopeGauge    = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "dojo_vulnerabilities_out_of_scope", Help: "Number of vulnerabilities out of scope in DefectDojo"}, []string{"product", "product_type", "severity", "cwe"})
	VulnRiskAcceptedGauge  = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "dojo_vulnerabilities_risk_accepted", Help: "Number of vulnerabilities with risk accepted in DefectDojo"}, []string{"product", "product_type", "severity", "cwe"})
	VulnVerifiedGauge      = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "dojo_vulnerabilities_verified", Help: "Number of verified vulnerabilities in DefectDojo"}, []string{"product", "product_type", "severity", "cwe"})
	VulnMitigatedGauge     = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "dojo_vulnerabilities_mitigated", Help: "Number of mitigated vulnerabilities in DefectDojo"}, []string{"product", "product_type", "severity", "cwe"})

	PrevEngagementUpdateTimes = make(map[string]time.Time)
	PrevActive                = make(map[string]map[string]float64)
	PrevDuplicate             = make(map[string]map[string]float64)
	PrevUnderReview           = make(map[string]map[string]float64)
	PrevFalsePositive         = make(map[string]map[string]float64)
	PrevOutOfScope            = make(map[string]map[string]float64)
	PrevRiskAccepted          = make(map[string]map[string]float64)
	PrevVerified              = make(map[string]map[string]float64)
	PrevMitigated             = make(map[string]map[string]float64)

	MU sync.Mutex
)

type Finding struct {
	Active       bool   `json:"active"`
	Severity     string `json:"severity"`
	CWE          int    `json:"cwe"`
	FalseP       bool   `json:"false_p"`
	Duplicate    bool   `json:"duplicate"`
	OutOfScope   bool   `json:"out_of_scope"`
	RiskAccepted bool   `json:"risk_accepted"`
	UnderReview  bool   `json:"under_review"`
	Verified     bool   `json:"verified"`
	Mitigated    bool   `json:"is_mitigated"`
}

type FindingsResponse struct {
	Next    string    `json:"next"`
	Results []Finding `json:"results"`
}

type Product struct {
	ID   int    `json:"id"`
	Type int    `json:"prod_type"`
	Name string `json:"name"`
}

type ProductsResponse struct {
	Next    string    `json:"next"`
	Results []Product `json:"results"`
}

type Engagement struct {
	ID      int       `json:"id"`
	Product int       `json:"product"`
	Updated time.Time `json:"updated"`
}

type EngagementsResponse struct {
	Next    string       `json:"next"`
	Results []Engagement `json:"results"`
}

type Type struct {
	Name string `json:"name"`
}

type TypeResponse struct {
	Results []Type `json:"results"`
}

// FetchProducts retrieves the list of products
func FetchProducts(link, token string) ([]Product, error) {
	products := []Product{}
	endpoint := fmt.Sprintf("%s/api/v2/products/", link)

	for endpoint != "" {
		resp, err := makeRequest(endpoint, token)
		if err != nil {
			log.Printf("Error fetching products: %v", err)
			return nil, err
		}
		var productsResp ProductsResponse
		if err := json.Unmarshal(resp, &productsResp); err != nil {
			log.Printf("Error unmarshalling products response: %v", err)
			return nil, err
		}

		products = append(products, productsResp.Results...)
		endpoint = productsResp.Next
	}
	return products, nil
}

// FetchVulnerabilities retrieves the list of findings
func FetchVulnerabilities(product, link, token string) ([]Finding, error) {
	vulnerabilities := []Finding{}
	endpoint := fmt.Sprintf("%s/api/v2/findings/?product_name=%s&limit=100", link, url.PathEscape(product))

	for endpoint != "" {
		resp, err := makeRequest(endpoint, token)
		if err != nil {
			log.Printf("Error fetching vulnerabilities for product %s: %v", product, err)
			return nil, err
		}
		var findingsResp FindingsResponse
		if err := json.Unmarshal(resp, &findingsResp); err != nil {
			log.Printf("Error unmarshalling vulnerabilities response for product %s: %v", product, err)
			return nil, err
		}

		vulnerabilities = append(vulnerabilities, findingsResp.Results...)
		endpoint = findingsResp.Next
	}

	return vulnerabilities, nil
}

// FetchProductType retrieves the product type
func FetchProductType(product int, link, token string) (string, error) {
	endpoint := fmt.Sprintf("%s/api/v2/product_types/?id=%d&limit=1", link, product)

	resp, err := makeRequest(endpoint, token)
	if err != nil {
		log.Printf("Error fetching product type for product %d: %v", product, err)
		return "", err
	}
	var productTypeResp TypeResponse
	if err := json.Unmarshal(resp, &productTypeResp); err != nil {
		log.Printf("Error unmarshalling product type response for product %d: %v", product, err)
		return "", err
	}

	if len(productTypeResp.Results) == 0 {
		return "", fmt.Errorf("no product type found for product %d", product)
	}

	return productTypeResp.Results[0].Name, nil
}

// FetchEngagementUpdatedTimestamp retrieves the timestamp of the most recent engagement
func FetchEngagementUpdatedTimestamp(product int, link, token string) (time.Time, error) {
	var latestUpdate time.Time
	endpoint := fmt.Sprintf("%s/api/v2/engagements/?product=%d&limit=100", link, product)

	for endpoint != "" {
		resp, err := makeRequest(endpoint, token)
		if err != nil {
			log.Printf("Error fetching engagements for product %d: %v", product, err)
			return time.Time{}, err
		}

		var engagementResp EngagementsResponse
		if err := json.Unmarshal(resp, &engagementResp); err != nil {
			log.Printf("Error unmarshalling engagements response for product %d: %v", product, err)
			return time.Time{}, err
		}

		for _, engagement := range engagementResp.Results {
			if engagement.Updated.After(latestUpdate) {
				latestUpdate = engagement.Updated
			}
		}

		endpoint = engagementResp.Next
	}

	return latestUpdate, nil
}

// CollectCWEs take all CWE in vulnerabilities
func CollectCWEs(vulnerabilities []Finding) map[int]bool {
	CWEs := make(map[int]bool)
	for _, vuln := range vulnerabilities {
		CWEs[vuln.CWE] = true
	}
	return CWEs
}

// makeRequest send request in API DefectDojo
func makeRequest(link, token string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}
