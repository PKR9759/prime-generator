package storage

import (
	"context"
	"testing"
	"time"
)

// check if db initializes properly and the table exists
func TestNewDB(t *testing.T) {
	ctx := context.Background()
	db, err := NewDB(ctx)
	if err != nil {
		t.Fatal("NewDB failed:", err)
	}
	defer db.Close()

	// Verify the executions table exists by running a simple query.
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM executions").Scan(&count)
	if err != nil {
		t.Fatalf("query executions table: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows, got %d", count)
	}
}

// insert some records and make sure they come back in the right order
func TestRecordAndListExecutions(t *testing.T) {
	ctx := context.Background()
	db, err := NewDB(ctx)
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Second)

	records := []ExecutionRecord{
		{Timestamp: now, Low: 1, High: 100, Strategy: "sieve", ElapsedMs: 5, PrimeCount: 25},
		{Timestamp: now.Add(time.Second), Low: 1, High: 1000, Strategy: "trial", ElapsedMs: 50, PrimeCount: 168},
		{Timestamp: now.Add(2 * time.Second), Low: 100, High: 200, Strategy: "segmented", ElapsedMs: 3, PrimeCount: 21},
	}

	for _, rec := range records {
		if err := RecordExecution(ctx, db, rec); err != nil {
			t.Fatalf("RecordExecution failed: %v", err)
		}
	}

	// List all records (limit 10).
	got, err := ListExecutions(ctx, db, 10)
	if err != nil {
		t.Fatalf("ListExecutions failed: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 records, got %d", len(got))
	}


	// Results should be ordered by id DESC (most recent first).
	if got[0].Strategy != "segmented" {
		t.Errorf("first record strategy = %q, want %q", got[0].Strategy, "segmented")
	}
	if got[0].PrimeCount != 21 {
		t.Errorf("first record prime_count = %d, want %d", got[0].PrimeCount, 21)
	}
	if got[2].Strategy != "sieve" {
		t.Errorf("last record strategy = %q, want %q", got[2].Strategy, "sieve")
	}

	// Test limit.
	limited, err := ListExecutions(ctx, db, 1)
	if err != nil {
		t.Fatalf("ListExecutions(limit=1) failed: %v", err)
	}
	if len(limited) != 1 {
		t.Fatalf("expected 1 record with limit=1, got %d", len(limited))
	}
}
