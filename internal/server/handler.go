package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"prime-generator/internal/primes"
	"prime-generator/internal/storage"
)

// response struct for successful /primes request
type primesResponse struct {
	Primes    []int64 `json:"primes"`
	Count     int     `json:"count"`
	Strategy  string  `json:"strategy"`
	ElapsedMs int64   `json:"elapsed_ms"`
}

// response struct for errors
type errorResponse struct {
	Error string `json:"error"`
}

// handlePrimes handles GET /primes requests
func handlePrimes(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters.
		lowStr := r.URL.Query().Get("low")
		highStr := r.URL.Query().Get("high")
		strategyName := r.URL.Query().Get("strategy")

		if lowStr == "" || highStr == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: "missing required query parameters: low and high",
			})
			return
		}

		low, err := strconv.ParseInt(lowStr, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: "invalid low parameter: " + err.Error(),
			})
			return
		}

		high, err := strconv.ParseInt(highStr, 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{
				Error: fmt.Sprintf("invalid high parameter: %v", err),
			})
			return
		}

		if strategyName == "" {
			strategyName = "sieve"
		}

		s, err := primes.Get(strategyName)
		if err != nil {
			if errors.Is(err, primes.ErrUnknownStrategy) {
				writeJSON(w, http.StatusBadRequest, errorResponse{
					Error: fmt.Sprintf("unknown strategy %q; valid strategies: %v", strategyName, primes.Names()),
				})
				return
			}
			slog.Error("unexpected error looking up strategy", "error", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Error: "internal server error",
			})
			return
		}

		// set a timeout for the generation call so it doesnt run forever
		genCtx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		start := time.Now()
		result, err := s.Generate(genCtx, low, high)
		elapsed := time.Since(start)
		elapsedMs := elapsed.Milliseconds()

		if err != nil {
			if errors.Is(err, primes.ErrInvalidRange) || errors.Is(err, primes.ErrRangeTooLarge) {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
				return
			}
			slog.Error("prime generation failed", "error", err, "low", low, "high", high, "strategy", strategyName)
			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Error: "internal server error",
			})
			return
		}

		// record the execution in db (with its own short timeout)
		if db != nil {
			dbCtx, dbCancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer dbCancel()

			rec := storage.ExecutionRecord{
				Timestamp:  time.Now(),
				Low:        low,
				High:       high,
				Strategy:   strategyName,
				ElapsedMs:  elapsedMs,
				PrimeCount: len(result),
			}
			if err := storage.RecordExecution(dbCtx, db, rec); err != nil {
				slog.Error("failed to record execution", "error", err)
				// Do not fail the /primes response — availability of the core feature
				// is more important than completeness of logging.
			}
		}

		writeJSON(w, http.StatusOK, primesResponse{
			Primes:    result,
			Count:     len(result),
			Strategy:  strategyName,
			ElapsedMs: elapsedMs,
		})
	}
}

// handleExecutions handles GET /executions requests
func handleExecutions(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Error: "database not configured",
			})
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 20 // default
		if limitStr != "" {
			parsed, err := strconv.Atoi(limitStr)
			if err == nil && parsed > 0 {
				limit = parsed
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		records, err := storage.ListExecutions(ctx, db, limit)
		if err != nil {
			slog.Error("failed to list executions", "error", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{
				Error: "internal server error",
			})
			return
		}

		writeJSON(w, http.StatusOK, records)
	}
}

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}
