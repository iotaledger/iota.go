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

package giota

import (
	"crypto/rand"
	"errors"
)

//errors used in sign.
var (
	ErrSeedTritsLength = errors.New("seed trit slice should be HashSize entries long")
	ErrKeyTritsLength  = errors.New("key trit slice should be a multiple of HashSize*27 entries long")
)

//NewSeed generate a random seed.
func NewSeed() Trytes {
	var seed Trytes
	b := make([]byte, 81)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	for _, r := range b {
		seed += Trytes(TryteAlphabet[int(r)%len(TryteAlphabet)])
	}
	return seed
}

// NewKey takes a seed encoded as Trits, an index and a security
// level to derive a private key.
func NewKey(seedTrits Trits, index, securityLevel int) Trits {
	// Utils.increment
	for i := 0; i < index; i++ {
		for j := range seedTrits {
			seedTrits[j]++
			if seedTrits[j] > 1 {
				seedTrits[j] = -1
			} else {
				break
			}
		}
	}
	hash := seedTrits.Hash()

	c := NewCurl()
	c.Absorb(hash)

	keyTrits := make(Trits, HashSize*27*securityLevel)
	for l := 0; l < securityLevel; l++ {
		for i := 0; i < 27; i++ {
			b := c.Squeeze()
			copy(keyTrits[(l*27+i)*HashSize:], b)
		}
	}

	return keyTrits
}

//Digests calculates hash x 26 for each segments in keyTrits.
func Digests(keyTrits Trits) (Trits, error) {
	if len(keyTrits) < HashSize*27 {
		return nil, ErrKeyTritsLength
	}

	// Integer division, becaue we don't care about impartial keys.
	numKeys := len(keyTrits) / (HashSize * 27)

	digests := make(Trits, HashSize*numKeys)
	b := make(Trits, HashSize)
	for i := 0; i < numKeys; i++ {
		keyFragment := keyTrits[i*HashSize*27 : (i+1)*HashSize*27]
		for j := 0; j < 27; j++ {
			copy(b, keyFragment[j*HashSize:])
			for k := 0; k < 26; k++ {
				b = b.Hash()
			}
			copy(keyFragment[j*HashSize:], b)
		}
		s := keyFragment.Hash()
		copy(digests[i*HashSize:], s)
	}

	return digests, nil
}

//Digest calculates hash x normalizedBundleFragment[i] for each segments in keyTrits.
func Digest(normalizedBundleFragment []int8, signatureFragment Trits) Trits {
	c := NewCurl()
	b := make(Trits, HashSize)
	for i := 0; i < 27; i++ {
		copy(b, signatureFragment[i*HashSize:])
		for j := normalizedBundleFragment[i] + 13; j > 0; j-- {
			b = b.Hash()
		}
		c.Absorb(b)
	}
	return c.Squeeze()
}

//AddressFromDigests makes address from digests.
func AddressFromDigests(dig Trits) Trits {
	return dig.Hash()
}

//Sign calculates signature from bundle hash and key
//by hashing x 13-normalizedBundleFragment[i] for each segments in keyTrits.
func Sign(normalizedBundleFragment []int8, keyFragment Trits) Trits {
	b := make(Trits, HashSize)
	signatureFragment := make(Trits, len(keyFragment))
	for i := 0; i < 27; i++ {
		copy(b, keyFragment[i*HashSize:])
		for j := 0; j < 13-int(normalizedBundleFragment[i]); j++ {
			b = b.Hash()
		}
		copy(signatureFragment[i*243:], b)
	}
	return signatureFragment
}

//ValidateSig validates signatureFragment.
func ValidateSig(expectedAddress Trytes, signatureFragments []Trits, bundleHash Trytes) bool {
	normalizedBundleHash := bundleHash.Normalize()

	// Get digests
	digests := make(Trits, 243*len(signatureFragments))
	for i := 0; i < len(signatureFragments); i++ {
		digestBuffer := Digest(normalizedBundleHash[i*27*(i%3):], signatureFragments[i])
		copy(digests[i*243:], digestBuffer)
	}
	address := AddressFromDigests(digests).Trytes()

	return expectedAddress == address
}

//Address represents address for iota.
type Address Trytes

//Error types for address.
var (
	ErrInvalidAddressTrytes = errors.New("addresses are either 81 or 90 trytes in length")
	ErrInvalidAddressTrits  = errors.New("addresses are either 243 or 270 trits in length")
)

//NewAddress generates new address from seed.
func NewAddress(seed Trytes, index, security int, checksum bool) (Address, error) {
	k := NewKey(seed.Trits(), index, security)
	d, err := Digests(k)
	if err != nil {
		return "", err
	}
	a := Address(AddressFromDigests(d).Trytes())
	if !checksum {
		return a, nil
	}
	return a.WithChecksum(), nil
}

//IsValid return nil if address is valid.
func (a Address) IsValid() error {
	if !(len(a) == 81 || len(a) == 90) {
		return ErrInvalidAddressTrytes
	}
	if err := Trytes(a).IsValid(); err != nil {
		return err
	}
	return nil
}

//Checksum returns checksum trytes.
func (a Address) Checksum() Trytes {
	if len(a) == 90 {
		return Trytes(a[81:])
	}
	return Trytes(a).Trits().Hash().Trytes()[:9]
}

//Trits convert address to trits.
func (a Address) Trits() Trits {
	return Trytes(a).Trits()
}

//WithChecksum returns Address+checksum.
func (a Address) WithChecksum() Address {
	if len(a) == 90 {
		return a
	}
	cu := a.Checksum()
	return a + Address(cu)
}

//WithoutChecksum returns checksum parts of address trytes.
func (a Address) WithoutChecksum() Address {
	return a[:81]
}
