package t5b1_test

import (
	"bytes"
	"strings"

	"github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/encoding/t5b1"
	"github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("t5b1 encoding", func() {

	DescribeTable("valid encodings",
		func(trytes trinary.Trytes, bytes []byte) {

			By("Encode()", func() {
				src := trinary.MustTrytesToTrits(trytes)
				dst := make([]byte, EncodedLen(len(src)))
				n := Encode(dst, src)
				Expect(n).To(Equal(len(dst)))
				Expect(dst).To(Equal(bytes))
			})

			By("EncodeTrytes()", func() {
				dst := EncodeTrytes(trytes)
				Expect(dst).To(Equal(bytes))
			})

			By("Decode()", func() {
				dst := make(trinary.Trits, DecodedLen(len(bytes)))
				n, err := Decode(dst, bytes)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(dst)))
				// add expected padding
				paddedLen := DecodedLen(EncodedLen(len(trytes) * consts.TritsPerTryte))
				Expect(dst).To(Equal(trinary.MustPadTrits(trinary.MustTrytesToTrits(trytes), paddedLen)))
			})

			By("DecodeToTrytes()", func() {
				dst, err := DecodeToTrytes(bytes)
				Expect(err).ToNot(HaveOccurred())
				// add expected padding
				paddedLen := DecodedLen(EncodedLen(len(trytes)*consts.TritsPerTryte)) / consts.TritsPerTryte
				Expect(dst).To(Equal(trinary.MustPad(trytes, paddedLen)))
			})
		},
		Entry("empty", "", []byte{}),
		Entry("positive tryte values", "9ABCDEFGHIJKLM9", []byte{0x1b, 0x06, 0x25, 0xb4, 0xc5, 0x54, 0x40, 0x76, 0x04}),
		Entry("negative tryte values", "9NOPQRSTUVWXYZ9", []byte{0x94, 0x2c, 0xa2, 0x12, 0xea, 0xd1, 0xab, 0xa9, 0x00}),
		Entry("long", strings.Repeat("YZ9AB", 20), bytes.Repeat([]byte{0xe3, 0x51, 0x12}, 20)),
		Entry("no padding", "MMMMM", []byte{0x79, 0x79, 0x79}),
		Entry("1 trit padding", "MMM", []byte{0x79, 0x28}),
		Entry("2 trit padding", "M", []byte{0x0d}),
		Entry("3 trit padding", "MMMM", []byte{0x79, 0x79, 0x04}),
		Entry("4 trit padding", "MM", []byte{0x79, 0x01}),
	)

	DescribeTable("invalid trit encodings",
		func(bytes []byte, trits trinary.Trits, err error) {
			By("Decode()", func() {
				dst := make(trinary.Trits, DecodedLen(len(bytes))+10)
				n, err := Decode(dst, bytes)
				Expect(err).To(MatchError(err))
				Expect(n).To(BeNumerically("<=", DecodedLen(len(bytes))))
				Expect(dst[:n]).To(Equal(trits))
			})
		},
		Entry("invalid group value", []byte{0x80}, []int8{}, consts.ErrInvalidByte),
		Entry("above max group value", []byte{0x7a}, []int8{}, consts.ErrInvalidByte),
		Entry("below min group value", []byte{0x86}, []int8{}, consts.ErrInvalidByte),
		Entry("second group invalid", []byte{0x79, 0x7a}, []int8{1, 1, 1, 1, 1}, consts.ErrInvalidByte),
		Entry("third group invalid", []byte{0x00, 0x01, 0x7a}, []int8{0, 0, 0, 0, 0, 1, 0, 0, 0, 0}, consts.ErrInvalidByte),
	)

	DescribeTable("invalid tryte encodings",
		func(bytes []byte, err error) {
			By("DecodeToTrytes()", func() {
				dst, err := DecodeToTrytes(bytes)
				Expect(err).To(MatchError(err))
				Expect(dst).To(BeZero())
			})
		},
		Entry("invalid group value", []byte{0x80}, consts.ErrInvalidByte),
		Entry("above max group value", []byte{0x7a}, consts.ErrInvalidByte),
		Entry("below min group value", []byte{0x86}, consts.ErrInvalidByte),
		Entry("second group invalid", []byte{0x79, 0x7a}, consts.ErrInvalidByte),
		Entry("third group invalid", []byte{0x00, 0x01, 0x7a}, consts.ErrInvalidByte),
		Entry("invalid padding: [1, 1, 1, 0, 1]", []byte{0x5e}, ErrNonZeroPadding),
		Entry("invalid padding: [1, 1, 1, 0, -1]", []byte{0xc1}, ErrNonZeroPadding),
		Entry("invalid padding: [1, 1, 1, 1, 0]", []byte{0x28}, ErrNonZeroPadding),
		Entry("invalid padding: [1, 1, 1, -1, 0]", []byte{0xf2}, ErrNonZeroPadding),
	)
})
