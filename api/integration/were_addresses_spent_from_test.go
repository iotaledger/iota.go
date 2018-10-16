package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/checksum"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("WereAddressesSpentFrom()", func() {

	var api *API
	BeforeEach(func() {
		a, err := ComposeAPI(HttpClientSettings{}, nil)
		if err != nil {
			panic(err)
		}
		api = a
	})

	addr := "MHERPFOFEKSBQ9L9RSFRWKDXGWUIWDGVSI99YL9E9ITXZXHLBTVAF9TOGCYGAOIIHZK9RR9ZDIS9ZYIYX"

	Context("call", func() {

		BeforeEach(func() {
			gock.New(DefaultLocalIRIURI).
				Post("/").
				MatchType("json").
				JSON(WereAddressesSpentFromCommand{
					Command:   WereAddressesSpentFromCmd,
					Addresses: Hashes{addr},
				}).
				Reply(200).
				JSON(WereAddressesSpentFromResponse{
					States: []bool{true},
				})
		})

		It("resolves to correct response", func() {
			defer gock.Flush()
			spent, err := api.WereAddressesSpentFrom(addr)
			Expect(err).ToNot(HaveOccurred())
			Expect(spent[0]).To(BeTrue())
		})

		It("removes checksum from addresses", func() {
			defer gock.Flush()
			addrWithChecksum, err := checksum.AddChecksum(addr, true, AddressChecksumTrytesSize)
			Expect(err).ToNot(HaveOccurred())
			spent, err := api.WereAddressesSpentFrom(addrWithChecksum)
			Expect(err).ToNot(HaveOccurred())
			Expect(spent[0]).To(BeTrue())
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid addresses", func() {
			_, err := api.StoreTransactions("")
			Expect(errors.Cause(err)).To(Equal(ErrInvalidAttachedTrytes))
			_, err = api.StoreTransactions("balalaika")
			Expect(errors.Cause(err)).To(Equal(ErrInvalidAttachedTrytes))
		})
	})

})
