// Package sponge provides an interface for the sponge functions in IOTA.
package sponge

import (
	. "github.com/iotaledger/iota.go/trinary"
)

type SpongeFunctionCreator func() SpongeFunction

// SpongeFunction is a hash function using the sponge construction.
type SpongeFunction interface {
	Squeeze(length int) (Trits, error)
	MustSqueeze(length int) Trits
	SqueezeTrytes(length int) (Trytes, error)
	MustSqueezeTrytes(length int) Trytes
	Absorb(in Trits) error
	AbsorbTrytes(in Trytes) error
	MustAbsorbTrytes(in Trytes)
	Reset()
	Clone() SpongeFunction
}

// GetSpongeFunc checks if a hash function was given, otherwise uses defaultSpongeFuncCreator. Panics if none given.
func GetSpongeFunc(spongeFunc []SpongeFunction, defaultSpongeFuncCreator ...SpongeFunctionCreator) SpongeFunction {
	if len(spongeFunc) > 0 {
		return spongeFunc[0]
	}
	if len(defaultSpongeFuncCreator) > 0 {
		return defaultSpongeFuncCreator[0]()
	}
	panic("No sponge function given")
}
