// Package primes provides prime generation strategies.
package primes

import (
	"context"
	"errors"
	"fmt"
	"sort"
)

// errors for input validation.
var (
	ErrInvalidRange    = errors.New("invalid range: low must be >= 0 and low <= high")
	ErrRangeTooLarge   = errors.New("range too large: high must be <= 1,000,000,000")
	ErrUnknownStrategy = errors.New("unknown strategy")
)

// upper bound cap to prevent resource exhaustion
const maxHigh int64 = 1_000_000_000

// Strategy is the interface that all prime generation algos must implement.
type Strategy interface {
	// Name retuns the lowercase name of this strategy.
	Name() string
	// Generate returns all primes in the inclusive range [low, high].
	Generate(ctx context.Context, low, high int64) ([]int64, error)
}

// Registry holds all registered strategies mapped by name.
var Registry = map[string]Strategy{}

// register adds a strategy to the registry.
func register(s Strategy) {
	Registry[s.Name()] = s
}

// Get looks up a strategy by name, returns error if not found.
func Get(name string) (Strategy, error) {
	s, ok := Registry[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q (available: %v)", ErrUnknownStrategy, name, Names())
	}
	return s, nil
}

// Names returns sorted list of all registered strategy names.
func Names() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// validate checks that the range [low, high] is valid.
// called by all strategies before generation starts.
func validate(low, high int64) error {
	if low < 0 || low > high {
		return ErrInvalidRange
	}
	if high > maxHigh {
		return ErrRangeTooLarge
	}
	return nil
}
