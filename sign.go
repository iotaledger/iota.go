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
	"math"
	"math/big"
)

//errors used in sign.
var (
	ErrSeedTritsLength = errors.New("seed trit slice should be HashSize entries long")
	ErrKeyTritsLength  = errors.New("key trit slice should be a multiple of HashSize*27 entries long")
)

//NewSeed generate a random Trytes.
func NewSeed() Trytes {
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
	return Trytes(t)
}

// NewKey takes a seed encoded as Trits, an index and a security
// level to derive a private key.
func NewKey(seed Trytes, index, securityLevel int) Trytes {
	seedTrits := seed.Trits()
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
	seed = seedTrits.Trytes()
	hash := seed.Hash()

	k := NewKerl()
	k.Absorb(hash.Trits())

	key := make([]byte, (HashSize*27*securityLevel)/3)
	for l := 0; l < securityLevel; l++ {
		for i := 0; i < 27; i++ {
			b, _ := k.Squeeze(HashSize)
			copy(key[(l*27+i)*HashSize/3:], []byte(b.Trytes()))
		}
	}

	return Trytes(key)
}

func clearState(l *[stateSize]uint64, h *[stateSize]uint64) {
	for j := HashSize; j < stateSize; j++ {
		l[j] = 0xffffffffffffffff
		h[j] = 0xffffffffffffffff
	}
}

