package curl

import (
	. "github.com/iotaledger/iota.go/consts"
)

// CurlRounds is the default number of rounds used in transform.
type CurlRounds int

const (
	// StateSize is the size of the Curl hash function.
	StateSize = HashTrinarySize * 3

	// CurlP27 is used for hashing with 27 rounds
	CurlP27 CurlRounds = 27

	// CurlP81 is used for hashing with 81 rounds
	CurlP81 CurlRounds = 81

	// NumberOfRounds is the default number of rounds in transform.
	NumberOfRounds = CurlP81
)

// SpongeDirection indicates the direction trits are flowing through the sponge.
type SpongeDirection int

const (
	// SpongeAbsorbing indicates that the sponge is absorbing input.
	SpongeAbsorbing SpongeDirection = iota
	// SpongeSqueezing indicates that the sponge is being squeezed.
	SpongeSqueezing
)

var (
	// Indices stores the rotation indices for a Curl round.
	Indices [StateSize + 1]int
)

func init() {
	for i := 1; i < len(Indices); i++ {
		Indices[i] = (Indices[i-1] + rotationOffset) % StateSize
	}
}
