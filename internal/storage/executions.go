package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ExecutionRecord represents a single prime generation run.
type ExecutionRecord struct {
	Timestamp  time.Time
	Low        int64
	High       int64
	Strategy   string
	ElapsedMs  int64
	PrimeCount int
}

// RecordExecution inserts an execution record into the db.
// caller is responsible for setting the timeout on ctx.
func RecordExecution(ctx context.Context, db *sql.DB, rec ExecutionRecord) error {
	const query = `
		INSERT INTO executions (timestamp, low, high, strategy, elapsed_ms, prime_count)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare insert: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		rec.Timestamp.UTC().Format(time.RFC3339),
		rec.Low,
		rec.High,
		rec.Strategy,
		rec.ElapsedMs,
		rec.PrimeCount,
	)
	if err != nil {
		return fmt.Errorf("storage: exec insert: %w", err)
	}
	return nil
}

// ListExecutions retuns the most recent execution records, ordered newest first.
func ListExecutions(ctx context.Context, db *sql.DB, limit int) ([]ExecutionRecord, error) {
	const query = `
		SELECT timestamp, low, high, strategy, elapsed_ms, prime_count
		FROM executions
		ORDER BY id DESC
		LIMIT ?
	`
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("storage: prepare list: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("storage: query list: %w", err)
	}
	defer rows.Close()

	var records []ExecutionRecord
	for rows.Next() {
		var rec ExecutionRecord
		var ts string
		if err := rows.Scan(&ts, &rec.Low, &rec.High, &rec.Strategy, &rec.ElapsedMs, &rec.PrimeCount); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		rec.Timestamp, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			return nil, fmt.Errorf("storage: parse timestamp %q: %w", ts, err)
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: rows iteration: %w", err)
	}
	return records, nil
}
