.PHONY: test bench build run-server cli clean

# Run all tests verbosely.
test:
	go test ./... -v

# Run benchmarks for the primes package.
bench:
	go test ./internal/primes -bench=. -benchmem

# Build both binaries to ./bin/.
build:
	@mkdir -p bin
	go build -o ./bin/primegen ./cmd/primegen
	go build -o ./bin/primeserver ./cmd/primeserver

# Run the HTTP server.
run-server: build
	./bin/primeserver

# Run the CLI with optional ARGS.
# Usage: make cli ARGS="-low 1 -high 100"
cli: build
	./bin/primegen $(ARGS)

# Remove built binaries.
clean:
	rm -rf bin/
