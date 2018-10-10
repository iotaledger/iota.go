package signing

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl"
	. "github.com/iotaledger/iota.go/trinary"
	"math"
)

const (
	KeyFragmentLength = 6561
)

// Subseed takes a seed and an index and returns the given subseed.
func Subseed(seed Trytes, index uint64) (Trits, error) {
	if err := ValidTrytes(seed); err != nil {
		return nil, err
	} else if len(seed) != HashTrinarySize/TrinaryRadix {
		return nil, ErrInvalidSeed
	}

	incrementedSeed := AddTrits(MustTrytesToTrits(seed), IntToTrits(int64(index)))

	k := kerl.NewKerl()
	err := k.Absorb(incrementedSeed)
	if err != nil {
		return nil, err
	}
	subseed, err := k.Squeeze(HashTrinarySize)
	if err != nil {
		return nil, err
	}
	return subseed, err
}

// Key computes a new private key from the given subseed using the given security level.
func Key(subseed Trits, securityLevel SecurityLevel) (Trits, error) {
	k := kerl.NewKerl()
	if err := k.Absorb(subseed); err != nil {
		return nil, err
	}

	key := make(Trits, HashTrinarySize*27*int(securityLevel))

	for i := 0; i < int(securityLevel); i++ {
		for j := 0; j < 27; j++ {
			b, err := k.Squeeze(HashTrinarySize)
			if err != nil {
				return nil, err
			}
			copy(key[(i*27+j)*HashTrinarySize:], b)
		}
	}

	return key, nil
}

// Digests hashes each segment of each key fragment 26 times and returns them.
func Digests(key Trits) (Trits, error) {
	var err error
	fragments := int(math.Floor(float64(len(key)) / 6561))
	digests := make(Trits, fragments*243)
	buf := make(Trits, HashTrinarySize)

	// iterate through each key fragment
	for i := 0; i < fragments; i++ {
		keyFragment := key[i*6561 : (i+1)*6561]

		// each fragment consists of 27 segments
		for j := 0; j < 27; j++ {
			copy(buf, keyFragment[j*243:(j+1)*243])

			// hash each segment 26 times
			for k := 0; k < 26; k++ {
				k := kerl.NewKerl()
				k.Absorb(buf)
				buf, err = k.Squeeze(HashTrinarySize)
				if err != nil {
					return nil, err
				}
			}

			for k := 0; k < 243; k++ {
				keyFragment[j*243+k] = buf[k]
			}
		}

		// hash the key fragment (which now consists of hashed segments)
		k := kerl.NewKerl()
		if err := k.Absorb(keyFragment); err != nil {
			return nil, err
		}

		buf, err := k.Squeeze(HashTrinarySize)
		if err != nil {
			return nil, err
		}
		for j := 0; j < 243; j++ {
			digests[i*243+j] = buf[j]
		}
	}

	return digests, nil
}

// Address generates the address trits from the given digests.
func Address(digests Trits) (Trits, error) {
	k := kerl.NewKerl()
	if err := k.Absorb(digests); err != nil {
		return nil, err
	}
	return k.Squeeze(HashTrinarySize)
}

// SignatureFragment returns signed fragments using the given key fragment.
func SignatureFragment(normalizedBundleFragments Trits, keyFragment Trits) (Trits, error) {
	sigFrag := make(Trits, len(keyFragment))
	copy(sigFrag, keyFragment)

	k := kerl.NewKerl()

	for i := 0; i < 27; i++ {
		hash := sigFrag[i*243 : (i+1)*243]

		to := 13 - normalizedBundleFragments[i]
		for j := 0; j < int(to); j++ {
			k.Reset()
			if err := k.Absorb(hash); err != nil {
				return nil, err
			}
			var err error
			hash, err = k.Squeeze(243)
			if err != nil {
				return nil, err
			}
		}

		for j := 0; j < 243; j++ {
			sigFrag[i*243+j] = hash[j]
		}
	}

	return sigFrag, nil
}

// ValidateSignatures validates the given fragments.
func ValidateSignatures(expectedAddress Hash, fragments []Trytes, bundleHash Hash) (bool, error) {
	normalizedBundleHashFragments := []Trits{}
	normalizeBundleHash := NormalizedBundleHash(bundleHash)

	for i := 0; i < 3; i++ {
		normalizedBundleHashFragments[i] = normalizeBundleHash[i*27 : (i+1)*27]
	}

	digests := make(Trits, len(fragments)*243)
	for i := 0; i < len(fragments); i++ {
		digest, err := Digest(normalizedBundleHashFragments[i%3], MustTrytesToTrits(fragments[i]))
		if err != nil {
			return false, err
		}
		for j := 0; j < 243; j++ {
			digests[i*243+j] = digest[j]
		}
	}

	addressTrits, err := Address(digests)
	if err != nil {
		return false, err
	}
	return expectedAddress == MustTritsToTrytes(addressTrits), nil
}

// Digest computes the digest derived from the signature fragment and normalized bundle hash.
func Digest(normalizedBundleHashFragment Trits, signatureFragment Trits) (Trits, error) {
	k := kerl.NewKerl()
	buf := make(Trits, HashTrinarySize)

	for i := 0; i < 27; i++ {
		copy(buf, signatureFragment[i*243:(i+1)*243])

		for j := normalizedBundleHashFragment[i] + 13; j > 0; j-- {
			kk := kerl.NewKerl()
			err := kk.Absorb(buf)
			if err != nil {
				return nil, err
			}
			buf, err = kk.Squeeze(HashTrinarySize)
			if err != nil {
				return nil, err
			}
		}

		if err := k.Absorb(buf); err != nil {
			return nil, err
		}
	}

	return k.Squeeze(HashTrinarySize)
}

// NormalizedBundleHash normalizes the given bundle hash, with resulting digits summing to zero.
func NormalizedBundleHash(bundleHash Hash) Trits {
	normalizedBundle := make([]int8, HashTrinarySize)
	for i := 0; i < 3; i++ {
		sum := 0
		for j := 0; j < 27; j++ {
			normalizedBundle[i*27+j] = int8(TritsToInt(MustTrytesToTrits(string(bundleHash[i*27*j]))))
			sum += int(normalizedBundle[i*27+j])
		}

		if sum >= 0 {
			for ; sum > 0; sum-- {
				for j := 0; j < 27; j++ {
					if normalizedBundle[i*27+j] > -13 {
						normalizedBundle[i*27+j]--
						break
					}
				}
			}
		} else {
			for ; sum < 0; sum++ {
				for j := 0; j < 27; j++ {
					if normalizedBundle[i*27+j] < 13 {
						normalizedBundle[i*27+j]++
						break
					}
				}
			}
		}
	}
	return normalizedBundle
}
