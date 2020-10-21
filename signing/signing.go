// Package signing provides functions for creating and validating essential cryptographic
// components in IOTA, such as subseeds, keys, digests and signatures.
package signing

import (
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
	if len(key)%KeyFragmentLength != 0 {
		return nil, ErrInvalidTritsLength
	}
	digests := make(Trits, 0, len(key)/KeyFragmentLength*HashTrinarySize)

	chainHash := GetSpongeFunc(spongeFunc, defaultCreator)
	defer chainHash.Reset()
	// create a second hash instance for the digests
	digestHash := chainHash.Clone()

	// iterate through each key fragment
	for len(key) >= KeyFragmentLength {
		// each fragment consists of 27 segments
		for j := 0; j < KeySegmentsPerFragment; j++ {
			// hash each segment 26 times
			tmp := key[j*HashTrinarySize : (j+1)*HashTrinarySize]
			for k := 0; k < KeySegmentHashRounds; k++ {
				err := chainHash.Absorb(tmp)
				if err != nil {
					return nil, err
				}
				tmp, err = chainHash.Squeeze(HashTrinarySize)
				if err != nil {
					return nil, err
				}
				chainHash.Reset()
			}
			// absorb all the hashed segments of one fragment
			if err := digestHash.Absorb(tmp); err != nil {
				return nil, err
			}
		}
		// append the fragment digest
		buf, err := digestHash.Squeeze(HashTrinarySize)
		if err != nil {
			return nil, err
		}
		digestHash.Reset()
		digests = append(digests, buf...)

		// advance to the next fragment
		key = key[KeyFragmentLength:]
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
			err := h.Absorb(hash)
			if err != nil {
				return nil, err
			}
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
func Digest(normalizedHashFragment []int8, signatureFragment Trits, spongeFunc ...SpongeFunction) (Trits, error) {
	if len(normalizedHashFragment) != KeySegmentsPerFragment {
		return nil, ErrInvalidTritsLength
	}
	if len(signatureFragment) != SignatureMessageFragmentTrinarySize {
		return nil, ErrInvalidTritsLength
	}

	chainHash := GetSpongeFunc(spongeFunc, defaultCreator)
	defer chainHash.Reset()
	// create a second hash instance for the digest
	digestHash := chainHash.Clone()

	for i := 0; i < KeySegmentsPerFragment; i++ {
		// hash each segment the remaining number of times to reach 26
		tmp := signatureFragment[i*HashTrinarySize : (i+1)*HashTrinarySize]
		for j := normalizedHashFragment[i] + MaxTryteValue; j > 0; j-- {
			err := chainHash.Absorb(tmp)
			if err != nil {
				return nil, err
			}
			tmp, err = chainHash.Squeeze(HashTrinarySize)
			if err != nil {
				return nil, err
			}
			chainHash.Reset()
		}
		// absorb all the hashed segments of the fragment
		if err := digestHash.Absorb(tmp); err != nil {
			return nil, err
		}
	}
	return digestHash.Squeeze(HashTrinarySize)
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

		// for backward compatibility with previous validation rules, non-zero 243rd trits in the signature are ignored
		for j := HashTrinarySize - 1; j < len(fragmentTrits); j += HashTrinarySize {
			fragmentTrits[j] = 0
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
