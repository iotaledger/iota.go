package curl

import (
	. "github.com/iotaledger/iota.go/consts"
)

const (
	// StateSize is the size of the Curl hash function.
	StateSize = HashTrinarySize * 3

	// NumRounds is the number of rounds in a Curl transform.
	NumRounds = 81
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
