package defectdojo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestFetchProducts(t *testing.T) {
	mockProducts := ProductsResponse{
		Next: "",
		Results: []Product{
			{
				ID:   1,
				Type: 2,
				Name: "Test Product 1",
			},
			{
				ID:   2,
				Type: 3,
				Name: "Test Product 2",
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Token dummy-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockProducts); err != nil {
			t.Fatalf("failed to encode mockProducts: %v", err)
		}
	}))
	defer ts.Close()

	products, err := FetchProducts(ts.URL, "dummy-token", 30*time.Second)
	if err != nil {
		t.Fatalf("FetchProducts error: %v", err)
	}

	if len(products) != len(mockProducts.Results) {
		t.Errorf("Expected %d products, got %d", len(mockProducts.Results), len(products))
	}

	for i, product := range products {
		expected := mockProducts.Results[i]
		if product.ID != expected.ID || product.Name != expected.Name || product.Type != expected.Type {
			t.Errorf("Mismatch at product %d: got %+v, want %+v", i, product, expected)
		}
	}
}

func TestFetchFindings(t *testing.T) {
	mockFindings := FindingsResponse{
		Next: "",
		Results: []Finding{
			{
				Active:       true,
				Severity:     "critical",
				CWE:          101,
				FalseP:       false,
				Duplicate:    false,
				OutOfScope:   false,
				RiskAccepted: true,
				UnderReview:  true,
				Verified:     true,
				Mitigated:    true,
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Token dummy-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockFindings); err != nil {
			t.Errorf("failed to encode mockFindings: %v", err)
		}
	}))
	defer ts.Close()

	findings, err := FetchVulnerabilities("Test Product", ts.URL, "dummy-token", 30*time.Second)
	if err != nil {
		t.Fatalf("FetchFindings error: %v", err)
	}

	if len(findings) != len(mockFindings.Results) {
		t.Errorf("Unexpected %d findings, got %d", len(mockFindings.Results), len(findings))
	}

	for i, finding := range findings {
		expected := mockFindings.Results[i]
		if !reflect.DeepEqual(finding, expected) {
			t.Errorf("Finding %d mismatch:\ngot %+v\nwant %+v", i, finding, expected)
		}
	}
}

func TestFetchProductType(t *testing.T) {
	mockProductType := TypeResponse{
		Results: []Type{
			{
				Name: "python",
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Token dummy-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockProductType); err != nil {
			t.Errorf("failed to encode mockProductType: %v", err)
		}
	}))
	defer ts.Close()

	types, err := FetchProductType(1, ts.URL, "dummy-token", 30*time.Second)
	if err != nil {
		t.Fatalf("FetchProductType error: %v", err)
	}

	if len(mockProductType.Results) == 0 {
		t.Errorf("Unexpected %d types, got %d", len(mockProductType.Results), len(types))
	}
}

func TestFetchEngagementUpdatedTimestamp(t *testing.T) {

	time1 := time.Date(2025, 06, 13, 11, 16, 13, 913679251, time.FixedZone("UTC+5", 5*60*60))
	time2 := time.Date(2025, 06, 13, 0, 0, 0, 0, time.UTC)

	mockEngagementUpdatedTimestamp := EngagementsResponse{
		Next: "",
		Results: []Engagement{
			{
				ID:      1,
				Product: 1,
				Updated: time1,
			},
			{
				ID:      2,
				Product: 1,
				Updated: time2,
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Token dummy-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(mockEngagementUpdatedTimestamp); err != nil {
			t.Errorf("failed to encode mockEngagementUpdatedTimestamp: %v", err)
		}
	}))
	defer ts.Close()

	latest, err := FetchEngagementUpdatedTimestamp(1, ts.URL, "dummy-token", 30*time.Second)
	if err != nil {
		t.Fatalf("FetchEngagementUpdatedTimestamp error: %v", err)
	}

	if !latest.Equal(time1) {
		t.Errorf("Expected latest timestamp %v, got %v", time1, latest)
	}
}
