package account

import (
	. "github.com/iotaledger/iota.go/trinary"
)

// SeedProvider is a provider which provides a seed.
type SeedProvider interface {
	Seed() (Trytes, error)
}

// InMemorySeedProvider is a SeedProvider which holds the seed in memory.
type InMemorySeedProvider struct {
	seed string
}

func (imsp *InMemorySeedProvider) Seed() (Trytes, error) { return imsp.seed, nil }

// NewInMemorySeedProvider creates a new InMemorySeedProvider providing the given seed.
func NewInMemorySeedProvider(seed Trytes) SeedProvider {
	return &InMemorySeedProvider{seed}
}
