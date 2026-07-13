// primegen is a CLI tool for generating prime numbers.
//
// Usage:
//
//	primegen -low 1 -high 100 -strategy sieve
//	primegen -list-strategies
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"prime-generator/internal/primes"
)

func main() {
	low := flag.Int64("low", 0, "lower bound of the range (inclusive)")
	high := flag.Int64("high", 0, "upper bound of the range (inclusive)")
	strategy := flag.String("strategy", "sieve", "prime generation strategy")
	listStrategies := flag.Bool("list-strategies", false, "list available strategies and exit")

	flag.Parse()

	if *listStrategies {
		for _, name := range primes.Names() {
			fmt.Println(name)
		}
		os.Exit(0)
	}

	s, err := primes.Get(*strategy)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	result, err := s.Generate(context.Background(), *low, *high)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// print as comma-separated list
	parts := make([]string, len(result))
	for i, p := range result {
		parts[i] = fmt.Sprintf("%d", p)
	}
	fmt.Println(strings.Join(parts, ", "))
}
