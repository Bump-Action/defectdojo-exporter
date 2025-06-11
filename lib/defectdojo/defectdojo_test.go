package defectdojo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchProducts(t *testing.T) {
	mockProducts := ProductsResponse{
		Next: "",
		Results: []Product{
			{ID: 1, Type: 2, Name: "Test Product 1"},
			{ID: 2, Type: 3, Name: "Test Product 2"},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Token dummy-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockProducts)
	}))
	defer ts.Close()

	products, err := FetchProducts(ts.URL, "dummy-token")
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
