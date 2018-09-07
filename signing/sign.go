/*
MIT License

Copyright (c) 2016 Sascha Hanse
Copyright (c) 2017 Shinya Yagyu

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package signing

import (
	"crypto/rand"
	"errors"
	"github.com/iotaledger/giota/curl"
	"github.com/iotaledger/giota/kerl"
	"github.com/iotaledger/giota/trinary"
	"math/big"
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
	EmptySig trinary.Trytes
	// EmptyAddress represents an empty address.
	EmptyAddress Address = "999999999999999999999999999999999999999999999999999999999999999999999999999999999"
)

func init() {
	bytes := make([]byte, SignatureSize/3)
	for i := 0; i < SignatureSize/3; i++ {
		bytes[i] = '9'
	}
	EmptySig = trinary.Trytes(bytes)
}

type SecurityLevel int

const (
	SecurityLevelLow    SecurityLevel = 1
	SecurityLevelMedium SecurityLevel = 2
	SecurityLevelHigh   SecurityLevel = 3
)

// NewSeed generate a random Trytes
func NewSeed() trinary.Trytes {
	b := make([]byte, 49)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	txt := new(big.Int).SetBytes(b).Text(27)
	t := make([]byte, 81)
	for i := range t {
		var c byte = '0'
		if len(txt) > i {
			c = txt[i]
		}
		if c == '0' {
			t[i] = '9'
		}
		if c >= '1' && c <= '9' {
			t[i] = c - '1' + 'A'
		}
		if c >= 'a' {
			t[i] = c - 'a' + ('A' + 9)
		}
	}
	return trinary.Trytes(t)
}

// NewSubseed takes a seed and an index and returns the given subseed
func NewSubseed(seed trinary.Trytes, index uint) (trinary.Trits, error) {
	if err := seed.IsValid(); err != nil {
		return nil, err
	} else if len(seed) != trinary.TritHashLength/trinary.Radix {
		return nil, ErrSeedTrytesLength
	}

	incrementedSeed := seed.Trits()
	var i uint
	for ; i < index; i++ {
		trinary.IncTrits(incrementedSeed)
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

// NewKeyTrits takes a seed encoded as Trytes, an index and a security
// level to derive a private key returned as Trits
func NewKeyTrits(seed trinary.Trytes, index uint, securityLevel SecurityLevel) (trinary.Trits, error) {
	subseed, err := NewSubseed(seed, index)
	if err != nil {
		return nil, err
	}

	k := kerl.NewKerl()
	err = k.Absorb(subseed)
	if err != nil {
		return nil, err
	}

	key := make(trinary.Trits, (curl.HashSize * 27 * int(securityLevel)))

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

// NewKey takes a seed encoded as Trytes, an index and a security
// level to derive a private key returned as Trytes
func NewKey(seed trinary.Trytes, index uint, securityLevel SecurityLevel) (trinary.Trytes, error) {
	ts, err := NewKeyTrits(seed, index, securityLevel)
	return ts.Trytes(), err
}

func clearState(l *[curl.StateSize]uint64, h *[curl.StateSize]uint64) {
	for j := curl.HashSize; j < curl.StateSize; j++ {
		l[j] = 0xffffffffffffffff
		h[j] = 0xffffffffffffffff
	}
}

// 01:-1 11:0 10:1
func para27(in trinary.Trytes) (*[curl.StateSize]uint64, *[curl.StateSize]uint64) {
	var l, h [curl.StateSize]uint64

	clearState(&l, &h)
	var j uint
	bb := in.Trits()
	for i := 0; i < curl.HashSize; i++ {
		for j = 0; j < 27; j++ {
			l[i] <<= 1
			h[i] <<= 1
			switch bb[int(j)*curl.HashSize+i] {
			case 0:
				l[i] |= 1
				h[i] |= 1
			case 1:
				l[i] |= 0
				h[i] |= 1
			case -1:
				l[i] |= 1
				h[i] |= 0
			}
		}
	}
	return &l, &h
}

func seri27(l *[curl.StateSize]uint64, h *[curl.StateSize]uint64) trinary.Trytes {
	keyFragment := make(trinary.Trits, curl.HashSize*27)
	r := make(trinary.Trits, curl.HashSize)
	var n uint
	for n = 0; n < 27; n++ {
		for i := 0; i < curl.HashSize; i++ {
			ll := (l[i] >> n) & 1
			hh := (h[i] >> n) & 1
			switch {
			case hh == 0 && ll == 1:
				r[i] = -1
			case hh == 1 && ll == 1:
				r[i] = 0
			case hh == 1 && ll == 0:
				r[i] = 1
			}
		}
		copy(keyFragment[(26-n)*curl.HashSize:], r)
	}
	return keyFragment.Trytes()
}

// Digests calculates hash x 26 for each segment in keyTrits
func Digests(key trinary.Trits) (trinary.Trits, error) {
	if len(key) < curl.HashSize*27 {
		return nil, ErrKeyTritsLength
	}

	// Integer division, becaue we don't care about impartial keys.
	numKeys := len(key) / (curl.HashSize * 27)
	digests := make(trinary.Trits, curl.HashSize*numKeys)
	buffer := make(trinary.Trits, curl.HashSize)

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
func digest(normalizedBundleFragment []int8, signatureFragment trinary.Trytes) (trinary.Trits, error) {
	k := kerl.NewKerl()
	for i := 0; i < 27; i++ {
		bb := signatureFragment[i*curl.HashSize/3 : (i+1)*curl.HashSize/3].Trits()
		for j := normalizedBundleFragment[i] + 13; j > 0; j-- {
			kerl := kerl.NewKerl()
			kerl.Absorb(bb)
			bb, _ = kerl.Squeeze(curl.HashSize)
		}
		k.Absorb(bb)
	}
	tr, err := k.Squeeze(curl.HashSize)
	return tr, err
}

// Sign calculates signature from bundle hash and key
// by hashing x 13-normalizedBundleFragment[i] for each segments in keyTrits.
func Sign(normalizedBundleFragment []int8, keyFragment trinary.Trytes) trinary.Trytes {
	signatureFragment := make(trinary.Trits, len(keyFragment)*3)
	for i := 0; i < 27; i++ {
		bb := keyFragment[i*curl.HashSize/3 : (i+1)*curl.HashSize/3].Trits()
		for j := 0; j < 13-int(normalizedBundleFragment[i]); j++ {
			kerl := kerl.NewKerl()
			kerl.Absorb(bb)
			// TODO: why is the error ignored here?
			bb, _ = kerl.Squeeze(curl.HashSize)
		}
		copy(signatureFragment[i*curl.HashSize:], bb)
	}
	return signatureFragment.Trytes()
}

// IsValidSig validates signatureFragment.
func IsValidSig(expectedAddress Address, signatureFragments []trinary.Trytes, bundleHash trinary.Trytes) bool {
	normalizedBundleHash := bundleHash.Normalize()

	// Get digests
	digests := make(trinary.Trits, curl.HashSize*len(signatureFragments))
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

	address, err := ToAddress(addrTrites.Trytes())
	if err != nil {
		return false
	}

	return expectedAddress == address
}
