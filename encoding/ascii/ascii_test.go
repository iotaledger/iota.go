package ascii_test

import (
	. "github.com/iotaledger/iota.go/encoding/ascii"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Converter", func() {

	Context("EncodeToTrytes()", func() {
		It("returns the correct trytes representation", func() {
			trytes, err := EncodeToTrytes("IOTA")
			Expect(err).ToNot(HaveOccurred())
			Expect(trytes).To(Equal("SBYBCCKB"))
		})

		It("returns an error for invalid input", func() {
			_, err := EncodeToTrytes("Γιώτα")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("DecodeTrytes()", func() {
		It("returns the correct ascii representation", func() {
			ascii, err := DecodeTrytes("SBYBCCKB")
			Expect(err).ToNot(HaveOccurred())
			Expect(ascii).To(Equal("IOTA"))
		})
	})
})
