// Package curl implements the Curl hashing function.
package curl

import (
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/signing/utils"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
)

// Curl is an implementation of Curl-P, it improves the speed of Curl single hasher,
// by applying the curl function in a batched manner.
// It's obvious for the first round: you can rotate the state by 364,
// then use opposite trits as inputs for the curl function.
// The resulting state will have resorted indices, but their offsets are consistent
// throughout the new state, Rotate again to have index 364 being in opposition to index 0,
// then again 'curl' them. Repeat, until rounds are done.
// Finally, sort the state.
// Similar to bct, it encodes one trit as two bits on different uint64 arrays,
// n (Negative) and p (Positive): 0 -> n=0, p=0; 1 -> n=0, p=1; -1 -> n=1, p=0.
// Top level arrays store 243 trits each, 2nd level arrays  get filled up 3*64+51.
// This way, when doing 81 rounds, resorting can be done by and-mask operations.
type Curl struct {
	n             [3][4]uint64    // negative part of the state
	p             [3][4]uint64    // positive part of the state
	rounds        CurlRounds      // number of rounds used
	direction     SpongeDirection // whether the sponge is absorbing or squeezing
	resortOffset  int             // where to find index 1 after rounds are done
	transformFunc func(n, p *[3][4]uint64, rounds, resort int)
}

var (
	fastestTransformFunc = transformGo
	fastestTransformIndex  = useGo
	availableTransformFuncs = make(map[int]func(n, p *[3][4]uint64, rounds, resort int))
)

func init(){
	fastestTransformFunc = transformGo
	fastestTransformIndex = useGo
	availableTransformFuncs[useGoGP] = transformGoGeneralPurpose
	availableTransformFuncs[useGo] = transformGo
}

// NewCurl initializes a new Curl instance.
func NewCurl(rounds ...CurlRounds) SpongeFunction {
	r := NumberOfRounds
	if len(rounds) > 0 {
		r = rounds[0]
	}
	c := Curl{}
	c.direction = SpongeAbsorbing
	c.rounds = r
	c.resortOffset = resortOffset81
	c.transformFunc = fastestTransformFunc
	if c.rounds != CurlP81 {
		c.transformFunc = transformGoGeneralPurpose
		c.resortOffset = resortOffsets[int(c.rounds) % len(resortOffsets)]
	}
	return &c
}

// NewCurlP27 returns a new Curl-P-27.
func NewCurlP27() SpongeFunction {
	return NewCurl(CurlP27)
}

// NewCurlP81 returns a new Curl-P-81.
func NewCurlP81() SpongeFunction {
	return NewCurl(CurlP81)
}

// NumRounds returns the number of rounds for this Curl instance.
func (c *Curl) NumRounds() int {
	return int(c.rounds)
}

// CopyState copy the content of the Curl state buffer into s.
func (c *Curl) CopyState(s Trits) {
	for i := range c.n {
	jLoop:
		for j := range c.n[0] {
			for k := 0; k < 64; k++ {
				if len(s) == 0 { return }
				if j == 3 && k == 51 {
					break jLoop
				}
				mask := uint64(1) << uint(k)
				if c.n[i][j] & mask != 0 {
					s[0] = -1
				} else if c.p[i][j] & mask != 0 {
					s[0] = 1
				} else {
					s[0] = 0
				}
				s = s[1:]
			}
		}
	}
}

// Squeeze squeezes out trits of the given length.
// The value tritsCount has to be a multiple of HashTrinarySize.
func (c *Curl) Squeeze(tritsCount int) (Trits, error) {
	if tritsCount%HashTrinarySize != 0 {
		return nil, ErrInvalidSqueezeLength
	}

	out := make(Trits, tritsCount)
	for p := out; len(p) >= HashTrinarySize; p = p[HashTrinarySize:] {
		c.squeeze(p)
	}
	return out, nil
}

func (c *Curl) squeeze(hash Trits) {
	// during squeezing, we only transform before each squeeze to avoid unnecessary transforms
	if c.direction == SpongeSqueezing {
		c.transform()
	}
	c.direction = SpongeSqueezing
	for i := 0; i < HashTrinarySize; i++ {
		mask := uint64(1) << uint(i % 64)
		if c.n[0][i/64] & mask != 0 {
			hash[i] = -1
		} else if c.p[0][i / 64] & mask != 0 {
			hash[i] = 1
		} else {
			hash[i] = 0
		}
	}
}

// MustSqueeze squeezes out trits of the given trit length.
// The value tritsCount has to be a multiple of HashTrinarySize. It panics if the length is not valid.
func (c *Curl) MustSqueeze(tritsCount int) Trits {
	out, err := c.Squeeze(tritsCount)
	if err != nil {
		panic(err)
	}
	return out
}

