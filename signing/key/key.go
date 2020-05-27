/*
Package key provides functions to derive private keys from the provided entropy, e.g. subseed.

For the IOTA W-OTS, private keys need to be secLvl*6561 trits long. As a 243-trit entropy is used to
deterministically derive the key, this entropy needs to be securely extended to the desired length.

Given a cryptographic ternary sponge construction sponge, we can use sponge as an extendable-output function to derive
the key:

	key := make([]Trit, secLvl*6561)
	sponge.Absorb(entropy)
	sponge.Squeeze(key)

However, Kerl is not a true cryptographic sponge construction as its used capacity is 0 and every 243-trit block of its
output deterministically depends only on the previous block. This breaks the security of the W-OTS.
In oder to use the Keccak sponge construction, we must use the SHAKE extendable-output function producing a digest of
arbitrary length and convert the result to ternary:

	shake.Absorb(convertToBytes(entropy))
	for len(key) < secLvl*6561 {
		bytes := make([]byte, 48)
		shake.Squeeze(bytes)
		key = append(key, convertToTrits(bytes))
	}

convertToBytes converts a 243-trit balanced ternary integer (little-endian) into a signed 384-bit integer (big-endian)
and vice versa. They are defined in more detail here:
https://github.com/iotaledger/kerl/blob/master/IOTA-Kerl-spec.md#trits---bytes-encoding

This package implements both options for the key derivation Sponge and Shake, where Sponge should only be used to
preserve backward compatibility.
*/
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
func Sponge(entropy Trits, securityLevel SecurityLevel, spongeFunc sponge.SpongeFunction) (Trits, error) {
	if err := validateEntropy(entropy); err != nil {
		return nil, err
	}
	// Kerl ignores the last trit, make sure it is always zero
	if _, isKerl := spongeFunc.(*kerl.Kerl); isKerl && entropy[HashTrinarySize-1] != 0 {
		return nil, errors.Wrapf(ErrInvalidTrit, "%d at index %d (last trit non 0)", entropy[HashTrinarySize-1], HashTrinarySize-1)
	}
	defer spongeFunc.Reset()

	// absorb the entropy
	if err := spongeFunc.Absorb(entropy); err != nil {
		return nil, err
	}

	key := make(Trits, KeyFragmentLength*int(securityLevel))
	for i := 0; i < len(key); i += HashTrinarySize {
		// squeeze the next 243 trits
		out, err := spongeFunc.Squeeze(HashTrinarySize)
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
