// Package signing provides functions for creating and validating essential cryptographic
// components in IOTA, such as subseeds, keys, digests and signatures.
package signing

import (
	"math"

	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/guards"
	"github.com/iotaledger/iota.go/kerl"
	. "github.com/iotaledger/iota.go/signing/utils"
	. "github.com/iotaledger/iota.go/trinary"
)

// the default SpongeFunction creator
var defaultCreator = func() SpongeFunction { return kerl.NewKerl() }

// Subseed takes a seed and an index and returns the given subseed.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func Subseed(seed Trytes, index uint64, spongeFunc ...SpongeFunction) (Trits, error) {
	if err := ValidTrytes(seed); err != nil {
		return nil, err
	} else if len(seed) != HashTrinarySize/TrinaryRadix {
		return nil, ErrInvalidSeed
	}

	trits, err := TrytesToTrits(seed)
	if err != nil {
		return nil, err
	}

	incrementedSeed := AddTrits(trits, IntToTrits(int64(index)))

	h := GetSpongeFunc(spongeFunc, defaultCreator)
	defer h.Reset()

	err = h.Absorb(incrementedSeed)
	if err != nil {
		return nil, err
	}

	subseed, err := h.Squeeze(HashTrinarySize)
	if err != nil {
		return nil, err
	}
	return subseed, err
}

// Digests hashes each segment of each key fragment 26 times and returns them.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func Digests(key Trits, spongeFunc ...SpongeFunction) (Trits, error) {
	var err error
	fragments := int(math.Floor(float64(len(key)) / KeyFragmentLength))
	digests := make(Trits, fragments*HashTrinarySize)
	buf := make(Trits, HashTrinarySize)

	h := GetSpongeFunc(spongeFunc, defaultCreator)
	defer h.Reset()

	// iterate through each key fragment
	for i := 0; i < fragments; i++ {
		keyFragment := key[i*KeyFragmentLength : (i+1)*KeyFragmentLength]

		// each fragment consists of 27 segments
		for j := 0; j < KeySegmentsPerFragment; j++ {
			copy(buf, keyFragment[j*HashTrinarySize:(j+1)*HashTrinarySize])

			// hash each segment 26 times
			for k := 0; k < KeySegmentHashRounds; k++ {
				h.Absorb(buf)
				buf, err = h.Squeeze(HashTrinarySize)
				if err != nil {
					return nil, err
				}
				h.Reset()
			}

			copy(keyFragment[j*HashTrinarySize:], buf)
		}

		// hash the key fragment (which now consists of hashed segments)
		if err := h.Absorb(keyFragment); err != nil {
			return nil, err
		}

		buf, err := h.Squeeze(HashTrinarySize)
		if err != nil {
			return nil, err
		}

		copy(digests[i*HashTrinarySize:], buf)

		h.Reset()
	}

	return digests, nil
}

// Address generates the address trits from the given digests.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func Address(digests Trits, spongeFunc ...SpongeFunction) (Trits, error) {
	h := GetSpongeFunc(spongeFunc, defaultCreator)
	defer h.Reset()

	if err := h.Absorb(digests); err != nil {
		return nil, err
	}
	return h.Squeeze(HashTrinarySize)
}

// NormalizedBundleHash normalizes the given bundle hash, with resulting digits summing to zero.
func NormalizedBundleHash(bundleHash Hash) []int8 {
	normalized := trytesToTryteValues(bundleHash)
	for i := 0; i < MaxSecurityLevel; i++ {
		normalizeHashFragment(normalized[i*27 : i*27+27])
	}
	return normalized
}

func normalizeHashFragment(fragmentTryteValues []int8) {
	sum := 0
	for i := range fragmentTryteValues {
		sum += int(fragmentTryteValues[i])
	}

	for i := range fragmentTryteValues {
		v := int(fragmentTryteValues[i]) - sum
		if v < MinTryteValue {
			sum = MinTryteValue - v
			fragmentTryteValues[i] = MinTryteValue
		} else if v > MaxTryteValue {
			sum = MaxTryteValue - v
			fragmentTryteValues[i] = MaxTryteValue
		} else {
			fragmentTryteValues[i] = int8(v)
			break
		}
	}
}