// SqueezeTrytes squeezes out trytes of the given trit length.
// The value tritsCount has to be a multiple of HashTrinarySize.
func (c *Curl) SqueezeTrytes(tritsCount int) (Trytes, error) {
	trits, err := c.Squeeze(tritsCount)
	if err != nil {
		return "", err
	}
	return MustTritsToTrytes(trits), nil
}

// MustSqueezeTrytes squeezes out trytes of the given trit length.
// The value tritsCount has to be a multiple of HashTrinarySize. It panics if the length is not valid.
func (c *Curl) MustSqueezeTrytes(tritsCount int) Trytes {
	return MustTritsToTrytes(c.MustSqueeze(tritsCount))
}

// Absorb fills the internal state of the sponge with the given trits.
func (c *Curl) Absorb(in Trits) error {
	if len(in) == 0 || len(in)%HashTrinarySize != 0 {
		return errors.Wrap(ErrInvalidTritsLength, "trits slice length must be a multiple of 243")
	}

	if c.direction != SpongeAbsorbing {
		panic("absorb after squeeze")
	}
	for len(in) >= HashTrinarySize {
		c.n[0] = [4]uint64{0, 0, 0, 0}; c.p[0] = [4]uint64{0, 0, 0, 0}
		for i, trit := range in[:HashTrinarySize] {
			mask := uint64(1) << uint(i % 64)
			switch trit {
			case -1:
				c.n[0][i / 64] |= mask
			case 1:
				c.p[0][i / 64] |= mask
			}  // same as bct: everything else is zero
		}
		in = in[HashTrinarySize:]
		c.transform()
	}
	return nil
}

// AbsorbTrytes fills the internal state of the sponge with the given trytes.
func (c *Curl) AbsorbTrytes(in Trytes) error {
	if len(in) == 0 || len(in)%HashTrytesSize != 0 {
		return errors.Wrap(ErrInvalidTrytesLength, "trytes length must be a multiple of 81")
	}

	trits, err := TrytesToTrits(in)
	if err != nil {
		return err
	}
	return c.Absorb(trits)
}

// AbsorbTrytes fills the internal state of the sponge with the given trytes.
// It panics if the given trytes are not valid.
func (c *Curl) MustAbsorbTrytes(in Trytes) {
	err := c.Absorb(MustTrytesToTrits(in))
	if err != nil {
		panic(err)
	}
}

// transform transforms the sponge.
func (c *Curl) transform() {
	c.transformFunc(&c.n, &c.p, int(c.rounds), c.resortOffset)
}

// transformGo does transformation with 81 rounds
func transformGo(n, p *[3][4]uint64, _, _ int) {
	rotationOffset := 364
	for i := 0; i < int(NumberOfRounds); i++ {
		n2 := rotate729(*n, rotationOffset)
		p2 := rotate729(*p, rotationOffset)
		for j := range n {
			for k := range n[0] {
				n[j][k], p[j][k] = curlFunc(n[j][k], p[j][k], n2[j][k], p2[j][k])
			}
			n[j][3] &= uint64(0x0007ffffffffffff) // clean up
			p[j][3] &= uint64(0x0007ffffffffffff)
		}
		rotationOffset = (rotationOffset * 364) % 729 // just this, no table
	}
	// trits are sorted 0, 487, 245, 3, 490, 248, ..., resort them
	m := [3]uint64{0x9249249249249249, 0x4924924924924924, 0x2492492492492492}
	var tmp Curl
	for i := range n {
		for j := range n[0] {
			tmp.n[i][j] = n[i][j] & m[j%3] | n[(1+i)%3][j] & m[(2+j)%3] | n[(2+i)%3][j] & m[(1+j)%3]
			tmp.p[i][j] = p[i][j] & m[j%3] | p[(1+i)%3][j] & m[(2+j)%3] | p[(2+i)%3][j] & m[(1+j)%3]
		}
	}
	*n, *p = tmp.n, tmp.p
}

// transformGoGeneralPurpose does the same rotate-and-curl,
// it can do any kind of rounds because it resorts only bit by bit.
func transformGoGeneralPurpose(n, p *[3][4]uint64, rounds, resortOffset int) {
	var trits = make([]Trits, StateSize)
	var resortedTrits = make([]Trits, StateSize)
	srcIndex := 0
	_ = trits; _ = resortedTrits; _ = srcIndex
	for dstIndex := 0; dstIndex < 729; dstIndex++ {

	}
	rotationOffset := 364
	for i := 0; i < rounds; i++ {
		n2 := rotate729(*n, rotationOffset)
		p2 := rotate729(*p, rotationOffset)
		for j := range n {
			for k := range n[0] {
				n[j][k], p[j][k] = curlFunc(n[j][k], p[j][k], n2[j][k], p2[j][k])
			}
			n[j][3] &= uint64(0x0007ffffffffffff) // clean up
			p[j][3] &= uint64(0x0007ffffffffffff)
		}
		rotationOffset = (rotationOffset * 364) % 729 // just this, no table
	}
	// resortOffset trits
	var newN, newP [3][4]uint64
	oldIndex := 0
	for newIndex := 0; newIndex < StateSize; newIndex++ {
		i := oldIndex / (StateSize / 3)
		j := (oldIndex % (StateSize / 3)) / 64
		k := (oldIndex % (StateSize / 3)) % 64
		x := newIndex / (StateSize / 3)
		y := (newIndex % (StateSize / 3)) / 64
		z := (newIndex % (StateSize / 3)) % 64
		if n[i][j] & (uint64(1) << uint(k)) != 0 {
			newN[x][y] |= uint64(1) << uint(z)
		}
		if p[i][j] & (uint64(1) << uint(k)) != 0 {
			newP[x][y] |= uint64(1) << uint(z)
		}
		oldIndex = (oldIndex + resortOffset + 1) % StateSize
	}
	*n, *p = newN, newP
}

