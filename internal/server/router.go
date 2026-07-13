package server

import (
	"database/sql"
	"net/http"
)

// NewRouter sets up the HTTP router with all routes and middleware.
func NewRouter(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	// Register routes using Go 1.22 method+path routing.
	mux.Handle("GET /primes", chain(handlePrimes(db), loggingMiddleware, recoverMiddleware))
	mux.Handle("GET /executions", chain(handleExecutions(db), loggingMiddleware, recoverMiddleware))

	return mux
}

// chain wraps a handler with the given middleware.
// first middleware in the list runs first (outermost).
func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
