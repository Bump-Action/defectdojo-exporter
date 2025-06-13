package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMainFunctionGET(t *testing.T) {

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, err := fmt.Fprint(w, "<h2>DefectDojo Exporter</h2>")
		if err != nil {
			t.Errorf("unexpected error writing to response: %v", err)
		}
		_, err = fmt.Fprint(w, "<p><a href='/metrics'>/metrics</a> -  available service metrics</p>")
		if err != nil {
			t.Errorf("unexpected error writing to response: %v", err)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	resp := w.Result()

	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Exptected status 200 OK, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("Expected Content-Type 'text/html', got %s", contentType)
	}
}

func TestMainFunctionPOST(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, err := fmt.Fprint(w, "<h2>DefectDojo Exporter</h2>")
		if err != nil {
			t.Errorf("unexpected error writing to response: %v", err)
		}
		_, err = fmt.Fprint(w, "<p><a href='/metrics'>/metrics</a> -  available service metrics</p>")
		if err != nil {
			t.Errorf("unexpected error writing to response: %v", err)
		}
	})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	resp := w.Result()

	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Exptected status 405 Method Not Allowed, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		t.Errorf("Expected Content-Type 'text/plain', got %s", contentType)
	}
}
