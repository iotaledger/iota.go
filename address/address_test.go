package address_test

import (
	. "github.com/iotaledger/iota.go/address"
	"github.com/iotaledger/iota.go/checksum"
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

	migrationAddress := "TRANSFERCDJWLVPAIXRWNAPXV9WYKVUZWWKXVBE9JBABJ9D9C9F9OEGADYO9CWDAGZHBRWIXLXG9MAJV9"
	migrationAddressWithChecksum, _ := checksum.AddChecksum(migrationAddress, true, AddressChecksumTrytesSize)
	ed25519Addr := [32]byte{111, 158, 133, 16, 184, 139, 14, 164, 251, 198, 132, 223, 144, 186, 49, 5, 64, 55, 10, 4, 3, 6, 123, 34, 206, 244, 151, 31, 236, 62, 139, 184}

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

	Context("GenerateMigrationAddress", func() {
		It("returns the correct migration address", func() {
			addr, err := GenerateMigrationAddress(ed25519Addr)
			Expect(err).NotTo(HaveOccurred())
			Expect(addr).To(Equal(migrationAddress))
		})
		It("returns the correct migration address with checksum", func() {
			addr, err := GenerateMigrationAddress(ed25519Addr, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(addr).To(Equal(migrationAddressWithChecksum))
		})
	})

	Context("ParseMigrationAddress", func() {
		It("returns an error for a migration address with a wrong prefix", func() {
			addr := "WILDLIFECDJWLVPAIXRWNAPXV9WYKVUZWWKXVBE9JBABJ9D9C9F9OEGADYO9CWDAGZHBRWIXLXG9MAJV9"
			_, err := ParseMigrationAddress(addr)
			Expect(err).To(HaveOccurred())
		})
		It("returns an error for a migration address with a non '9' tryte as the suffix", func() {
			addr := "TRANSFERCDJWLVPAIXRWNAPXV9WYKVUZWWKXVBE9JBABJ9D9C9F9OEGADYO9CWDAGZHBRWIXLXG9MAJVA"
			_, err := ParseMigrationAddress(addr)
			Expect(err).To(HaveOccurred())
		})
		It("returns an error for a migration address with an invalid Ed25519 checksum", func() {
			addr := "TRANSFERCDJWLVPAIXRWNAPXV9WYKVUZWWKXVBE9JBABJ9D9C9F9OEGADYO9CWDAGZHBRWIXLXG9MAJZ9"
			_, err := ParseMigrationAddress(addr)
			Expect(err).To(HaveOccurred())
		})
		It("parses a valid migration address", func() {
			parsed, err := ParseMigrationAddress(migrationAddress)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed).To(Equal(ed25519Addr))
		})
		It("parses a valid migration address with checksum", func() {
			parsed, err := ParseMigrationAddress(migrationAddressWithChecksum)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed).To(Equal(ed25519Addr))
		})
	})

})
