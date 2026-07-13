# prime-generator

A Go project for generating prime numbers using multiple strategies, available as both a CLI tool and an HTTP server.

## Quick Start

### Build

```bash
make build
```

This creates two binaries in `./bin/`:
- `primegen` — CLI tool
- `primeserver` — HTTP server

### CLI Usage

```bash
# Generate primes between 1 and 100 using the default sieve strategy
./bin/primegen -low 1 -high 100
# Output: 2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97

# Use a specific strategy
./bin/primegen -low 1 -high 10 -strategy trial
# Output: 2, 3, 5, 7

# List available strategies
./bin/primegen -list-strategies
# Output:
# segmented
# sieve
# trial
```

### HTTP Server

```bash
# Start the server on :8080
make run-server
```

**Generate primes:**

```bash
curl "http://localhost:8080/primes?low=1&high=10&strategy=sieve"
```

Response:
```json
{
  "primes": [2, 3, 5, 7],
  "count": 4,
  "strategy": "sieve",
  "elapsed_ms": 0
}
```

**List recent executions:**

```bash
curl "http://localhost:8080/executions"
```

## Strategy Comparison

| Strategy | Time Complexity | Space Complexity | Best Use Case |
|----------|----------------|-----------------|---------------|
| **Sieve** (default) | O(n log log n) | O(n) | General-purpose; best for bounded ranges where n is the upper bound |
| **Segmented Sieve** | O(n log log n) | O(√n + block) | Narrow ranges with large upper bounds (e.g., primes between 999M and 1B) |
| **Trial Division** | O(n√n) | O(1) | Tiny ranges; correctness baseline for testing |

## Running Tests

```bash
# Run all tests
make test

# Run benchmarks
make bench
```

## Project Structure

```
prime-generator/
├── cmd/
│   ├── primegen/       CLI entry point
│   └── primeserver/    HTTP server entry point
├── internal/
│   ├── primes/         Prime generation strategies (Strategy interface + 3 implementations)
│   ├── server/         HTTP handlers, router, middleware
│   └── storage/        SQLite-based execution logging
```

All engine/business logic lives under `internal/` and cannot be imported outside this module.
