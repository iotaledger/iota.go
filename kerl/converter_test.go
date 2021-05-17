package kerl_test

import (
	"bytes"
	"encoding/hex"

	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/kerl"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Converter", func() {

	Context("KerlBytesZeroLastTrit()", func() {
		It("0 → 0", func() {
			bs := bytes.Repeat([]byte{0}, HashBytesSize)
			KerlBytesZeroLastTrit(bs)
			Expect(bs).To(Equal(bytes.Repeat([]byte{0}, HashBytesSize)))
		})

		It("⌊3²⁴² / 2⌋ → ⌊3²⁴² / 2⌋", func() {
			// in: ⌊3²⁴² / 2⌋
			in := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8964"
			// expected:  ⌊3²⁴² / 2⌋
			expected := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8964"

			bs, err := hex.DecodeString(in)
			Expect(err).ToNot(HaveOccurred())
			KerlBytesZeroLastTrit(bs)
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("-⌊3²⁴² / 2⌋ → -⌊3²⁴² / 2⌋", func() {
			// in: -⌊3²⁴² / 2⌋
			in := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"
			// expected:  -⌊3²⁴² / 2⌋
			expected := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

			bs, err := hex.DecodeString(in)
			Expect(err).ToNot(HaveOccurred())
			KerlBytesZeroLastTrit(bs)
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("-⌊3²⁴² / 2⌋ - 1 → ⌊3²⁴² / 2⌋", func() {
			// in: -⌊3²⁴² / 2⌋ - 1
			in := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769b"
			// expected:  -⌊3²⁴² / 2⌋ - 1 + 3²⁴²
			expected := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8964"

			bs, err := hex.DecodeString(in)
			Expect(err).ToNot(HaveOccurred())
			KerlBytesZeroLastTrit(bs)
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("⌊3²⁴² / 2⌋ + 1 → -⌊3²⁴² / 2⌋", func() {
			// in: ⌊3²⁴² / 2⌋ + 1
			in := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8965"
			// expected:  ⌊3²⁴² / 2⌋ + 1 - 3²⁴²
			expected := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

			bs, err := hex.DecodeString(in)
			Expect(err).ToNot(HaveOccurred())
			KerlBytesZeroLastTrit(bs)
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("⌊3²⁴² / 2⌋ + 1 → -⌊3²⁴² / 2⌋", func() {
			// in: ⌊3²⁴² / 2⌋ + 1
			in := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8965"
			// expected:  ⌊3²⁴² / 2⌋ + 1 - 3²⁴²
			expected := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

			bs, err := hex.DecodeString(in)
			Expect(err).ToNot(HaveOccurred())
			KerlBytesZeroLastTrit(bs)
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("2³⁸³ - 1 → 2³⁸³ - 1 - 3²⁴²", func() {
			// in: 2³⁸³ - 1
			in := "7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
			// expected:  2³⁸³ - 1 - 3²⁴²
			expected := "c32c2820af00a840abf2aff4ac12fee96e88a3930cb671fb5ed0c551e7b76f238a43fe4dd6f75f60c1ff132cb462ed36"

			bs, err := hex.DecodeString(in)
			Expect(err).ToNot(HaveOccurred())
			KerlBytesZeroLastTrit(bs)
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("-2³⁸³ → -2³⁸³ + 3²⁴²", func() {
			// in: -2³⁸³
			in := "800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
			// expected:  -2³⁸³ + 3²⁴²
			expected := "3cd3d7df50ff57bf540d500b53ed011691775c6cf3498e04a12f3aae184890dc75bc01b22908a09f3e00ecd34b9d12c9"

			bs, err := hex.DecodeString(in)
			Expect(err).ToNot(HaveOccurred())
			KerlBytesZeroLastTrit(bs)
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})
	})

	Context("KerlTritsToBytes()", func() {
		It("should return bytes for 0", func() {
			// in: balanced 243-trit representation of 0
			trits := make(Trits, HashTrinarySize)
			// expected: unsigned 384-bit representation of 0
			expected := "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"

			bs, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for 1", func() {
			// in: balanced 243-trit representation of 1
			trits, _ := PadTrits(Trits{1}, HashTrinarySize)
			// expected: unsigned 384-bit representation of 1
			expected := "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"

			bs, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for -1", func() {
			// in: balanced 243-trit representation of -1
			trits, _ := PadTrits(Trits{-1}, HashTrinarySize)
			// expected: unsigned 384-bit representation of -1
			expected := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

			bs, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for all '1's", func() {
			// in: balanced 243-trit representation of ⌊3²⁴³ / 2⌋
			trits := MustTrytesToTrits("MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM")
			// expected: unsigned 384-bit representation of ⌊3²⁴² / 2⌋
			expected := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8964"

			bs, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for all '-1's", func() {
			// in: balanced 243-trit representation of -⌊3²⁴³ / 2⌋
			trits := MustTrytesToTrits("NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN")
			// expected: unsigned 384-bit representation of -⌊3²⁴² / 2⌋
			expected := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

			bs, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for the uint32 chunk size", func() {
			// in: balanced 243-trit representation of 27⁶
			trits := IntToTrits(0x17179149, HashTrinarySize)
			// expected: unsigned 384-bit representation of -⌊3²⁴² / 2⌋
			expected := "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000017179149"

			bs, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for min number with positive last trit", func() {
			// in: balanced 243-trit representation of ⌊3²⁴² / 2⌋ + 1
			trits := MustTrytesToTrits("NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNE")
			// expected: unsigned 384-bit representation of -⌊3²⁴² / 2⌋
			expected := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

			bs, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for max number with negative last trit", func() {
			// in: balanced 243-trit representation of -⌊3²⁴² / 2⌋ - 1
			trits := MustTrytesToTrits("MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMV")
			// expected: unsigned 384-bit representation of ⌊3²⁴² / 2⌋
			expected := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8964"

			bs, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should handle internal carries", func() {
			// the following trytes correspond to 2^320 leading to additions with many carries
			trits := MustTrytesToTrits("NNNNNNNNNNNNIPWAK9KOEYFFRZLJXRFLFLBRBFQATTA9TLIDNFNIEMCSPPUHKUGISALJSLL9PSXBQXEPW")
			expected := "a3ab5256e3121af155f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

			bs, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for valid trits", func() {
			trits := MustTrytesToTrits("9RFAOVEWQDNGBPEGFZTVJKKITBASFWCQBSTZYWTYIJETVZJYNFFIEQ9JMQWEHQ9ZKARYTE9GGDYZHIPJX")
			expected := "c88581022f0df1dd6289b737d911363a2390e2d379a2940a77ca15203024629b02fd392859dc58d3774ef615792ce00f"

			bs, err := KerlTritsToBytes(trits)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return an error for invalid trits slice length", func() {
			_, err := KerlTritsToBytes(Trits{1, 1})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("KerlTrytesToBytes()", func() {
		It("should return bytes for 0", func() {
			// in: balanced 81-tryte representation of 0
			trytes := IntToTrytes(0, HashTrytesSize)
			// expected: unsigned 384-bit representation of 0
			expected := "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"

			bs, err := KerlTrytesToBytes(trytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for 1", func() {
			// in: balanced 81-tryte representation of 1
			trytes := IntToTrytes(1, HashTrytesSize)
			// expected: unsigned 384-bit representation of 1
			expected := "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"

			bs, err := KerlTrytesToBytes(trytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for -1", func() {
			// in: balanced 81-tryte representation of -1
			trytes := IntToTrytes(-1, HashTrytesSize)
			// expected: unsigned 384-bit representation of -1
			expected := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

			bs, err := KerlTrytesToBytes(trytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for all 'M's", func() {
			// in: balanced 81-tryte representation of ⌊3²⁴³ / 2⌋
			trytes := "MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM"
			// expected: unsigned 384-bit representation of ⌊3²⁴² / 2⌋
			expected := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8964"

			bs, err := KerlTrytesToBytes(trytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for all 'N's", func() {
			// in: balanced 81-tryte representation of -⌊3²⁴³ / 2⌋
			trytes := "NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN"
			// expected: unsigned 384-bit representation of -⌊3²⁴² / 2⌋
			expected := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

			bs, err := KerlTrytesToBytes(trytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for the uint32 chunk size", func() {
			// in: balanced 81-tryte representation of 27⁶
			trytes := IntToTrytes(0x17179149, HashTrytesSize)
			// expected: unsigned 384-bit representation of -⌊3²⁴² / 2⌋
			expected := "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000017179149"

			bs, err := KerlTrytesToBytes(trytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for min number with positive last trit", func() {
			// in: balanced 81-tryte representation of ⌊3²⁴² / 2⌋ + 1
			trytes := "NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNE"
			// expected: unsigned 384-bit representation of -⌊3²⁴² / 2⌋
			expected := "a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

			bs, err := KerlTrytesToBytes(trytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for max number with negative last trit", func() {
			// in: balanced 81-tryte representation of -⌊3²⁴² / 2⌋ - 1
			trytes := "MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMV"
			// expected: unsigned 384-bit representation of ⌊3²⁴² / 2⌋
			expected := "5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8964"

			bs, err := KerlTrytesToBytes(trytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should handle internal carries", func() {
			// the following trytes correspond to 2^320 leading to additions with many carries
			trytes := "NNNNNNNNNNNNIPWAK9KOEYFFRZLJXRFLFLBRBFQATTA9TLIDNFNIEMCSPPUHKUGISALJSLL9PSXBQXEPW"
			expected := "a3ab5256e3121af155f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769c"

			bs, err := KerlTrytesToBytes(trytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return bytes for valid trytes", func() {
			trytes := "9RFAOVEWQDNGBPEGFZTVJKKITBASFWCQBSTZYWTYIJETVZJYNFFIEQ9JMQWEHQ9ZKARYTE9GGDYZHIPJX"
			expected := "c88581022f0df1dd6289b737d911363a2390e2d379a2940a77ca15203024629b02fd392859dc58d3774ef615792ce00f"

			bs, err := KerlTrytesToBytes(trytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(hex.EncodeToString(bs)).To(Equal(expected))
		})

		It("should return an error for invalid trytes slice length", func() {
			_, err := KerlTrytesToBytes("99")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("KerlBytesToTrits()", func() {
		It("should return trits for all '0x00's", func() {
			bytes, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
			expected := IntToTrits(0, HashTrinarySize)

			trits, err := KerlBytesToTrits(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(trits).To(Equal(expected))
		})

		It("should return trits for max", func() {
			bytes, _ := hex.DecodeString("800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
			expected := MustTrytesToTrits("VTPNBFOMDQVXGNPDQGHWKHKXLWMYDWXHLVJIQZVJGGAIPXJ9MAAAGXXEQBCBCXBWEWRNODU9NGUQWNHXC")

			trits, err := KerlBytesToTrits(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(trits).To(Equal(expected))
		})

		It("should return trits for min", func() {
			bytes, _ := hex.DecodeString("7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
			expected := MustTrytesToTrits("DGKMYULNWJECTMKWJTSDPSPCODNBWDCSOEQRJAEQTTZRKCQ9NZZZTCCVJYXYXCYDVDIMLWF9MTFJDMSCX")

			trits, err := KerlBytesToTrits(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(trits).To(Equal(expected))
		})

		It("should return trytes for ⌊3²⁴² / 2⌋ + 1", func() {
			bytes, _ := hex.DecodeString("5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8965")
			expected := MustTrytesToTrits("NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNW")

			trits, err := KerlBytesToTrits(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(trits).To(Equal(expected))
		})

		It("should return trytes for -⌊3²⁴² / 2⌋ - 1", func() {
			bytes, _ := hex.DecodeString("a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769b")
			expected := MustTrytesToTrits("MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMD")

			trits, err := KerlBytesToTrits(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(trits).To(Equal(expected))
		})

		It("should return trits for valid bytes", func() {
			bytes, _ := hex.DecodeString("c88581022f0df1dd6289b737d911363a2390e2d379a2940a77ca15203024629b02fd392859dc58d3774ef615792ce00f")
			expected := MustTrytesToTrits("9RFAOVEWQDNGBPEGFZTVJKKITBASFWCQBSTZYWTYIJETVZJYNFFIEQ9JMQWEHQ9ZKARYTE9GGDYZHIPJX")

			trits, err := KerlBytesToTrits(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(trits).To(Equal(expected))
		})

		It("should return an error for invalid bytes slice length", func() {
			_, err := KerlBytesToTrits([]byte{1, 45, 62})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("KerlBytesToTrytes()", func() {
		It("should return trytes for all '0x00's", func() {
			bytes, _ := hex.DecodeString("000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
			expected := IntToTrytes(0, HashTrytesSize)

			trytes, err := KerlBytesToTrytes(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(trytes).To(Equal(expected))
		})

		It("should return trytes for max", func() {
			bytes, _ := hex.DecodeString("800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
			expected := "VTPNBFOMDQVXGNPDQGHWKHKXLWMYDWXHLVJIQZVJGGAIPXJ9MAAAGXXEQBCBCXBWEWRNODU9NGUQWNHXC"

			trytes, err := KerlBytesToTrytes(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(trytes).To(Equal(expected))
		})

		It("should return trytes for min", func() {
			bytes, _ := hex.DecodeString("7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
			expected := "DGKMYULNWJECTMKWJTSDPSPCODNBWDCSOEQRJAEQTTZRKCQ9NZZZTCCVJYXYXCYDVDIMLWF9MTFJDMSCX"

			trytes, err := KerlBytesToTrytes(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(trytes).To(Equal(expected))
		})

		It("should return trytes for ⌊3²⁴² / 2⌋ + 1", func() {
			bytes, _ := hex.DecodeString("5e69ebefa87fabdfaa06a805a9f6808b48bbae3679a4c70250979d570c24486e3ade00d91484504f9f007669a5ce8965")
			expected := "NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNW"

			trytes, err := KerlBytesToTrytes(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(trytes).To(Equal(expected))
		})

		It("should return trytes for -⌊3²⁴² / 2⌋ - 1", func() {
			bytes, _ := hex.DecodeString("a19614105780542055f957fa56097f74b74451c9865b38fdaf6862a8f3dbb791c521ff26eb7bafb060ff89965a31769b")
			expected := "MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMD"

			trytes, err := KerlBytesToTrytes(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(trytes).To(Equal(expected))
		})

		It("should return trytes for valid bytes", func() {
			bytes, _ := hex.DecodeString("c88581022f0df1dd6289b737d911363a2390e2d379a2940a77ca15203024629b02fd392859dc58d3774ef615792ce00f")
			expected := "9RFAOVEWQDNGBPEGFZTVJKKITBASFWCQBSTZYWTYIJETVZJYNFFIEQ9JMQWEHQ9ZKARYTE9GGDYZHIPJX"

			trytes, err := KerlBytesToTrytes(bytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(trytes).To(Equal(expected))
		})

		It("should return an error for invalid bytes slice length", func() {
			_, err := KerlBytesToTrytes([]byte{1, 45, 62})
			Expect(err).To(HaveOccurred())
		})
	})
})
