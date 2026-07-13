package server

import (
	"log/slog"
	"net/http"
	"time"
)

// recoverMiddleware catches panics so the server doesn't crash.
// logs the error and returns 500 instead.
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "error", err, "method", r.Method, "path", r.URL.Path)
				writeJSON(w, http.StatusInternalServerError, errorResponse{
					Error: "internal server error",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// statusRecorder wraps ResponseWriter to capture the status code
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware logs method, path, status and duration for each request.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sr := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(sr, r)

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sr.statusCode,
			"duration", time.Since(start).String(),
		)
	})
}
