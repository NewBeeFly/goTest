package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHelloHandler(t *testing.T) {
	tests := []struct {
		name         string
		queryParam   string
		expectStatus int
		expectBody   string
	}{
		{"default greeting", "", http.StatusOK, "Hello, World!"},
		{"named greeting", "?name=Go", http.StatusOK, "Hello, Go!"},
		{"empty name falls back", "?name=", http.StatusOK, "Hello, World!"},
		{"xss escaped", "?name=<script>", http.StatusOK, "Hello, &lt;script&gt;!"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tc.queryParam, nil)
			rec := httptest.NewRecorder()

			helloHandler(rec, req)

			if rec.Code != tc.expectStatus {
				t.Errorf("expected status %d, got %d", tc.expectStatus, rec.Code)
			}
			body := rec.Body.String()
			if !strings.Contains(body, tc.expectBody) {
				t.Errorf("expected body to contain %q, got %q", tc.expectBody, body)
			}
			if !strings.Contains(body, "<!DOCTYPE html>") {
				t.Errorf("expected HTML response, got %q", body[:min(len(body), 200)])
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "OK" {
		t.Errorf("expected body %q, got %q", "OK", rec.Body.String())
	}
}
