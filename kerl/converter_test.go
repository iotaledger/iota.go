package kerl_test

import (
	"bytes"
	"strings"

	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/kerl"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Converter", func() {

	Context("KerlTritsToBytes()", func() {
		It("should return zero bytes", func() {
			trits := MustTrytesToTrits(strings.Repeat("9", HashTrytesSize))
			bs, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(bs).To(Equal(bytes.Repeat([]byte{0}, HashBytesSize)))
		})

		It("should return bytes for largest", func() {
			trits := MustTrytesToTrits(strings.Repeat("M", HashTrytesSize))
			bs, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(bs).To(Equal([]byte{94, 105, 235, 239, 168, 127, 171, 223, 170, 6, 168, 5, 169, 246, 128, 139, 72, 187, 174, 54, 121, 164, 199, 2, 80, 151, 157, 87, 12, 36, 72, 110, 58, 222, 0, 217, 20, 132, 80, 79, 159, 0, 118, 105, 165, 206, 137, 100}))
		})

		It("should return bytes for smallest", func() {
			trits := MustTrytesToTrits(strings.Repeat("N", HashTrytesSize))
			bs, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(bs).To(Equal([]byte{161, 150, 20, 16, 87, 128, 84, 32, 85, 249, 87, 250, 86, 9, 127, 116, 183, 68, 81, 201, 134, 91, 56, 253, 175, 104, 98, 168, 243, 219, 183, 145, 197, 33, 255, 38, 235, 123, 175, 176, 96, 255, 137, 150, 90, 49, 118, 156}))
		})

		It("should handle internal carries", func() {
			// the following trytes correspond to 2^320 leading to additions with many carries
			trits := MustTrytesToTrits("NNNNNNNNNNNNIPWAK9KOEYFFRZLJXRFLFLBRBFQATTA9TLIDNFNIEMCSPPUHKUGISALJSLL9PSXBQXEPW")
			bs, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(bs).To(Equal([]byte{163, 171, 82, 86, 227, 18, 26, 241, 85, 249, 87, 250, 86, 9, 127, 116, 183, 68, 81, 201, 134, 91, 56, 253, 175, 104, 98, 168, 243, 219, 183, 145, 197, 33, 255, 38, 235, 123, 175, 176, 96, 255, 137, 150, 90, 49, 118, 156}))
		})

		It("should return bytes for valid trits", func() {
			trits := MustTrytesToTrits("9RFAOVEWQDNGBPEGFZTVJKKITBASFWCQBSTZYWTYIJETVZJYNFFIEQ9JMQWEHQ9ZKARYTE9GGDYZHIPJX")
			bs, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(bs).To(Equal([]byte{200, 133, 129, 2, 47, 13, 241, 221, 98, 137, 183, 55, 217, 17, 54, 58, 35, 144, 226, 211, 121, 162, 148, 10, 119, 202, 21, 32, 48, 36, 98, 155, 2, 253, 57, 40, 89, 220, 88, 211, 119, 78, 246, 21, 121, 44, 224, 15}))
		})

		It("should return an error for invalid trits slice length", func() {
			_, err := KerlTritsToBytes(Trits{1, 1})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("KerlTrytesToBytes()", func() {
		It("should return zero bytes", func() {
			bs, err := KerlTrytesToBytes(strings.Repeat("9", HashTrytesSize))
			Expect(err).ToNot(HaveOccurred())
			Expect(bs).To(Equal(bytes.Repeat([]byte{0}, HashBytesSize)))
		})

		It("should return bytes for largest", func() {
			bs, err := KerlTrytesToBytes(strings.Repeat("M", HashTrytesSize))
			Expect(err).ToNot(HaveOccurred())
			Expect(bs).To(Equal([]byte{94, 105, 235, 239, 168, 127, 171, 223, 170, 6, 168, 5, 169, 246, 128, 139, 72, 187, 174, 54, 121, 164, 199, 2, 80, 151, 157, 87, 12, 36, 72, 110, 58, 222, 0, 217, 20, 132, 80, 79, 159, 0, 118, 105, 165, 206, 137, 100}))
		})

		It("should return bytes for smallest", func() {
			bs, err := KerlTrytesToBytes(strings.Repeat("N", HashTrytesSize))
			Expect(err).ToNot(HaveOccurred())
			Expect(bs).To(Equal([]byte{161, 150, 20, 16, 87, 128, 84, 32, 85, 249, 87, 250, 86, 9, 127, 116, 183, 68, 81, 201, 134, 91, 56, 253, 175, 104, 98, 168, 243, 219, 183, 145, 197, 33, 255, 38, 235, 123, 175, 176, 96, 255, 137, 150, 90, 49, 118, 156}))
		})

		It("should handle internal carries", func() {
			// the following trytes correspond to 2^320 leading to additions with many carries
			bs, err := KerlTrytesToBytes("NNNNNNNNNNNNIPWAK9KOEYFFRZLJXRFLFLBRBFQATTA9TLIDNFNIEMCSPPUHKUGISALJSLL9PSXBQXEPW")
			Expect(err).ToNot(HaveOccurred())
			Expect(bs).To(Equal([]byte{163, 171, 82, 86, 227, 18, 26, 241, 85, 249, 87, 250, 86, 9, 127, 116, 183, 68, 81, 201, 134, 91, 56, 253, 175, 104, 98, 168, 243, 219, 183, 145, 197, 33, 255, 38, 235, 123, 175, 176, 96, 255, 137, 150, 90, 49, 118, 156}))
		})

		It("should return bytes for valid trytes", func() {
			bs, err := KerlTrytesToBytes("9RFAOVEWQDNGBPEGFZTVJKKITBASFWCQBSTZYWTYIJETVZJYNFFIEQ9JMQWEHQ9ZKARYTE9GGDYZHIPJX")
			Expect(err).ToNot(HaveOccurred())
			Expect(bs).To(Equal([]byte{200, 133, 129, 2, 47, 13, 241, 221, 98, 137, 183, 55, 217, 17, 54, 58, 35, 144, 226, 211, 121, 162, 148, 10, 119, 202, 21, 32, 48, 36, 98, 155, 2, 253, 57, 40, 89, 220, 88, 211, 119, 78, 246, 21, 121, 44, 224, 15}))
		})

		It("should return an error for invalid trytes slice length", func() {
			_, err := KerlTrytesToBytes("99")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("KerlBytesToTrits()", func() {
		It("should return all 0 trits", func() {
			trits, err := KerlBytesToTrits(bytes.Repeat([]byte{0}, HashBytesSize))
			Expect(err).ToNot(HaveOccurred())
			Expect(trits).To(Equal(MustTrytesToTrits(strings.Repeat("9", HashTrytesSize))))
		})

		It("should return trytes for largest", func() {
			trits, err := KerlBytesToTrits([]byte{127, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255})
			Expect(err).ToNot(HaveOccurred())
			Expect(trits).To(Equal(MustTrytesToTrits("DGKMYULNWJECTMKWJTSDPSPCODNBWDCSOEQRJAEQTTZRKCQ9NZZZTCCVJYXYXCYDVDIMLWF9MTFJDMSCX")))
		})

		It("should return trytes for smallest", func() {
			trits, err := KerlBytesToTrits([]byte{128, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
			Expect(err).ToNot(HaveOccurred())
			Expect(trits).To(Equal(MustTrytesToTrits("VTPNBFOMDQVXGNPDQGHWKHKXLWMYDWXHLVJIQZVJGGAIPXJ9MAAAGXXEQBCBCXBWEWRNODU9NGUQWNHXC")))
		})

		It("should return trits for valid bytes", func() {
			trits, err := KerlBytesToTrits([]byte{200, 133, 129, 2, 47, 13, 241, 221, 98, 137, 183, 55, 217, 17, 54, 58, 35, 144, 226, 211, 121, 162, 148, 10, 119, 202, 21, 32, 48, 36, 98, 155, 2, 253, 57, 40, 89, 220, 88, 211, 119, 78, 246, 21, 121, 44, 224, 15})
			Expect(err).ToNot(HaveOccurred())
			Expect(trits).To(Equal(MustTrytesToTrits("9RFAOVEWQDNGBPEGFZTVJKKITBASFWCQBSTZYWTYIJETVZJYNFFIEQ9JMQWEHQ9ZKARYTE9GGDYZHIPJX")))
		})

		It("should return an error for invalid bytes slice length", func() {
			_, err := KerlBytesToTrits([]byte{1, 45, 62})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("KerlBytesToTrytes()", func() {
		It("should return all 9 trytes", func() {
			ts, err := KerlBytesToTrytes(bytes.Repeat([]byte{0}, HashBytesSize))
			Expect(err).ToNot(HaveOccurred())
			Expect(ts).To(Equal(strings.Repeat("9", HashTrytesSize)))
		})

		It("should return trytes for largest", func() {
			ts, err := KerlBytesToTrytes([]byte{127, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255})
			Expect(err).ToNot(HaveOccurred())
			Expect(ts).To(Equal("DGKMYULNWJECTMKWJTSDPSPCODNBWDCSOEQRJAEQTTZRKCQ9NZZZTCCVJYXYXCYDVDIMLWF9MTFJDMSCX"))
		})

		It("should return trytes for smallest", func() {
			ts, err := KerlBytesToTrytes([]byte{128, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
			Expect(err).ToNot(HaveOccurred())
			Expect(ts).To(Equal("VTPNBFOMDQVXGNPDQGHWKHKXLWMYDWXHLVJIQZVJGGAIPXJ9MAAAGXXEQBCBCXBWEWRNODU9NGUQWNHXC"))
		})

		It("should return trytes for valid bytes", func() {
			ts, err := KerlBytesToTrytes([]byte{200, 133, 129, 2, 47, 13, 241, 221, 98, 137, 183, 55, 217, 17, 54, 58, 35, 144, 226, 211, 121, 162, 148, 10, 119, 202, 21, 32, 48, 36, 98, 155, 2, 253, 57, 40, 89, 220, 88, 211, 119, 78, 246, 21, 121, 44, 224, 15})
			Expect(err).ToNot(HaveOccurred())
			Expect(ts).To(Equal("9RFAOVEWQDNGBPEGFZTVJKKITBASFWCQBSTZYWTYIJETVZJYNFFIEQ9JMQWEHQ9ZKARYTE9GGDYZHIPJX"))
		})

		It("should return an error for invalid bytes slice length", func() {
			_, err := KerlBytesToTrytes([]byte{1, 45, 62})
			Expect(err).To(HaveOccurred())
		})
	})
})