func trytesToTryteValues(trytes Trytes) []int8 {
	vs := make([]int8, len(trytes))
	for i := 0; i < len(trytes); i++ {
		vs[i] = MustTryteToTryteValue(trytes[i])
	}
	return vs
}

// SignatureFragment returns signed fragments using the given bundle hash and key fragment.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func SignatureFragment(normalizedBundleHashFragment Trits, keyFragment Trits, spongeFunc ...SpongeFunction) (Trits, error) {
	sigFrag := make(Trits, len(keyFragment))
	copy(sigFrag, keyFragment)

	h := GetSpongeFunc(spongeFunc, defaultCreator)
	defer h.Reset()

	for i := 0; i < KeySegmentsPerFragment; i++ {
		hash := sigFrag[i*HashTrinarySize : (i+1)*HashTrinarySize]

		to := MaxTryteValue - normalizedBundleHashFragment[i]
		for j := 0; j < int(to); j++ {
			if err := h.Absorb(hash); err != nil {
				return nil, err
			}
			var err error
			hash, err = h.Squeeze(HashTrinarySize)
			if err != nil {
				return nil, err
			}
			h.Reset()
		}

		copy(sigFrag[i*HashTrinarySize:], hash)
	}

	return sigFrag, nil
}

// Digest computes the digest derived from the signature fragment and normalized bundle hash.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func Digest(normalizedBundleHashFragment []int8, signatureFragment Trits, spongeFunc ...SpongeFunction) (Trits, error) {
	h := GetSpongeFunc(spongeFunc, defaultCreator)
	defer h.Reset()

	buf := make(Trits, HashTrinarySize)
	sig := make(Trits, KeySegmentsPerFragment*HashTrinarySize)

	for i := 0; i < KeySegmentsPerFragment; i++ {
		copy(buf, signatureFragment[i*HashTrinarySize:(i+1)*HashTrinarySize])

		for j := normalizedBundleHashFragment[i] + MaxTryteValue; j > 0; j-- {
			err := h.Absorb(buf)
			if err != nil {
				return nil, err
			}
			buf, err = h.Squeeze(HashTrinarySize)
			if err != nil {
				return nil, err
			}
			h.Reset()
		}

		copy(sig[i*HashTrinarySize:(i+1)*HashTrinarySize], buf)
	}

	if err := h.Absorb(sig); err != nil {
		return nil, err
	}

	return h.Squeeze(HashTrinarySize)
}

// SignatureAddress computes the address corresponding to the given signature fragments.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func SignatureAddress(fragments []Trytes, hashToSign Hash, spongeFunc ...SpongeFunction) (Hash, error) {
	if len(fragments) == 0 {
		return "", ErrInvalidSignature
	}
	if !guards.IsTrytesOfExactLength(hashToSign, HashTrytesSize) {
		return "", ErrInvalidTrytes
	}

	normalized := NormalizedBundleHash(hashToSign)
	h := GetSpongeFunc(spongeFunc, defaultCreator)
	defer h.Reset()

	digests := make(Trits, len(fragments)*HashTrinarySize)
	for i := range fragments {
		fragmentTrits, err := TrytesToTrits(fragments[i])
		if err != nil {
			return "", err
		}

		// for longer signatures (multisig) cycle through the hash fragments to compute the digest
		frag := i % (HashTrytesSize / KeySegmentsPerFragment)
		digest, err := Digest(normalized[frag*KeySegmentsPerFragment:(frag+1)*KeySegmentsPerFragment], fragmentTrits, h)
		if err != nil {
			return "", err
		}
		copy(digests[i*HashTrinarySize:], digest)
	}

	addressTrits, err := Address(digests, h)
	if err != nil {
		return "", err
	}

	address, err := TritsToTrytes(addressTrits)
	if err != nil {
		return "", err
	}
	return address, nil
}

// ValidateSignatures validates the given signature fragments by checking whether the
// digests computed from the bundle hash and fragments equal the passed address.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func ValidateSignatures(expectedAddress Hash, fragments []Trytes, bundleHash Hash, spongeFunc ...SpongeFunction) (bool, error) {
	address, err := SignatureAddress(fragments, bundleHash, spongeFunc...)
	if err != nil {
		return false, err
	}

	return expectedAddress == address, nil
}
