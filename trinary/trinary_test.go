package trinary_test

import (
	"strings"

	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Trinary", func() {

	Context("ValidTrit()", func() {

		It("should return true for valid trits", func() {
			Expect(ValidTrit(-1)).To(BeTrue())
			Expect(ValidTrit(1)).To(BeTrue())
			Expect(ValidTrit(1)).To(BeTrue())
		})

		It("should return false for invalid trits", func() {
			Expect(ValidTrit(2)).To(BeFalse())
			Expect(ValidTrit(-2)).To(BeFalse())
		})
	})

	Context("ValidTrits()", func() {
		It("should not return an error for valid trits", func() {
			Expect(ValidTrits(Trits{0, -1, 1, -1, 0, 0, 1, 1})).NotTo(HaveOccurred())
		})

		It("should return an error for invalid trits", func() {
			Expect(ValidTrits(Trits{-1, 0, 3, -1, 0, 0, 1})).To(HaveOccurred())
		})
	})

	Context("NewTrits()", func() {
		It("should return trits and no error with valid trits", func() {
			trits, err := NewTrits([]int8{0, 0, 0, 0, -1, 1, 1, 0})
			Expect(trits).To(Equal([]int8{0, 0, 0, 0, -1, 1, 1, 0}))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return an error for invalid trits", func() {
			_, err := NewTrits([]int8{122, 0, -1, 60, -10, -50})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("TritsEqual()", func() {
		It("should return true for equal trits", func() {
			a := Trits{0, 1, 0}
			b := Trits{0, 1, 0}
			equal, err := TritsEqual(a, b)
			Expect(equal).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return false for unequal trits", func() {
			a := Trits{0, 1, 0}
			b := Trits{1, 0, 0}
			equal, err := TritsEqual(a, b)
			Expect(equal).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return an error for invalid trits", func() {
			a := Trits{120, 50, -33}
			equal, err := TritsEqual(a, a)
			Expect(equal).To(BeFalse())
			Expect(err).To(HaveOccurred())
		})
	})

	Context("IntToTrits()", func() {
		It("should return correct trits representation for positive int64", func() {
			Expect(IntToTrits(12)).To(Equal(Trits{0, 1, 1}))
			Expect(IntToTrits(2)).To(Equal(Trits{-1, 1}))
			Expect(IntToTrits(3332727)).To(Equal(Trits{0, 0, 1, -1, 0, -1, 0, 0, 1, 1, -1, 1, 0, -1, 1}))
			Expect(IntToTrits(0)).To(Equal(Trits{0}))
		})

		It("should return correct trits representation for negative int64", func() {
			Expect(IntToTrits(-7)).To(Equal(Trits{-1, 1, -1}))
			Expect(IntToTrits(-1094385)).To(Equal(Trits{0, -1, 1, 0, 1, -1, -1, 1, 1, 1, -1, 0, 1, -1}))
		})
	})

	Context("TritsToInt", func() {
		It("should return correct nums for positive trits", func() {
			Expect(TritsToInt(Trits{0, 1, 1})).To(Equal(int64(12)))
			Expect(TritsToInt(Trits{-1, 1})).To(Equal(int64(2)))
			Expect(TritsToInt(Trits{0, 0, 1, -1, 0, -1, 0, 0, 1, 1, -1, 1, 0, -1, 1})).To(Equal(int64(3332727)))
			Expect(TritsToInt(Trits{0})).To(Equal(int64(0)))
		})

		It("should return correct nums for negative trits", func() {
			Expect(TritsToInt(Trits{-1, 1, -1})).To(Equal(int64(-7)))
			Expect(TritsToInt(Trits{0, -1, 1, 0, 1, -1, -1, 1, 1, 1, -1, 0, 1, -1})).To(Equal(int64(-1094385)))
		})
	})

	Context("CanTritsToTrytes()", func() {
		It("returns true for valid lengths", func() {
			Expect(CanTritsToTrytes(Trits{1, 1, 1})).To(BeTrue())
			Expect(CanTritsToTrytes(Trits{1, 1, 1, 1, 1, 1})).To(BeTrue())
		})

		It("returns false for invalid lengths", func() {
			Expect(CanTritsToTrytes(Trits{1, 1})).To(BeFalse())
			Expect(CanTritsToTrytes(Trits{1, 1, 1, 1})).To(BeFalse())
		})

		It("returns false for empty trits slices", func() {
			Expect(CanTritsToTrytes(Trits{})).To(BeFalse())
		})
	})

	Context("TrailingZeros()", func() {
		It("should return count of zeroes", func() {
			Expect(TrailingZeros(Trits{1, 0, 0, 0})).To(Equal(int64(3)))
			Expect(TrailingZeros(Trits{0, 0, 0, 0})).To(Equal(int64(4)))
		})
	})

	Context("TritsToTrytes()", func() {
		It("should return trytes and no errors for valid trits", func() {
			trytes, err := TritsToTrytes(Trits{1, 1, 1})
			Expect(trytes).To(Equal("M"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return an error for invalid trits slice length", func() {
			_, err := TritsToTrytes(Trits{1, 1})
			Expect(err).To(HaveOccurred())
		})

		It("should return an error for invalid trits", func() {
			_, err := TritsToTrytes(Trits{12, -45})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("MustTritsToTrytes()", func() {
		It("should return trytes and not panic for valid trits", func() {
			trytes := MustTritsToTrytes(Trits{1, 1, 1})
			Expect(trytes).To(Equal("M"))
		})
		It("should panic with invalid trits", func() {
			Expect(func() { MustTritsToTrytes(Trits{12, -45}) }).To(Panic())
		})
	})

	Context("CanBeHash()", func() {
		It("should return true for a valid trits slice", func() {
			Expect(CanBeHash(make(Trits, HashTrinarySize))).To(BeTrue())
		})
		It("should return false for an invalid trits slice", func() {
			Expect(CanBeHash(make(Trits, 100))).To(BeFalse())
			Expect(CanBeHash(make(Trits, 250))).To(BeFalse())
		})
	})

	Context("TritsToBytes()", func() {
		It("should return bytes for valid trits", func() {
			trits := Trits{-1, 0, 1, 1, 0, 1, 1, 0, -1, -1, 1, 0, 1}
			bytes := TritsToBytes(trits)
			Expect(bytes).To(Equal([]byte{0x23, 0x98, 0x0A}))
		})
	})

	Context("BytesToTrits()", func() {
		It("should return trits for valid bytes", func() {
			expTrits := Trits{-1, 0, 1, 1, 0, 1, 1, 0, -1, -1, 1, 0, 1}
			trits, err := BytesToTrits([]byte{0x23, 0x98, 0x0A}, len(expTrits))
			Expect(trits).To(Equal(expTrits))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return an error for invalid numTrits length", func() {
			_, err := BytesToTrits([]byte{0x23, 0x98, 0x0A}, 10)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("ReverseTrits()", func() {
		It("should correctly reverse trits", func() {
			rev := ReverseTrits(Trits{1, 0, -1})
			Expect(rev).To(Equal(Trits{-1, 0, 1}))
		})

		It("should return an empty trits slice for empty trits slice", func() {
			rev := ReverseTrits(Trits{})
			Expect(rev).To(Equal(Trits{}))
		})
	})

	Context("ValidTryte()", func() {
		It("should return true for valid tryte", func() {
			Expect(ValidTryte('A')).ToNot(HaveOccurred())
			Expect(ValidTryte('X')).ToNot(HaveOccurred())
			Expect(ValidTryte('F')).ToNot(HaveOccurred())
		})

		It("should return false for invalid tryte", func() {
			Expect(ValidTryte('a')).To(HaveOccurred())
			Expect(ValidTryte('x')).To(HaveOccurred())
			Expect(ValidTryte('f')).To(HaveOccurred())
		})
	})

	Context("ValidTrytes()", func() {
		It("should not return any error for valid trytes", func() {
			Expect(ValidTrytes("AAA")).ToNot(HaveOccurred())
			Expect(ValidTrytes("XXX")).ToNot(HaveOccurred())
			Expect(ValidTrytes("FFF")).ToNot(HaveOccurred())
		})

		It("should return an error for invalid trytes", func() {
			Expect(ValidTrytes("f")).To(HaveOccurred())
			Expect(ValidTrytes("xx")).To(HaveOccurred())
			Expect(ValidTrytes("203984")).To(HaveOccurred())
			Expect(ValidTrytes("")).To(HaveOccurred())
		})
	})

	Context("NewTrytes()", func() {
		It("should return trytes for valid string input", func() {
			trytes, err := NewTrytes("BLABLABLA")
			Expect(trytes).To(Equal("BLABLABLA"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return an error for invalid string input", func() {
			_, err := NewTrytes("abcd")
			Expect(err).To(HaveOccurred())
		})

		It("should return an error for empty string input", func() {
			_, err := NewTrytes("")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("TrytesToTrits()", func() {
		It("should return trits for valid trytes", func() {
			trits, err := TrytesToTrits("M")
			Expect(trits).To(Equal(Trits{1, 1, 1}))
			Expect(err).ToNot(HaveOccurred())
			trits, err = TrytesToTrits("O")
			Expect(trits).To(Equal(Trits{0, -1, -1}))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return an error for empty trytes", func() {
			_, err := TrytesToTrits("")
			Expect(err).To(HaveOccurred())
		})

		It("should return an error for invalid trytes", func() {
			_, err := TrytesToTrits("abcd")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("AbsInt64()", func() {
		It("should return only positive values", func() {
			v := MustAbsInt64(-9223372036854775807)
			Expect(v).To(Equal(int64(9223372036854775807)))
		})

		It("should panic for Int64_Min", func() {
			Expect(func() { MustAbsInt64(-9223372036854775808) }).To(Panic())
		})

	})

	Context("MinTrits()", func() {
		It("should return correct length", func() {
			v := MinTrits(1)
			Expect(v).To(Equal(uint64(1)))

			v = MinTrits(4)
			Expect(v).To(Equal(uint64(2)))
		})
	})

	Context("EncodedLength()", func() {
		It("should return correct length", func() {
			v := EncodedLength(-1)
			Expect(v).To(Equal(uint64(4)))

			v = EncodedLength(-4)
			Expect(v).To(Equal(uint64(4)))
		})
	})

	Context("TrytesToBytes()", func() {
		It("should return correct bytes", func() {
			v, err := TrytesToBytes("ABCD")
			Expect(err).ToNot(HaveOccurred())
			Expect(v).To(Equal([]byte{55, 90, 1}))

			v, err = TrytesToBytes("GKLVNYXFYXUDURQE9HQKEAETVRREBKKTL9TOKZRICQMRNTVODHSKGVSFRLFAGGMPRSSMJ9ROBWUZZ9999")
			Expect(err).ToNot(HaveOccurred())
			Expect(v).To(Equal([]byte{61, 118, 137, 173, 99, 228, 102, 238, 163, 5, 199, 96, 32, 190, 209, 247, 189, 100, 65, 35, 193, 42, 254, 78, 232, 38, 136, 101, 44, 73, 46, 103, 182, 6, 35, 11, 209, 121, 171, 19, 119, 3, 166, 180, 201, 228, 0, 0, 0}))
		})
	})
	Context("BytesToTrytes()", func() {
		It("should return correct trytes", func() {
			v, err := BytesToTrytes([]byte{61, 118, 137, 173, 99, 228, 102, 238, 163, 5, 199, 96, 32, 190, 209, 247, 189, 100, 65, 35, 193, 42, 254, 78, 232, 38, 136, 101, 44, 73, 46, 103, 182, 6, 35, 11, 209, 121, 171, 19, 119, 3, 166, 180, 201, 228, 0, 0, 0}, 81)
			Expect(err).ToNot(HaveOccurred())
			Expect(v).To(Equal("GKLVNYXFYXUDURQE9HQKEAETVRREBKKTL9TOKZRICQMRNTVODHSKGVSFRLFAGGMPRSSMJ9ROBWUZZ9999"))
		})
	})

	Context("IntToTrytes()", func() {
		It("should return correct trytes", func() {
			v := IntToTrytes(-1, 1)
			Expect(v).To(Equal("Z"))

			v = IntToTrytes(500, 5)
			Expect(v).To(Equal("NSA99"))
		})
	})

	Context("TrytesToInt()", func() {
		It("should return correct int", func() {
			v := TrytesToInt("ABCD")
			Expect(v).To(Equal(int64(80974)))

			v = TrytesToInt("ABCDEFGH")
			Expect(v).To(Equal(int64(86483600668)))
		})
	})

	Context("EncodeInt64()", func() {
		It("should return correct trits", func() {
			v, s, err := EncodeInt64(6)
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(Equal(uint64(4)))
			Expect(v).To(Equal(Trits{0, -1, 1, 0}))

			v, s, err = EncodeInt64(500)
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(Equal(uint64(12)))
			Expect(v).To(Equal(Trits{-1, -1, -1, 1, 0, -1, 1, 0, 0, 0, 0, 0}))
		})
	})

	Context("DecodeInt64()", func() {
		It("should return correct int64", func() {
			v, s, err := DecodeInt64(Trits{0, -1, 1, 0, -1, 1})
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(Equal(uint64(4)))
			Expect(v).To(Equal(int64(6)))

			v, s, err = DecodeInt64(Trits{-1, -1, -1, 1, 0, -1, 1, 0, 0, 0, 0, 0, -1, 1, 0, 1, 0, 0, 0})
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(Equal(uint64(12)))
			Expect(v).To(Equal(int64(500)))
		})
	})

	Context("MustTrytesToTrits()", func() {
		It("should return trits for valid trytes", func() {
			trits := MustTrytesToTrits("M")
			Expect(trits).To(Equal(Trits{1, 1, 1}))
			trits = MustTrytesToTrits("O")
			Expect(trits).To(Equal(Trits{0, -1, -1}))
		})

		It("should panic for emptry trytes", func() {
			Expect(func() { MustTrytesToTrits("") }).To(Panic())
		})

		It("should panic for invalid trytes", func() {
			Expect(func() { MustTrytesToTrits("abcd") }).To(Panic())
		})
	})

	Context("Pad()", func() {
		It("should pad up to the given size", func() {
			Expect(Pad("A", 5)).To(Equal("A9999"))
			Expect(Pad("", 81)).To(Equal(strings.Repeat("9", 81)))
		})
	})

	Context("PadTrits()", func() {
		It("should pad up to the given size", func() {
			Expect(PadTrits(Trits{}, 5)).To(Equal(Trits{0, 0, 0, 0, 0}))
			Expect(PadTrits(Trits{1, 1}, 5)).To(Equal(Trits{1, 1, 0, 0, 0}))
			Expect(PadTrits(Trits{1, -1, 0, 1}, 5)).To(Equal(Trits{1, -1, 0, 1, 0}))
		})
	})

	Context("AddTrits()", func() {
		It("should correctly add trits together (positive)", func() {
			Expect(TritsToInt(AddTrits(IntToTrits(5), IntToTrits(5)))).To(Equal(int64(10)))
			Expect(TritsToInt(AddTrits(IntToTrits(0), IntToTrits(0)))).To(Equal(int64(0)))
			Expect(TritsToInt(AddTrits(IntToTrits(-100), IntToTrits(-20)))).To(Equal(int64(-120)))
		})
	})
})
