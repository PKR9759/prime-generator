package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"prime-generator/internal/storage"
)

// helper to create a test server with a fresh db
func setupTestServer(t *testing.T) *http.ServeMux {
	t.Helper()
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewRouter(db)
}

// check that a valid request returns correct primes with 200
func TestHandlePrimes_ValidRequest(t *testing.T) {
	mux := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/primes?low=1&high=10&strategy=sieve", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp primesResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	expectedPrimes := []int64{2, 3, 5, 7}
	if len(resp.Primes) != len(expectedPrimes) {
		t.Fatalf("expected %d primes, got %d: %v", len(expectedPrimes), len(resp.Primes), resp.Primes)
	}
	for i, p := range expectedPrimes {
		if resp.Primes[i] != p {
			t.Errorf("primes[%d] = %d, want %d", i, resp.Primes[i], p)
		}
	}
	if resp.Count != 4 {
		t.Errorf("count = %d, want 4", resp.Count)
	}
	if resp.Strategy != "sieve" {
		t.Errorf("strategy = %q, want %q", resp.Strategy, "sieve")
	}
}

// make sure default strategy is sieve when none is specified
func TestHandlePrimes_DefaultStrategy(t *testing.T) {
	mux := setupTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/primes?low=1&high=10", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp primesResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Strategy != "sieve" {
		t.Errorf("default strategy = %q, want %q", resp.Strategy, "sieve")
	}
}

// test various bad requests to make sure they all return 400
func TestHandlePrimes_InvalidRequests(t *testing.T) {
	mux := setupTestServer(t)

	tests := []struct {
		name string
		url  string
	}{
		{"missing both parameters", "/primes"},
		{"missing low parameter", "/primes?high=10"},
		{"missing high parameter", "/primes?low=1"},
		{"invalid strategy name", "/primes?low=1&high=10&strategy=bogus"},
		{"low greater than high range", "/primes?low=10&high=5"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for %s, got %d: %s", tc.name, w.Code, w.Body.String())
			}
		})
	}
}

// check that executions endpoint returns data after calling /primes
func TestHandleExecutions(t *testing.T) {
	ctx := context.Background()
	db, err := storage.NewDB(ctx)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	mux := NewRouter(db)

	// First, create an execution by calling /primes.
	req := httptest.NewRequest(http.MethodGet, "/primes?low=1&high=10", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Then, fetch executions.
	req = httptest.NewRequest(http.MethodGet, "/executions", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var records []storage.ExecutionRecord
	if err := json.NewDecoder(w.Body).Decode(&records); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(records) < 1 {
		t.Error("expected at least 1 execution record")
	}
}
