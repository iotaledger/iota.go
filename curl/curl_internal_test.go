package curl

import (
	"fmt"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("Curl Internal", func() {
	It("compares transform functions to do the same", func() {
		c0 := NewCurl(CurlP81).(*Curl)
		c0.transformFunc = transformGoGeneralPurpose
		in := MustPad("G99GLEABCDEVILPROOFOFFRONTENDDEVINTEGRATIONCURL", TransactionTrytesSize)
		err := c0.AbsorbTrytes(in)
		Expect(err).ToNot(HaveOccurred())
		out0 := c0.MustSqueezeTrytes(3 * HashTrinarySize)
		for impl := useGo; impl < useTheVeryImpossible; impl++ {
			if f, exists := availableTransformFuncs[impl]; exists == true {
				c1 := NewCurl(CurlP81).(*Curl)
				c1.transformFunc = f
				err := c1.AbsorbTrytes(in)
				Expect(err).ToNot(HaveOccurred())
				out1 := c1.MustSqueezeTrytes(3 * HashTrinarySize)
				if out0 != out1 {
					fmt.Println("")
				}
				Expect(out0).To(Equal(out1))
			}
		}
	})
	It("tests CopyState() and importState()", func() {
		a := strings.Repeat("A", HashTrytesSize)
		b := strings.Repeat("B", HashTrytesSize)

		c1 := NewCurl().(*Curl)
		err := c1.AbsorbTrytes(a)
		Expect(err).ToNot(HaveOccurred())

		c2 := NewCurl().(*Curl)
		err = c2.AbsorbTrytes(b)
		Expect(err).ToNot(HaveOccurred())

		stateTrits := make(Trits, 729)
		c1.CopyState(stateTrits[:729])
		Expect(err).ToNot(HaveOccurred())
		err = c2.importState(stateTrits[:729])
		Expect(err).ToNot(HaveOccurred())
		Expect(err).ToNot(HaveOccurred())

		Expect(c2.MustSqueezeTrytes(HashTrinarySize)).To(Equal(c1.MustSqueezeTrytes(HashTrinarySize)))
	})
})

// importState copy the content of s into the Curl state buffer.
func (c *Curl) importState(s Trits) error {
	if len(s) != StateSize {
		return ErrInvalidTritsLength
	}
	for i := range c.n {
	jLoop:
		for j := range c.n[0] {
			c.n[i][j] = 0; c.p[i][j] = 0
			for k := 0; k < 64; k++ {
				if k == 51 && j == 3 {
					break jLoop
				}
				mask := uint64(1) << uint(k)
				switch s[0] {
				case -1:
					c.n[i][j] |= mask
				case 1:
					c.p[i][j] |= mask
				} // like bct: everything else is treated as zero
				s = s[1:]
			}
		}
	}
	return nil
}
