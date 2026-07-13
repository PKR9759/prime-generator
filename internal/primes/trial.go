package primes

import (
	"context"
	"math"
)

// TrialDivision implements Strategy using trial division.
// For each n in [low, high], it checks primality by dividing
// by all integers up to sqrt(n). Used as baseline for testing.
type TrialDivision struct{}

func init() {
	register(&TrialDivision{})
}

func (t *TrialDivision) Name() string { return "trial" }

// Generate retuns all primes in [low, high] using trial division.
func (t *TrialDivision) Generate(ctx context.Context, low, high int64) ([]int64, error) {
	if err := validate(low, high); err != nil {
		return nil, err
	}

	var res []int64
	for n := max(low, 2); n <= high; n++ {
		// periodically check if context was cancelled
		if n%10000 == 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
		}
		if isPrime(n) {
			res = append(res, n)
		}
	}
	return res, nil
}

// isPrime checks whether n is prime by dividing up to sqrt(n).
func isPrime(n int64) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	limit := int64(math.Sqrt(float64(n)))
	for i := int64(3); i <= limit; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}
