package checksum_test

import (
	. "github.com/iotaledger/iota.go/checksum"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Checksum", func() {
	addrs := []Trytes{
		"GEXLJVJNKFPRGZSTOEVTODEEUJDQCFWOSLVBVVMTRVESTCCTPKILEADWUGMMZVUG9YTJSKNYQUNCSCBDY",
		"JSQITWMZVFUMJONOPSG9TG9SAWXODGVRHCZLYLGNYBUOWIBLDILU9FYGONNPSEQVLLJWB9D9IYCLTXJND",
		"STDNQP9USJOEZDFIMDMIRUVHDFDFUGJSTCDZGJFBBFOQTDPZAMLBYPWFXPDHWDLBUAKWLGTFQZWEFYKEB",
	}

	checksum := []Trytes{"WDUCWPRQW", "JEADZTWUW", "DBKEDYNQC"}

	addrsWithChecksums := make([]Trytes, 3)
	for i := range addrsWithChecksums {
		addrsWithChecksums[i] = addrs[i] + checksum[i]
	}

	Context("AddChecksum()", func() {
		DescribeTable("adds the correct checksum given an address",
			func(addrs Hash, expected Hash) {
				addrWithChecksum, err := AddChecksum(addrs, true, AddressChecksumTrytesSize)
				Expect(err).ToNot(HaveOccurred())
				Expect(addrWithChecksum).To(Equal(addrs + expected))
			},
			Entry("address #1", addrs[0], checksum[0]),
			Entry("address #2", addrs[1], checksum[1]),
			Entry("address #3", addrs[2], checksum[2]),
		)

		It("return an error for invalid address trytes", func() {
			_, err := AddChecksum("", true, AddressChecksumTrytesSize)
			Expect(err).To(HaveOccurred())
		})

		It("return an error for invalid checksum lengths", func() {
			_, err := AddChecksum("", false, 2)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("AddChecksums()", func() {
		It("adds the correct checksum given some addresses", func() {
			withChecksums, err := AddChecksums(addrs, true, AddressChecksumTrytesSize)
			Expect(err).ToNot(HaveOccurred())
			Expect(withChecksums).To(Equal(addrsWithChecksums))
		})
	})

	Context("RemoveChecksum()", func() {
		DescribeTable("removes the checksum given an address",
			func(addrs Hash) {
				addrWithoutChecksum, err := RemoveChecksum(addrs)
				Expect(err).ToNot(HaveOccurred())
				Expect(addrWithoutChecksum).To(Equal(addrs[:HashTrytesSize]))
			},
			Entry("address #1", addrs[0]+checksum[0]),
			Entry("address #2", addrs[1]+checksum[1]),
			Entry("address #3", addrs[2]+checksum[2]),
		)
	})

	Context("RemoveChecksums()", func() {
		It("remove the checksum given some addresses", func() {
			withoutChecksums, err := RemoveChecksums(addrsWithChecksums)
			Expect(err).ToNot(HaveOccurred())
			Expect(withoutChecksums).To(Equal(addrs))
		})
	})

})
