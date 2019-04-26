// Package signing provides functions for creating and validating essential cryptographic
// components in IOTA, such as subseeds, keys, digests and signatures.
package signing

import (
	"errors"
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

	keyLength := KeyFragmentLength * int(securityLevel)
	key := make(Trits, keyLength)

	err := h.Absorb(subseed)
	if err != nil {
		return nil, err
	}

	key, err = h.Squeeze(keyLength)
	if err != nil {
		return nil, err
	}

	h.Reset()

	for i := 0; i < KeyFragmentLength*int(securityLevel); i += HashTrinarySize {
		if err := h.Absorb(key[i : i+HashTrinarySize]); err != nil {
			return nil, err
		}

		buf, err := h.Squeeze(HashTrinarySize)
		if err != nil {
			return nil, err
		}
		copy(key[i:], buf)

		h.Reset()
	}

	return key, nil
}

// Digests hashes each segment of each key fragment 26 times and returns them.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func Digests(key Trits, spongeFunc ...SpongeFunction) (Trits, error) {
	var err error
	fragments := int(math.Floor(float64(len(key)) / KeyFragmentLength))
	digests := make(Trits, HashTrinarySize)
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
				err = h.Absorb(buf)
				if err != nil {
					return nil, err
				}

				buf, err = h.Squeeze(HashTrinarySize)
				if err != nil {
					return nil, err
				}
				h.Reset()
			}

			copy(key[i*KeyFragmentLength+j*HashTrinarySize:], buf)
		}

	}

	if err := h.Absorb(key); err != nil {
		return nil, err
	}

	digests, err = h.Squeeze(HashTrinarySize)
	if err != nil {
		return nil, err
	}

	h.Reset()

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

func GetSecurityLevel(hashTrits Trits) (secLvl SecurityLevel, err error) {
	var sum int16 = 0

	for secLvl := 1; secLvl <= MaxSecurityLevel; secLvl++ {
		hashPart := hashTrits[(secLvl-1)*(HashTrinarySize/3) : (secLvl)*(HashTrinarySize/3)]

		for _, trit := range hashPart {
			sum += int16(trit)
		}
		if sum == 0 {
			return SecurityLevel(secLvl), nil
		}
	}
	return SecurityLevel(0), errors.New("No security level found")
}

// SignatureFragment returns signed fragments using the given bundle hash and key fragment.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func SignatureFragment(bundleHash Trits, keyFragment Trits, startOffset int, spongeFunc ...SpongeFunction) (Trits, error) {

	secLvl, err := GetSecurityLevel(bundleHash)
	if err != nil {
		return nil, err
	}

	keyLength := int(secLvl) * int(ISSKeyLength)
	sigFrag := make(Trits, keyLength)

	h := GetSpongeFunc(spongeFunc, kerl.NewKerl)
	defer h.Reset()

	sig := make(Trits, HashTrinarySize)

	i := startOffset % int(secLvl)
	for sigLength := 0; sigLength < keyLength; {
		for j := i * ISSChunkLength; j < (i+1)*ISSChunkLength && sigLength < keyLength; j += TrinaryRadix {
			copy(sig, keyFragment[sigLength:sigLength+HashTrinarySize])

			to := MaxTryteValue - (bundleHash[j] + bundleHash[j+1]*3 + bundleHash[j+2]*9)

			for k := 0; k < int(to); k++ {
				err := h.Absorb(sig)
				if err != nil {
					return nil, err
				}
				sig, err = h.Squeeze(HashTrinarySize)
				if err != nil {
					return nil, err
				}
				h.Reset()
			}

			copy(sigFrag[sigLength:], sig)
			sigLength += HashTrinarySize
		}

		i = (i + 1) % int(secLvl)
	}

	return sigFrag, nil
}

// Digest computes the digest derived from the signature fragment and bundle hash.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func Digest(bundleHash []int8, signatureFragment Trits, startOffset int, spongeFunc ...SpongeFunction) (Trits, error) {
	sigLength := len(signatureFragment)

	secLvl, err := GetSecurityLevel(bundleHash)
	if err != nil {
		return nil, err
	}

	sigFrag := make(Trits, sigLength)
	copy(sigFrag, signatureFragment[:sigLength])

	h := GetSpongeFunc(spongeFunc, kerl.NewKerl)
	defer h.Reset()

	dig := make(Trits, HashTrinarySize)
	sig := make(Trits, HashTrinarySize)

	i := startOffset % int(secLvl)
	for digLength := 0; digLength < sigLength; {
		for j := i * ISSChunkLength; j < (i+1)*ISSChunkLength && digLength < sigLength; j += TrinaryRadix {
			copy(sig, sigFrag[digLength:digLength+HashTrinarySize])

			to := (bundleHash[j] + bundleHash[j+1]*3 + bundleHash[j+2]*9) - MinTryteValue

			for k := 0; k < int(to); k++ {
				err := h.Absorb(sig)
				if err != nil {
					return nil, err
				}
				sig, err = h.Squeeze(HashTrinarySize)
				if err != nil {
					return nil, err
				}
				h.Reset()
			}

			copy(sigFrag[digLength:], sig)
			digLength += HashTrinarySize
		}

		i = (i + 1) % int(secLvl)
	}

	err = h.Absorb(sigFrag)
	if err != nil {
		return nil, err
	}

	dig, err = h.Squeeze(HashTrinarySize)
	if err != nil {
		return nil, err
	}

	return dig, nil
}

// ValidateSignatures validates the given signature fragments by checking whether the
// digests computed from the bundle hash and fragments equal the passed address.
// Optionally takes the SpongeFunction to use. Default is Kerl.
func ValidateSignatures(expectedAddress Hash, fragments Trytes, bundleHash Hash, spongeFunc ...SpongeFunction) (bool, error) {
	bundleHashTrits, err := TrytesToTrits(bundleHash)
	if err != nil {
		return false, err
	}

	fragmentsTrits, err := TrytesToTrits(fragments)
	if err != nil {
		return false, err
	}

	digest, err := Digest(bundleHashTrits, fragmentsTrits, 0, spongeFunc...)
	if err != nil {
		return false, err
	}

	addressTrits, err := Address(digest, spongeFunc...)
	if err != nil {
		return false, err
	}

	trytes, err := TritsToTrytes(addressTrits)
	if err != nil {
		return false, err
	}

	return expectedAddress == trytes, nil
}
