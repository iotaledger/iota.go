package kerl_test

import (
	. "github.com/iotaledger/iota.go/kerl"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Converter", func() {

	const trytes = "9RFAOVEWQDNGBPEGFZTVJKKITBASFWCQBSTZYWTYIJETVZJYNFFIEQ9JMQWEHQ9ZKARYTE9GGDYZHIPJX"
	var bytes = []byte{200, 133, 129, 2, 47, 13, 241, 221, 98, 137, 183, 55, 217, 17, 54, 58, 35, 144, 226, 211, 121, 162, 148, 10, 119, 202, 21, 32, 48, 36, 98, 155, 2, 253, 57, 40, 89, 220, 88, 211, 119, 78, 246, 21, 121, 44, 224, 15}

	Context("KerlTritsToBytes()", func() {
		It("should return bytes for valid trits", func() {
			trits := MustTrytesToTrits(trytes)
			bytes, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(Equal(bytes))
		})

		It("should return an error for invalid trits slice length", func() {
			_, err := KerlTritsToBytes(Trits{1, 1})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("KerlTrytesToBytes()", func() {
		It("should return bytes for valid trytes", func() {
			bytes, err := KerlTrytesToBytes(trytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(Equal(bytes))
		})

		It("should return an error for invalid trytes slice length", func() {
			_, err := KerlTrytesToBytes("99")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("KerlBytesToTrits()", func() {
		It("should return trits for valid bytes", func() {
			trits, err := KerlBytesToTrits(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(trits).To(Equal(MustTrytesToTrits(trytes)))
		})

		It("should return an error for invalid bytes slice length", func() {
			_, err := KerlBytesToTrits([]byte{1, 45, 62})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("KerlBytesToTrytes()", func() {
		It("should return trytes for valid bytes", func() {
			ts, err := KerlBytesToTrytes(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(ts).To(Equal(trytes))
		})

		It("should return an error for invalid bytes slice length", func() {
			_, err := KerlBytesToTrytes([]byte{1, 45, 62})
			Expect(err).To(HaveOccurred())
		})
	})
})
