package address_test

import (
	. "github.com/iotaledger/iota.go/address"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Address", func() {

	const seed = "ZLNM9UHJWKTTDEZOTH9CXDEIFUJQCIACDPJIXPOWBDW9LTBHC9AQRIXTIHYLIIURLZCXNSTGNIVC9ISVB"

	addresses := []Trytes{
		"CLAAFXEY9AHHCSZCXNKDRZEJHIAFVKYORWNOZAGFPAZYNTSLCXUAG9WBSXBRXYEDPVPLXYVDCBCEKRUBD",
		"CDWOADSZWJMDCLYKEDMPIBTYIFAUUAGM9ZQYDKARBUKFXW9LDRQLNG9MI9DGXSOSPDDFFWWJCB9PTGXPW",
		"VVFGHNRFUQEQILXZYUIHWQFUVEEBQCXCUUENADOKRLTVGULYBNMITSYHVRWMAPKPERRLLTC9ELIWSMMMD",
		"IEKMSNDJVIQMLEFUMVQUUFOMI9RWVMJUXKYABPVOYWVMOOVVABYJUKMHZSXDSACYNTEKBCXRJGRJCKXRY",
		"QELVIIRYZZFJSRKMJSDAEOQJRSAWCGMZOGMTBDNJPIOXQTUGMVPCYLWHGHREKDRRVABPULZI9BOWZQPF9",
	}

	addrWithChecksum := "9OUUHSXJCDAMBWCRUJNYDP9UONYTZPSWYAEPUWUDNTBHCINBJY9QBLERA9OKCBJSUUIADQSIVVFNKTPRYKOAXNZYRX"
	checksums := []Trytes{"OHHJEQVCY", "QUGONROEZ", "DEJHGSUFY", "GEBW9UBHZ", "L9ACZQYCA"}

	addressesWithChecksum := make([]Trytes, len(addresses))
	for i := range addresses {
		addressesWithChecksum[i] = addresses[i] + checksums[i]
	}

	Context("Checksum()", func() {
		It("returns the correct checksum", func() {
			for i := 0; i < len(addresses); i++ {
				check, err := Checksum(addresses[i])
				Expect(err).ToNot(HaveOccurred())
				Expect(check).To(Equal(checksums[i]))
			}
		})

		It("returns an error if the address is not 81 trytes long", func() {
			_, err := Checksum("BALALAIKA")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("ValidateAddress()", func() {
		It("should return nil for valid address", func() {
			Expect(ValidAddress(addresses[0])).ToNot(HaveOccurred())
		})

		It("should return nil for valid address with checksum", func() {
			Expect(ValidAddress(addrWithChecksum)).ToNot(HaveOccurred())
		})

		It("should return an error for an invalid address", func() {
			Expect(ValidAddress("BLABLA")).To(HaveOccurred())
		})
	})

	Context("ValidateChecksum()", func() {
		It("should return nil for valid address/checksum", func() {
			Expect(ValidChecksum(addresses[0], checksums[0])).ToNot(HaveOccurred())
		})

		It("should return an error for an invalid address/checksum", func() {
			Expect(ValidChecksum("BLABLA", "DDFDF")).To(HaveOccurred())
		})
	})

	Context("GenerateAddress()", func() {
		It("returns the correct addresses (without checksum)", func() {
			for i := 0; i < len(addresses); i++ {
				address, err := GenerateAddress(seed, uint64(i), SecurityLevelMedium)
				Expect(err).ToNot(HaveOccurred())
				Expect(address).To(Equal(addresses[i]))
			}
		})

		It("returns the correct addresses (with checksum)", func() {
			for i := 0; i < len(addresses); i++ {
				address, err := GenerateAddress(seed, uint64(i), SecurityLevelMedium, true)
				Expect(err).ToNot(HaveOccurred())
				Expect(address).To(Equal(addresses[i] + checksums[i]))
			}
		})
	})

	Context("GenerateAddresses()", func() {
		It("returns the correct addresses (without checksum)", func() {
			addrs, err := GenerateAddresses(seed, 0, 5, SecurityLevelMedium, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(addrs).To(Equal(addresses))
		})

		It("returns the correct addresses (with checksum)", func() {
			addrs, err := GenerateAddresses(seed, 0, 5, SecurityLevelMedium, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(addrs).To(Equal(addressesWithChecksum))
		})
	})

})
