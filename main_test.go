package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelloHandler(t *testing.T) {
	tests := []struct {
		name         string
		queryParam   string
		expectedBody string
	}{
		{"default greeting", "", "Hello, World!\n"},
		{"named greeting", "?name=Go", "Hello, Go!\n"},
		{"empty name falls back", "?name=", "Hello, !\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tc.queryParam, nil)
			rec := httptest.NewRecorder()

			helloHandler(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
			}
			if rec.Body.String() != tc.expectedBody {
				t.Errorf("expected %q, got %q", tc.expectedBody, rec.Body.String())
			}
		})
	}
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
