// Package storage handles SQLite-based storage for execution records.
package storage

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" 
)

// NewDB creates a new in-memory SQLite database with shared cache.
// also runs the schema migration on startup.
func NewDB(ctx context.Context) (*sql.DB, error) {
	// Use shared cache so all pooled connections share the same in-memory database.
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		return nil, fmt.Errorf("storage: open db: %w", err)
	}

	// need to keep at least 1 idle connection alive, otherwise
	// the shared in-memory database gets dropped and data is lost.
	// MaxOpenConns >= MaxIdleConns gives headroom for concurrent requests.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	if err := migrate(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("storage: migrate: %w", err)
	}

	return db, nil
}

// migrate creates the tables if they dont exist.
func migrate(ctx context.Context, db *sql.DB) error {
	const schema = `
		CREATE TABLE IF NOT EXISTS executions (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp  TEXT    NOT NULL,
			low        INTEGER NOT NULL,
			high       INTEGER NOT NULL,
			strategy   TEXT    NOT NULL,
			elapsed_ms INTEGER NOT NULL,
			prime_count INTEGER NOT NULL
		);
	`
	_, err := db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create executions table: %w", err)
	}
	return nil
}
