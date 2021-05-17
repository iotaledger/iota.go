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
			Expect(IntToTrits(12, MinTrits(12))).To(Equal(Trits{0, 1, 1}))
			Expect(IntToTrits(2, MinTrits(2))).To(Equal(Trits{-1, 1}))
			Expect(IntToTrits(3332727, MinTrits(3332727))).To(Equal(Trits{0, 0, 1, -1, 0, -1, 0, 0, 1, 1, -1, 1, 0, -1, 1}))
			Expect(IntToTrits(0, MinTrits(0))).To(Equal(Trits{0}))
		})

		It("should return correct trits representation for negative int64", func() {
			Expect(IntToTrits(-7, MinTrits(-7))).To(Equal(Trits{-1, 1, -1}))
			Expect(IntToTrits(-1094385, MinTrits(-1094385))).To(Equal(Trits{0, -1, 1, 0, 1, -1, -1, 1, 1, 1, -1, 0, 1, -1}))
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
			Expect(TrailingZeros(Trits{1, 0, 0, 0})).To(Equal(3))
			Expect(TrailingZeros(Trits{0, 0, 0, 0})).To(Equal(4))
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

	Context("MinTrits()", func() {
		It("should return correct length", func() {
			v := MinTrits(1)
			Expect(v).To(Equal(1))

			v = MinTrits(4)
			Expect(v).To(Equal(2))
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

	Context("MustTrytesToTrits()", func() {
		It("should return trits for valid trytes", func() {
			trits := MustTrytesToTrits("M")
			Expect(trits).To(Equal(Trits{1, 1, 1}))
			trits = MustTrytesToTrits("O")
			Expect(trits).To(Equal(Trits{0, -1, -1}))
		})

		It("should panic for invalid trytes", func() {
			Expect(func() { MustTrytesToTrits("abcd") }).To(Panic())
		})
	})

	Context("MustPad()", func() {
		It("should pad up to the given size", func() {
			Expect(MustPad("A", 5)).To(Equal("A9999"))
			Expect(MustPad("", 81)).To(Equal(strings.Repeat("9", 81)))
		})
	})

	Context("MustPadTrits()", func() {
		It("should pad up to the given size", func() {
			Expect(MustPadTrits(Trits{}, 5)).To(Equal(Trits{0, 0, 0, 0, 0}))
			Expect(MustPadTrits(Trits{1, 1}, 5)).To(Equal(Trits{1, 1, 0, 0, 0}))
			Expect(MustPadTrits(Trits{1, -1, 0, 1}, 5)).To(Equal(Trits{1, -1, 0, 1, 0}))
		})
	})

	Context("AddTrits()", func() {
		It("should correctly add trits together (positive)", func() {
			Expect(TritsToInt(AddTrits(IntToTrits(5, MinTrits(5)), IntToTrits(5, MinTrits(5))))).To(Equal(int64(10)))
			Expect(TritsToInt(AddTrits(IntToTrits(0, MinTrits(0)), IntToTrits(0, MinTrits(0))))).To(Equal(int64(0)))
			Expect(TritsToInt(AddTrits(IntToTrits(-100, MinTrits(-100)), IntToTrits(-20, MinTrits(-20))))).To(Equal(int64(-120)))
		})
	})
})
