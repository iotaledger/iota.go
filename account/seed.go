package account

import (
	. "github.com/iotaledger/iota.go/trinary"
)

type SeedProvider interface {
	Seed() (Trytes, error)
}

type InMemorySeedProvider struct {
	seed string
}

func (imsp *InMemorySeedProvider) Seed() (Trytes, error) { return imsp.seed, nil }

func NewInMemorySeedProvider(seed Trytes) SeedProvider {
	return &InMemorySeedProvider{seed}
}
