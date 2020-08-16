package b1t6_test

import (
	"bytes"
	"encoding/hex"
	"strings"

	. "github.com/iotaledger/iota.go/encoding/b1t6"
	"github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("b1t6 encoding", func() {

	DescribeTable("valid encodings",
		func(bytes []byte, trytes trinary.Trytes) {

			By("Encode()", func() {
				dst := make(trinary.Trits, EncodedLen(len(bytes)))
				n := Encode(dst, bytes)
				Expect(n).To(Equal(len(dst)))
				Expect(dst).To(Equal(trinary.MustTrytesToTrits(trytes)))
			})

			By("EncodeToTrytes()", func() {
				dst := EncodeToTrytes(bytes)
				Expect(dst).To(Equal(trytes))
			})

			By("Decode()", func() {
				src := trinary.MustTrytesToTrits(trytes)
				dst := make([]byte, DecodedLen(len(src)))
				n, err := Decode(dst, src)
				Expect(err).ToNot(HaveOccurred())
				Expect(n).To(Equal(len(dst)))
				Expect(dst).To(Equal(bytes))
			})

			By("DecodeTrytes()", func() {
				dst, err := DecodeTrytes(trytes)
				Expect(err).ToNot(HaveOccurred())
				Expect(dst).To(Equal(bytes))
			})
		},
		Entry("empty", []byte{}, ""),
		Entry("normal", []byte{1}, "A9"),
		Entry("min byte value", []byte{0}, "99"),
		Entry("max byte value", []byte{255}, "Z9"),
		Entry("max trytes value", []byte{127}, "SE"),
		Entry("min trytes value", []byte{128}, "GV"),
		Entry("endianness", []byte{0, 1}, "99A9"),
		Entry("long", bytes.Repeat([]byte{0, 1}, 25), strings.Repeat("99A9", 25)),
		Entry("RFC example I", MustDecodeHex("00"), "99"),
		Entry("RFC example II", MustDecodeHex("0001027e7f8081fdfeff"), "99A9B9RESEGVHVX9Y9Z9"),
		Entry("RFC example III", MustDecodeHex("9ba06c78552776a596dfe360cc2b5bf644c0f9d343a10e2e71debecd30730d03"), "GWLW9DLDDCLAJDQXBWUZYZODBYPBJCQ9NCQYT9IYMBMWNASBEDTZOYCYUBGDM9C9"),
	)

	DescribeTable("invalid encodings",
		func(trytes trinary.Trytes, bytes []byte, err error) {

			By("Decode()", func() {
				trits := trinary.MustTrytesToTrits(trytes)
				dst := make([]byte, DecodedLen(len(trits))+10)
				n, err := Decode(dst, trits)
				Expect(err).To(MatchError(err))
				Expect(n).To(BeNumerically("<=", DecodedLen(len(trits))))
				Expect(dst[:n]).To(Equal(bytes))
			})

			By("DecodeTrytes()", func() {
				dst, err := DecodeTrytes(trytes)
				Expect(err).To(MatchError(err))
				Expect(dst).To(BeZero())
			})
		},
		Entry("one tryte", "A", []byte{}, ErrInvalidLength),
		Entry("three trytes", "A9A", []byte{1}, ErrInvalidLength),
		Entry("five trytes", "99A9A", []byte{0, 1}, ErrInvalidLength),
		Entry("above max group value", "TE", []byte{}, ErrInvalidTrits),
		Entry("below min group value", "FV", []byte{}, ErrInvalidTrits),
		Entry("max trytes value", "MM", []byte{}, ErrInvalidTrits),
		Entry("min trytes value", "NN", []byte{}, ErrInvalidTrits),
		Entry("second group invalid", "Z9TE", []byte{255}, ErrInvalidTrits),
		Entry("third group invalid", "99A9AFV", []byte{0, 1}, ErrInvalidTrits),
	)
})

func MustDecodeHex(s string) []byte {
	dst, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return dst
}
