// Package key provides functions to derive private keys from the provided entropy, e.g. subseed.
package key

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl"
	"github.com/iotaledger/iota.go/kerl/sha3"
	sponge "github.com/iotaledger/iota.go/signing/utils"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
)

// Shake derives a private key of the given securityLevel from entropy using the SHAKE256 extendable-output function.
// The entropy must be a slice of exactly 243 trits where the last trit is zero.
// Shake derives its security assumptions from the properties of the underlying SHAKE function.
func Shake(entropy Trits, securityLevel SecurityLevel) (Trits, error) {
	if err := validateEntropy(entropy); err != nil {
		return nil, err
	}
	// the last trit will always be ignored, make sure it is always zero
	if entropy[HashTrinarySize-1] != 0 {
		return nil, errors.Wrapf(ErrInvalidTrit, "%d at index %d (last trit non 0)", entropy[HashTrinarySize-1], HashTrinarySize-1)
	}

	// use SHAKE instead of Kerl
	h := sha3.NewShake256()

	// absorb the entropy
	in, err := kerl.KerlTritsToBytes(entropy)
	if err != nil {
		return nil, err
	}
	h.Write(in)

	key := make(Trits, KeyFragmentLength*int(securityLevel))
	buf := make([]byte, HashBytesSize)
	for i := 0; i < len(key); i += HashTrinarySize {
		// squeeze the next 48 bytes
		h.Read(buf)
		out, err := kerl.KerlBytesToTrits(buf)
		if err != nil {
			return nil, err
		}
		copy(key[i:], out)
	}

	return key, nil
}

// Sponge derives a private key of the given securityLevel from entropy using the provided ternary sponge construction.
// The entropy must be a slice of exactly 243 trits.
//
// Deprecated: Sponge only generates secure keys for sponge constructions, but Kerl is not a true sponge construction.
// Consider using Shake instead or Sponge with Curl. In case that Kerl must be used in Sponge, it must be assured that
// no chunk of the private key is ever revealed, as this would allow the reconstruction of successive chunks
// (also known as "M-bug").
// This function only allows the usage of Kerl to provide compatibility to the currently used key derivation.
func Sponge(entropy Trits, securityLevel SecurityLevel, h sponge.SpongeFunction) (Trits, error) {
	if err := validateEntropy(entropy); err != nil {
		return nil, err
	}
	// Kerl ignores the last trit, make sure it is always zero
	if _, isKerl := h.(*kerl.Kerl); isKerl && entropy[HashTrinarySize-1] != 0 {
		return nil, errors.Wrapf(ErrInvalidTrit, "%d at index %d (last trit non 0)", entropy[HashTrinarySize-1], HashTrinarySize-1)
	}
	defer h.Reset()

	// absorb the entropy
	if err := h.Absorb(entropy); err != nil {
		return nil, err
	}

	key := make(Trits, KeyFragmentLength*int(securityLevel))
	for i := 0; i < len(key); i += HashTrinarySize {
		// squeeze the next 243 trits
		out, err := h.Squeeze(HashTrinarySize)
		if err != nil {
			return nil, err
		}
		copy(key[i:], out)
	}

	return key, nil
}

func validateEntropy(entropy Trits) error {
	// the entropy must be exactly 243 trits
	if len(entropy) != HashTrinarySize {
		return errors.Wrapf(ErrInvalidTritsLength, "must be %d in size", HashTrinarySize)
	}
	return ValidTrits(entropy)
}