func curlFunc(aN, aP, bN, bP uint64) (cN, cP uint64) {
	tmp := aN ^ bP
	cP = tmp &^ aP
	cN = ^tmp &^ (aP ^ bN)
	return cN, cP
}

func rotate729(src [3][4]uint64, offset int) (dst [3][4]uint64) {
	div := offset / 243; mod := offset % 243
	for i := range src {
		rshift256orInto(&dst[(i + 3 - div) % 3], &src[i], mod)
		lshift256orInto(&dst[(i + 2 - div) % 3], &src[i], 243 - mod)
	}
	return dst
}

func lshift256orInto(dst, src *[4]uint64, offset int) {
	if offset <= 0 || offset >= 256 {
		return
	}
	mod := offset % 64; div := offset / 64
	switch div {
	case 0:
		dst[0] |= src[0] << uint(mod)
		dst[1] |= src[1] << uint(mod) | src[0] >> uint(64 - mod)
		dst[2] |= src[2] << uint(mod) | src[1] >> uint(64 - mod)
		dst[3] |= src[3] << uint(mod) | src[2] >> uint(64 - mod)
	case 1:
		dst[1] |= src[0] << uint(mod)
		dst[2] |= src[1] << uint(mod) | src[0] >> uint(64 - mod)
		dst[3] |= src[2] << uint(mod) | src[1] >> uint(64 - mod)
	case 2:
		dst[2] |= src[0] << uint(mod)
		dst[3] |= src[1] << uint(mod) | src[0] >> uint(64 - mod)
	case 3:
		dst[3] |= src[0] << uint(mod)
	}
}

func rshift256orInto(dst, src *[4]uint64, offset int) {
	if offset <= 0 || offset >= 256 {
		return
	}
	mod := offset % 64; div := offset / 64
	switch div {
	case 0:
		dst[0] |= src[0] >> uint(mod) | src[1] << uint(64 - mod)
		dst[1] |= src[1] >> uint(mod) | src[2] << uint(64 - mod)
		dst[2] |= src[2] >> uint(mod) | src[3] << uint(64 - mod)
		dst[3] |= src[3] >> uint(mod)
	case 1:
		dst[0] |= src[1] >> uint(mod) | src[2] << uint(64 - mod)
		dst[1] |= src[2] >> uint(mod) | src[3] << uint(64 - mod)
		dst[2] |= src[3] >> uint(mod)
	case 2:
		dst[0] |= src[2] >> uint(mod) | src[3] << uint(64 - mod)
		dst[1] |= src[3] >> uint(mod)
	case 3:
		dst[0] |= src[3] >> uint(mod)
	}
}

// Reset the internal state of the Curl sponge.
func (c *Curl) Reset() {
	for i := 0; i < 3; i++ {
		c.n[i] = [4]uint64{0, 0, 0, 0}
		c.p[i] = [4]uint64{0, 0, 0, 0}
	}
	c.direction = SpongeAbsorbing
}

// HashTrits returns the hash of the given trits.
func HashTrits(trits Trits, rounds ...CurlRounds) (Trits, error) {
	c := NewCurl(rounds...)
	if err := c.Absorb(trits); err != nil {
		return nil, err
	}
	return c.Squeeze(HashTrinarySize)
}

// HashTrytes returns the hash of the given trytes.
func HashTrytes(t Trytes, rounds ...CurlRounds) (Trytes, error) {
	c := NewCurl(rounds...)
	if err := c.AbsorbTrytes(t); err != nil {
		return "", err
	}
	return c.SqueezeTrytes(HashTrinarySize)
}

// MustHashTrytes returns the hash of the given trytes.
// It panics if the given trytes are not valid.
func MustHashTrytes(t Trytes, rounds ...CurlRounds) Trytes {
	trytes, err := HashTrytes(t, rounds...)
	if err != nil {
		panic(err)
	}
	return trytes
}

// Clone returns a deep copy of the current Curl
func (c *Curl) Clone() SpongeFunction {
	return &Curl{
		c.n,
		c.p,
		c.rounds,
		c.direction,
		c.resortOffset,
		c.transformFunc}
}