// 01:-1 11:0 10:1
func para27(in Trytes) (*[stateSize]uint64, *[stateSize]uint64) {
	var l, h [stateSize]uint64

	clearState(&l, &h)
	var j uint
	bb := in.Trits()
	for i := 0; i < HashSize; i++ {
		for j = 0; j < 27; j++ {
			l[i] <<= 1
			h[i] <<= 1
			switch bb[int(j)*HashSize+i] {
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

func seri27(l *[stateSize]uint64, h *[stateSize]uint64) Trytes {
	keyFragment := make(Trits, HashSize*27)
	r := make(Trits, HashSize)
	var n uint
	for n = 0; n < 27; n++ {
		for i := 0; i < HashSize; i++ {
			ll := (l[i] >> n) & 1
			hh := (h[i] >> n) & 1
			if hh == 0 && ll == 1 {
				r[i] = -1
			}
			if hh == 1 && ll == 1 {
				r[i] = 0
			}
			if hh == 1 && ll == 0 {
				r[i] = 1
			}
		}
		copy(keyFragment[(26-n)*HashSize:], r)
	}
	return keyFragment.Trytes()
}

func Digests(key Trytes) (Trytes, error) {
	//var digests Trits
	var buffer Trits
	k := key.Trits()
	digests := make(Trits, int(math.Floor(float64(len(k))/float64(SignatureSize)))*HashSize)
	for i := 0; i < int(math.Floor(float64(len(k))/float64(SignatureSize))); i++ {
		keyFragment := k[i*SignatureSize : (i+1)*SignatureSize]
		for j := 0; j < 27; j++ {
			buffer = keyFragment[j*HashSize : (j+1)*HashSize]
			for k := 0; k < MaxTryteValue-MinTryteValue; k++ {
				kKerl := NewKerl()
				kKerl.Reset()
				kKerl.Absorb(buffer)
				buffer, _ = kKerl.Squeeze(HashSize)
			}
			for k := 0; k < HashSize; k++ {
				t := keyFragment
				t[j*HashSize+k] = buffer[k]
				keyFragment = t
			}
		}

		var kerl = NewKerl()
		kerl.Reset()
		kerl.Absorb(keyFragment)
		buffer, _ = kerl.Squeeze(HashSize)
		for j := 0; j < HashSize; j++ {
			t := digests
			t[i*HashSize+j] = buffer[j]
			digests = t
		}
	}
	return digests.Trytes(), nil
}

//digest calculates hash x normalizedBundleFragment[i] for each segments in keyTrits.
func digest(normalizedBundleFragment []int8, signatureFragment Trytes) Trytes {
	c := NewCurl()
	for i := 0; i < 27; i++ {
		bb := signatureFragment[i*HashSize/3 : (i+1)*HashSize/3]
		for j := normalizedBundleFragment[i] + 13; j > 0; j-- {
			bb = bb.Hash()
		}
		c.Absorb(bb)
	}
	return c.Squeeze()
}

//Sign calculates signature from bundle hash and key
//by hashing x 13-normalizedBundleFragment[i] for each segments in keyTrits.
func Sign(normalizedBundleFragment []int8, keyFragment Trytes) Trytes {
	signatureFragment := make([]byte, len(keyFragment))
	for i := 0; i < 27; i++ {
		bb := keyFragment[i*HashSize/3 : (i+1)*HashSize/3]
		for j := 0; j < 13-int(normalizedBundleFragment[i]); j++ {
			bb = bb.Hash()
		}
		copy(signatureFragment[i*81:], bb)
	}
	return Trytes(signatureFragment)
}

//IsValidSig validates signatureFragment.
func IsValidSig(expectedAddress Address, signatureFragments []Trytes, bundleHash Trytes) bool {
	normalizedBundleHash := bundleHash.Normalize()

	// Get digests
	digests := make([]byte, 81*len(signatureFragments))
	for i := range signatureFragments {
		start := i * 27 * (i % 3)
		digestBuffer := digest(normalizedBundleHash[start:start+27], signatureFragments[i])
		copy(digests[i*81:], digestBuffer)
	}
	address := Address(Trytes(digests).Hash())
	return expectedAddress == address
}

//Address represents address without checksum for iota.
//Don't use type cast,  use ToAddress instead
//to check the validity.
type Address Trytes

//Error types for address.
var (
	ErrInvalidAddressTrytes = errors.New("addresses without checksum are 81 trytes in length")
	ErrInvalidAddressTrits  = errors.New("addresses without checksum are 243 trits in length")
)

//NewAddress generates a new address from seed without checksum.
func NewAddress(seed Trytes, index, security int) (Address, error) {
	k := NewKey(seed, index, security)
	d, err := Digests(k)
	if err != nil {
		return "", err
	}
	return d.Hash().ToAddress()
}

//NewAddresses generates new count addresses from seed without checksum.
func NewAddresses(seed Trytes, start, count, security int) ([]Address, error) {
	as := make([]Address, count)
	var err error
	for i := 0; i < count; i++ {
		as[i], err = NewAddress(seed, start+i, security)
		if err != nil {
			return nil, err
		}
	}
	return as, nil
}

//ToAddress convert string to address,
//and checks the validity.
func ToAddress(t string) (Address, error) {
	return Trytes(t).ToAddress()
}

//ToAddress convert trytes(with and without checksum) to address,
//and checks the validity.
func (t Trytes) ToAddress() (Address, error) {
	if len(t) == 90 {
		t = t[:81]
	}
	a := Address(t)
	err := a.IsValid()
	if err != nil {
		return "", err
	}
	if len(t) == 90 {
		cs := a.Checksum()
		if t[81:] != cs {
			return "", errors.New("checksum is illegal")
		}
	}
	return a, nil
}

//IsValid return nil if address is valid.
func (a Address) IsValid() error {
	if !(len(a) == 81) {
		return ErrInvalidAddressTrytes
	}
	if err := Trytes(a).IsValid(); err != nil {
		return err
	}
	return nil
}

//Checksum returns checksum trytes.
//This panics if len(address)<81
func (a Address) Checksum() Trytes {
	if len(a) != 81 {
		panic("len(address) must be 81")
	}

	addressTrits := Trytes(a).Trits()

	kerl := NewKerl()
	kerl.Reset()
	kerl.Absorb(addressTrits)
	checksumTrits, _ := kerl.Squeeze(HashSize)

	checksumTrytes := checksumTrits.Trytes()
	checksum := checksumTrytes[len(checksumTrytes)-9:]

	return checksum
}

//WithChecksum returns Address+checksum.
//This panics if len(address)<81
func (a Address) WithChecksum() Trytes {
	if len(a) != 81 {
		panic("len(address) must be 81")
	}
	cu := a.Checksum()
	return Trytes(a) + cu
}
