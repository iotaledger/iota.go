package signing

import (
	"errors"
	"github.com/iotaledger/giota/curl"
	"github.com/iotaledger/giota/kerl"
	. "github.com/iotaledger/giota/trinary"
)

const (
	SignatureSize = 6561
)

// errors used in sign
var (
	ErrSeedTritsLength  = errors.New("seed trit slice should be HashSize entries long")
	ErrSeedTrytesLength = errors.New("seed string needs to be HashSize / 3 characters long")
	ErrKeyTritsLength   = errors.New("key trit slice should be a multiple of HashSize*27 entries long")
)

var (
	// emptySig represents an empty signature.
	EmptySig Trytes
	// EmptyAddress represents an empty address.
	EmptyAddress AddressHash = "999999999999999999999999999999999999999999999999999999999999999999999999999999999"
)

func init() {
	bytes := make([]byte, SignatureSize/3)
	for i := 0; i < SignatureSize/3; i++ {
		bytes[i] = '9'
	}
	EmptySig = Trytes(bytes)
}

type SecurityLevel int

const (
	SecurityLevelLow    SecurityLevel = 1
	SecurityLevelMedium SecurityLevel = 2
	SecurityLevelHigh   SecurityLevel = 3
)

// NewSubseed takes a seed and an index and returns the given subseed
func NewSubseed(seed Trytes, index uint) (Trits, error) {
	if err := ValidTrytes(seed); err != nil {
		return nil, err
	} else if len(seed) != TritHashLength/Radix {
		return nil, ErrSeedTrytesLength
	}

	incrementedSeed := TrytesToTrits(seed)
	var i uint
	for ; i < index; i++ {
		IncTrits(incrementedSeed)
	}

	k := kerl.NewKerl()
	err := k.Absorb(incrementedSeed)
	if err != nil {
		return nil, err
	}
	subseed, err := k.Squeeze(curl.HashSize)
	if err != nil {
		return nil, err
	}
	return subseed, err
}

// NewPrivateKeyTrits takes a seed encoded as Trytes, an index and a security
// level to derive a private key returned as Trits
func NewPrivateKeyTrits(seed Trytes, index uint, securityLevel SecurityLevel) (Trits, error) {
	subseed, err := NewSubseed(seed, index)
	if err != nil {
		return nil, err
	}

	k := kerl.NewKerl()
	err = k.Absorb(subseed)
	if err != nil {
		return nil, err
	}

	key := make(Trits, curl.HashSize*27*int(securityLevel))

	for l := 0; l < int(securityLevel); l++ {
		for i := 0; i < 27; i++ {
			b, err := k.Squeeze(curl.HashSize)
			if err != nil {
				return nil, err
			}
			copy(key[(l*27+i)*curl.HashSize:], b)
		}
	}

	return key, nil
}

// NewPrivateKey takes a seed encoded as Trytes, an index and a security
// level to derive a private key returned as Trytes
func NewPrivateKey(seed Trytes, index uint, securityLevel SecurityLevel) (Trytes, error) {
	ts, err := NewPrivateKeyTrits(seed, index, securityLevel)
	return MustTritsToTrytes(ts), err
}

func clearState(l *[curl.StateSize]uint64, h *[curl.StateSize]uint64) {
	for j := curl.HashSize; j < curl.StateSize; j++ {
		l[j] = 0xffffffffffffffff
		h[j] = 0xffffffffffffffff
	}
}

// Digests calculates hash x 26 for each segment in keyTrits
func Digests(key Trits) (Trits, error) {
	if len(key) < curl.HashSize*27 {
		return nil, ErrKeyTritsLength
	}

	// Integer division, because we don't care about impartial keys.
	numKeys := len(key) / (curl.HashSize * 27)
	digests := make(Trits, curl.HashSize*numKeys)
	buffer := make(Trits, curl.HashSize)

	for i := 0; i < numKeys; i++ {
		k2 := kerl.NewKerl()
		for j := 0; j < 27; j++ {
			copy(buffer, key[i*SignatureSize+j*curl.HashSize:i*SignatureSize+(j+1)*curl.HashSize])

			for k := 0; k < 26; k++ {
				k := kerl.NewKerl()
				k.Absorb(buffer)
				buffer, _ = k.Squeeze(curl.HashSize)
			}
			k2.Absorb(buffer)
		}
		buffer, _ = k2.Squeeze(curl.HashSize)
		copy(digests[i*curl.HashSize:], buffer)
	}
	return digests, nil
}

// digest calculates hash x normalizedBundleFragment[i] for each segment in keyTrits.
func digest(normalizedBundleFragment []int8, signatureFragment Trytes) (Trits, error) {
	k := kerl.NewKerl()
	var err error
	for i := 0; i < 27; i++ {
		bb := TrytesToTrits(signatureFragment[i*curl.HashSize/3 : (i+1)*curl.HashSize/3])
		for j := normalizedBundleFragment[i] + 13; j > 0; j-- {
			k := kerl.NewKerl()
			k.Absorb(bb)
			bb, err = k.Squeeze(curl.HashSize)
			if err != nil {
				return nil, err
			}
		}
		k.Absorb(bb)
	}
	return k.Squeeze(curl.HashSize)
}

// Sign calculates signature from bundle hash and key
// by hashing x 13-normalizedBundleFragment[i] for each segments in keyTrits.
func Sign(normalizedBundleFragment []int8, keyFragment Trytes) (Trytes, error) {
	signatureFragment := make(Trits, len(keyFragment)*3)
	var err error
	for i := 0; i < 27; i++ {
		bb := TrytesToTrits(keyFragment[i*curl.HashSize/3 : (i+1)*curl.HashSize/3])
		for j := 0; j < 13-int(normalizedBundleFragment[i]); j++ {
			k := kerl.NewKerl()
			k.Absorb(bb)
			bb, err = k.Squeeze(curl.HashSize)
			if err != nil {
				return "", err
			}
		}
		copy(signatureFragment[i*curl.HashSize:], bb)
	}
	return MustTritsToTrytes(signatureFragment), nil
}

// IsValidSig validates the given signature message fragments.
func IsValidSig(expectedAddress AddressHash, signatureFragments []Trytes, bundleHash Trytes) bool {
	normalizedBundleHash := Normalize(bundleHash)

	// get digests
	digests := make(Trits, curl.HashSize*len(signatureFragments))
	for i := range signatureFragments {
		start := 27 * (i % 3)
		digestBuffer, err := digest(normalizedBundleHash[start:start+27], signatureFragments[i])
		if err != nil {
			return false
		}
		copy(digests[i*curl.HashSize:], digestBuffer)
	}

	addrTrites, err := AddressFromDigests(digests)
	if err != nil {
		return false
	}

	address, err := NewAddressHashFromTrytes(MustTritsToTrytes(addrTrites))
	if err != nil {
		return false
	}

	return expectedAddress == address
}
