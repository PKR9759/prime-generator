package primes

import (
	"context"
	"errors"
	"testing"
)

// test case struct for known prime ranges
type knownRange struct {
	name     string
	low      int64
	high     int64
	expected []int64
}

var knownRanges = []knownRange{
	{"1 to 10", 1, 10, []int64{2, 3, 5, 7}},
	{"0 to 1", 0, 1, []int64{}},
	{"2 to 2", 2, 2, []int64{2}},
	{"14 to 16", 14, 16, []int64{}},
	{"0 to 2", 0, 2, []int64{2}},
	{"10 to 30", 10, 30, []int64{11, 13, 17, 19, 23, 29}},
	{"1 to 100", 1, 100, []int64{
		2, 3, 5, 7, 11, 13, 17, 19, 23, 29,
		31, 37, 41, 43, 47, 53, 59, 61, 67, 71,
		73, 79, 83, 89, 97,
	}},
}

// helper to compare two int64 slices
func equalSlices(a, b []int64) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// checks all strategies against known prime ranges to make sure math is correct
func TestAllStrategies_KnownRanges(t *testing.T) {
	strategies := []Strategy{&TrialDivision{}, &Sieve{}, &SegmentedSieve{}}
	ctx := context.Background()

	for _, s := range strategies {
		t.Run(s.Name(), func(t *testing.T) {
			for _, tc := range knownRanges {
				t.Run(tc.name, func(t *testing.T) {
					got, err := s.Generate(ctx, tc.low, tc.high)
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
					if !equalSlices(got, tc.expected) {
						t.Errorf("Generate(%d, %d) = %v, want %v", tc.low, tc.high, got, tc.expected)
					}
				})
			}
		})
	}
}

// test case for error conditions
type errorCase struct {
	name    string
	low     int64
	high    int64
	wantErr error
}

var errorCases = []errorCase{
	{"low > high", 10, 5, ErrInvalidRange},
	{"negative low", -1, 10, ErrInvalidRange},
	{"high exceeds cap", 0, 1_000_000_001, ErrRangeTooLarge},
}

// checks that all strategies properly reject invalid inputs
func TestAllStrategies_Errors(t *testing.T) {
	strategies := []Strategy{&TrialDivision{}, &Sieve{}, &SegmentedSieve{}}
	ctx := context.Background()

	for _, s := range strategies {
		t.Run(s.Name(), func(t *testing.T) {
			for _, tc := range errorCases {
				t.Run(tc.name, func(t *testing.T) {
					_, err := s.Generate(ctx, tc.low, tc.high)
					if err == nil {
						t.Fatal("expected error, got nil")
					}
					if !errors.Is(err, tc.wantErr) {
						t.Errorf("got error %v, want %v", err, tc.wantErr)
					}
				})
			}
		})
	}
}

// make sure unknown strategy name returns proper error
func TestGet_UnknownStrategy(t *testing.T) {
	_, err := Get("nonexistent")
	if !errors.Is(err, ErrUnknownStrategy) {
		t.Errorf("Get(nonexistent) error = %v, want ErrUnknownStrategy", err)
	}
}

// verify all strategies can be fetched from registry by name
func TestGet_AllRegistered(t *testing.T) {
	for _, name := range []string{"trial", "sieve", "segmented"} {
		s, err := Get(name)
		if err != nil {
			t.Fatalf("Get(%q) unexpected error: %v", name, err)
		}
		if s.Name() != name {
			t.Errorf("Name() = %q, want %q", s.Name(), name)
		}
	}
}

// check that Names() returns a sorted list
func TestNames(t *testing.T) {
	names := Names()
	if len(names) != 3 {
		t.Fatalf("Names() returned %d entries, want 3: %v", len(names), names)
	}
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("Names() not sorted: %v", names)
			break
		}
	}
}

// benchmark for trial division
func BenchmarkTrialDivision(b *testing.B) {
	s := &TrialDivision{}
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _ = s.Generate(ctx, 1, 1_000_000)
	}
}

// benchmark for sieve
func BenchmarkSieve(b *testing.B) {
	s := &Sieve{}
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _ = s.Generate(ctx, 1, 1_000_000)
	}
}

// benchmark for segmented sieve
func BenchmarkSegmentedSieve(b *testing.B) {
	s := &SegmentedSieve{}
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_, _ = s.Generate(ctx, 1, 1_000_000)
	}
}
