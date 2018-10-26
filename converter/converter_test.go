package converter_test

import (
	. "github.com/iotaledger/iota.go/converter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Converter", func() {

	Context("ASCIIToTrytes()", func() {
		It("returns the correct trytes representation", func() {
			trytes, err := ASCIIToTrytes("IOTA")
			Expect(err).ToNot(HaveOccurred())
			Expect(trytes).To(Equal("SBYBCCKB"))
		})

		It("returns an error for invalid input", func() {
			_, err := ASCIIToTrytes("Γιώτα")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("TrytesToASCII()", func() {
		It("returns the correct ascii representation", func() {
			ascii, err := TrytesToASCII("SBYBCCKB")
			Expect(err).ToNot(HaveOccurred())
			Expect(ascii).To(Equal("IOTA"))
		})

		It("returns an error for invalid trytes", func() {
			_, err := TrytesToASCII("AAAfasds")
			Expect(err).To(HaveOccurred())
		})
	})
})
