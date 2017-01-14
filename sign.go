package giota

import (
	"errors"
	//"log"
)

var (
	ErrSeedTritsLength = errors.New("seed trit slice should be HashSize entries long")
	ErrKeyTritsLength  = errors.New("key trit slice should be a multiple of HashSize*27 entries long")
)

// GenerateKey takes a seed encoded as Trits, an index and a security
// level to derive a private key.
func GenerateKey(seedTrits []int, index, securityLevel int) ([]int, error) {
	// Utils.increment
	for i := 0; i < index; i++ {
		for j, _ := range seedTrits {
			seedTrits[j] += 1
			if seedTrits[j] > 1 {
				seedTrits[j] = -1
			} else {
				break
			}
		}
	}

	c := &Curl{}
	c.Init(nil)
	c.Absorb(seedTrits)
	hash := c.Squeeze()

	c.Reset()
	c.Absorb(hash)

	keyTrits := make([]int, HashSize*27*securityLevel)
	//b := make([]int, HashSize)
	offset := 0

	for l := securityLevel; l > 0; l-- {
		for i := 0; i < 27; i++ {
			b := c.Squeeze()
			for j := 0; j < HashSize; j++ {
				keyTrits[offset] = b[j]
				offset += 1
			}
		}
	}

	return keyTrits, nil
}

func Digests(keyTrits []int) ([]int, error) {
	if len(keyTrits) < HashSize*27 {
		return nil, ErrKeyTritsLength
	}

	// Integer division, becaue we don't care about impartial keys.
	numKeys := len(keyTrits) / (HashSize * 27)

	digests := make([]int, HashSize*numKeys)
	b := make([]int, HashSize)
	for i := 0; i < numKeys; i++ {
		keyFragment := keyTrits[i*HashSize*27 : (i+1)*HashSize*27]
		for j := 0; j < 27; j++ {
			copy(b, keyFragment[j*HashSize:(j+1)*HashSize])
			c := &Curl{}
			c.Init(nil)
			for k := 0; k < 26; k++ {
				c.Absorb(b)
				b = c.Squeeze()
				c.Reset()
			}

			copy(keyFragment[j*HashSize:], b)
		}

		c := &Curl{}
		c.Init(nil)
		c.Absorb(keyFragment)
		s := c.Squeeze()
		copy(digests[i*HashSize:], s)
	}

	return digests, nil
}

func Digest(normalizedBundleFragment, signatureFragment []int) ([]int, error) {
	c := &Curl{}
	c.Init(nil)

	b := make([]int, HashSize)
	for i := 0; i < 27; i++ {
		copy(b, signatureFragment[i*HashSize:(i+1)*HashSize])
		ic := &Curl{}
		ic.Init(nil)
		for j := normalizedBundleFragment[i] + 13; j > 0; j-- {
			ic.Absorb(b)
			b = ic.Squeeze()
			ic.Reset()
		}
		c.Absorb(b)
	}

	return c.Squeeze(), nil
}
