// Package signing provides functions for creating and validating essential cryptographic
// components in IOTA, such as subseeds, keys, digests and signatures.
package sponge

import (
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/kerl"
	. "github.com/iotaledger/iota.go/trinary"
)

type SpongeFunctionCreator func() SpongeFunction

// SpongeFunction is a hash function using the sponge construction.
type SpongeFunction interface {
	Absorb(in Trits) error
	Squeeze(length int) (Trits, error)
	Reset()
}

// NewCurlP27 returns a new CurlP27.
func NewCurlP27() SpongeFunction {
	return curl.NewCurl(curl.CurlP27)
}

// NewCurlP81 returns a new CurlP81.
func NewCurlP81() SpongeFunction {
	return curl.NewCurl(curl.CurlP81)
}

// NewKerl returns a new Kerl.
func NewKerl() SpongeFunction {
	return kerl.NewKerl()
}

// GetSpongeFunc checks if a hash function was given, otherwise uses defaultSpongeFuncCreator, or Kerl.
func GetSpongeFunc(spongeFunc []SpongeFunction, defaultSpongeFuncCreator ...SpongeFunctionCreator) SpongeFunction {
	if len(spongeFunc) > 0 {
		return spongeFunc[0]
	}
	if len(defaultSpongeFuncCreator) > 0 {
		return defaultSpongeFuncCreator[0]()
	}
	return NewKerl()
}
