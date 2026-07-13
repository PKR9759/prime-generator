package primes

import (
	"context"
	"math"
	"runtime"
	"sort"
	"sync"
)

// SegmentedSieve implements Strategy using a segmented sieve of eratosthenes.
// Its memory-efficient because it only allocates block-sized segments.
// Block processing is parallelized with goroutines bounded by runtime.NumCPU().
type SegmentedSieve struct{}

func init() {
	register(&SegmentedSieve{})
}

func (s *SegmentedSieve) Name() string { return "segmented" }

// segment size for each block
const blockSize int64 = 32768

// Generate retuns all primes in [low, high] using segmented sieve.
func (s *SegmentedSieve) Generate(ctx context.Context, low, high int64) ([]int64, error) {
	if err := validate(low, high); err != nil {
		return nil, err
	}

	if high < 2 {
		return []int64{}, nil
	}

	// Sieve small primes up to sqrt(high).
	sqrtHigh := int64(math.Sqrt(float64(high)))
	smallPrimes := smallSieve(sqrtHigh)

	// Determine effective low (at least 2).
	effLow := low
	if effLow < 2 {
		effLow = 2
	}

	// Divide [effLow, high] into blocks and process in parallel.
	type blockResult struct {
		index  int
		primes []int64
	}

	// Build list of block ranges.
	type blockRange struct {
		lo, hi int64
	}
	var blocks []blockRange
	for lo := effLow; lo <= high; lo += blockSize {
		hi := lo + blockSize - 1
		if hi > high {
			hi = high
		}
		blocks = append(blocks, blockRange{lo, hi})
	}

	numWorkers := runtime.NumCPU()
	if numWorkers > len(blocks) {
		numWorkers = len(blocks)
	}

	results := make([]blockResult, len(blocks))
	var wg sync.WaitGroup
	sem := make(chan struct{}, numWorkers)
	var cancelErr error
	var mu sync.Mutex

	for i, blk := range blocks {
		wg.Add(1)
		sem <- struct{}{} // acquire semaphore slot

		go func(idx int, lo, hi int64) {
			defer wg.Done()
			defer func() { <-sem }() // release slot

			// check context before starting
			select {
			case <-ctx.Done():
				mu.Lock()
				if cancelErr == nil {
					cancelErr = ctx.Err()
				}
				mu.Unlock()
				return
			default:
			}

			primes := sieveBlock(lo, hi, smallPrimes)
			results[idx] = blockResult{index: idx, primes: primes}
		}(i, blk.lo, blk.hi)
	}

	wg.Wait()

	if cancelErr != nil {
		return nil, cancelErr
	}

	// Merge results in order.
	var merged []int64
	for _, r := range results {
		merged = append(merged, r.primes...)
	}

	if merged == nil {
		return []int64{}, nil
	}
	return merged, nil
}

// smallSieve retuns all primes up to limit using a basic sieve.
func smallSieve(limit int64) []int64 {
	if limit < 2 {
		return nil
	}
	composite := make([]bool, limit+1)
	composite[0] = true
	composite[1] = true
	for i := int64(2); i*i <= limit; i++ {
		if !composite[i] {
			for j := i * i; j <= limit; j += i {
				composite[j] = true
			}
		}
	}
	var primes []int64
	for i := int64(2); i <= limit; i++ {
		if !composite[i] {
			primes = append(primes, i)
		}
	}
	return primes
}

// sieveBlock sieves the range [lo, hi] using the small primes
// and returns primes found in that range.
func sieveBlock(lo, hi int64, smallPrimes []int64) []int64 {
	size := hi - lo + 1
	composite := make([]bool, size)

	for _, p := range smallPrimes {
		// Find the first multiple of p >= lo.
		start := ((lo + p - 1) / p) * p
		if start == p {
			start += p // Don't mark the prime itself as composite.
		}
		for j := start; j <= hi; j += p {
			composite[j-lo] = true
		}
	}

	var primes []int64
	for i := int64(0); i < size; i++ {
		if !composite[i] {
			n := lo + i
			if n >= 2 {
				primes = append(primes, n)
			}
		}
	}

	// Ensure sorted output (should already be in order within a block).
	sort.Slice(primes, func(i, j int) bool { return primes[i] < primes[j] })
	return primes
}
