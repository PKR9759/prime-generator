package primes

import "context"

// Sieve implements Strategy using the Sieve of Eratosthenes.
// builds a boolean slice sized high+1, marks composites, then collects primes.

type Sieve struct{}

func init() {
	register(&Sieve{})
}

func (s *Sieve) Name() string { return "sieve" }

// Generate returns primes in [low, high] using sieve of eratosthenes.
func (s *Sieve) Generate(ctx context.Context, low, high int64) ([]int64, error) {
	if err := validate(low, high); err != nil {
		return nil, err
	}

	// Edge case: no primes below 2.
	if high < 2 {
		return []int64{}, nil
	}

	// Build sieve: true means composite.
	size := high + 1
	composite := make([]bool, size)
	composite[0] = true
	composite[1] = true

	for i := int64(2); i*i <= high; i++ {
		if !composite[i] {
			for j := i * i; j <= high; j += i {
				composite[j] = true
			}
		}
		// check for context cancellation periodically
		if i%10000 == 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
		}
	}

	// Collect primes in [low, high].
	start := low
	if start < 2 {
		start = 2
	}
	res := []int64{}
	for i := start; i <= high; i++ {
		if !composite[i] {
			res = append(res, i)
		}
	}

	if len(res) == 0 {
		return []int64{}, nil
	}
	return res, nil
}
