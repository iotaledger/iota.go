// Package signing provides functions for creating and validating essential cryptographic
// components in IOTA, such as subseeds, keys, digests and signatures.
package signing

import (
	"math"

	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl"
	. "github.com/iotaledger/iota.go/signing/utils"
	. "github.com/iotaledger/iota.go/trinary"
)

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

	h := GetSpongeFunc(spongeFunc, kerl.NewKerl)
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

// Key computes a new private key from the given subseed using the given security level.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func Key(subseed Trits, securityLevel SecurityLevel, spongeFunc ...SpongeFunction) (Trits, error) {
	h := GetSpongeFunc(spongeFunc, kerl.NewKerl)
	defer h.Reset()

	if err := h.Absorb(subseed); err != nil {
		return nil, err
	}

	key := make(Trits, KeyFragmentLength*int(securityLevel))

	for i := 0; i < int(securityLevel); i++ {
		for j := 0; j < KeySegmentsPerFragment; j++ {
			buf, err := h.Squeeze(HashTrinarySize)
			if err != nil {
				return nil, err
			}
			copy(key[(i*KeySegmentsPerFragment+j)*HashTrinarySize:], buf)
		}
	}

	return key, nil
}

// Digests hashes each segment of each key fragment 26 times and returns them.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func Digests(key Trits, spongeFunc ...SpongeFunction) (Trits, error) {
	var err error
	fragments := int(math.Floor(float64(len(key)) / KeyFragmentLength))
	digests := make(Trits, fragments*HashTrinarySize)
	buf := make(Trits, HashTrinarySize)

	h := GetSpongeFunc(spongeFunc, kerl.NewKerl)
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
	h := GetSpongeFunc(spongeFunc, kerl.NewKerl)
	defer h.Reset()

	if err := h.Absorb(digests); err != nil {
		return nil, err
	}
	return h.Squeeze(HashTrinarySize)
}

// NormalizedBundleHash normalizes the given bundle hash, with resulting digits summing to zero.
// It returns a slice with the tryte decimal representation without any 13/M values.
func NormalizedBundleHash(bundleHash Hash) []int8 {
	normalizedBundle := make([]int8, HashTrytesSize)
	for i := 0; i < MaxSecurityLevel; i++ {
		sum := 0
		for j := 0; j < 27; j++ {
			normalizedBundle[i*27+j] = int8(TritsToInt(MustTrytesToTrits(string(bundleHash[i*27+j]))))
			sum += int(normalizedBundle[i*27+j])
		}

		if sum >= 0 {
			for ; sum > 0; sum-- {
				for j := 0; j < 27; j++ {
					if normalizedBundle[i*27+j] > MinTryteValue {
						normalizedBundle[i*27+j]--
						break
					}
				}
			}
		} else {
			for ; sum < 0; sum++ {
				for j := 0; j < 27; j++ {
					if normalizedBundle[i*27+j] < MaxTryteValue {
						normalizedBundle[i*27+j]++
						break
					}
				}
			}
		}
	}
	return normalizedBundle
}

// SignatureFragment returns signed fragments using the given bundle hash and key fragment.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func SignatureFragment(normalizedBundleHashFragment Trits, keyFragment Trits, spongeFunc ...SpongeFunction) (Trits, error) {
	sigFrag := make(Trits, len(keyFragment))
	copy(sigFrag, keyFragment)

	h := GetSpongeFunc(spongeFunc, kerl.NewKerl)
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
	h := GetSpongeFunc(spongeFunc, kerl.NewKerl)
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

// ValidateSignatures validates the given signature fragments by checking whether the
// digests computed from the bundle hash and fragments equal the passed address.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func ValidateSignatures(expectedAddress Hash, fragments []Trytes, bundleHash Hash, spongeFunc ...SpongeFunction) (bool, error) {
	normalizedBundleHashFragments := make([][]int8, MaxSecurityLevel)
	normalizeBundleHash := NormalizedBundleHash(bundleHash)

	for i := 0; i < MaxSecurityLevel; i++ {
		normalizedBundleHashFragments[i] = normalizeBundleHash[i*KeySegmentsPerFragment : (i+1)*KeySegmentsPerFragment]
	}

	digests := make(Trits, len(fragments)*HashTrinarySize)
	for i := 0; i < len(fragments); i++ {
		trits, err := TrytesToTrits(fragments[i])
		if err != nil {
			return false, err
		}

		digest, err := Digest(normalizedBundleHashFragments[i%MaxSecurityLevel], trits, spongeFunc...)
		if err != nil {
			return false, err
		}

		copy(digests[i*HashTrinarySize:], digest)
	}

	addressTrits, err := Address(digests, spongeFunc...)
	if err != nil {
		return false, err
	}

	trytes, err := TritsToTrytes(addressTrits)
	if err != nil {
		return false, err
	}

	return expectedAddress == trytes, nil
}
