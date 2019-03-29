package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/gocks"
	"github.com/iotaledger/iota.go/checksum"
	"github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"strings"
)

var _ = Describe("FindTransactions()", func() {
	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	expect := Hashes{strings.Repeat("A", 81), strings.Repeat("B", 81)}

	Context("address query", func() {

		It("resolves to correct response", func() {
			hashes, err := api.FindTransactions(FindTransactionsQuery{Addresses: FindTransactionsByAddressesQuery})
			Expect(err).ToNot(HaveOccurred())
			Expect(hashes).To(Equal(expect))
		})

		It("removes the checksum from the query addresses", func() {
			hashWithChecksum, err := checksum.AddChecksum(FindTransactionsByAddresses[0], true, consts.AddressChecksumTrytesSize)
			Expect(err).ToNot(HaveOccurred())

			hashes, err := api.FindTransactions(FindTransactionsQuery{Addresses: Hashes{hashWithChecksum}})
			Expect(err).ToNot(HaveOccurred())
			Expect(hashes).To(Equal(expect))
		})

		It("returns an error for invalid addresses", func() {
			_, err := api.FindTransactions(FindTransactionsQuery{Addresses: Hashes{"balalaika"}})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidHash))
			_, err = api.FindTransactions(FindTransactionsQuery{Addresses: Hashes{}})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidAddress))
		})

	})

	Context("bundle query", func() {

		It("resolves to correct response", func() {
			hashes, err := api.FindTransactions(FindTransactionsQuery{Bundles: FindTransactionsByBundles})
			Expect(err).ToNot(HaveOccurred())
			Expect(hashes).To(Equal(expect))
		})

		It("returns an error for invalid bundles", func() {
			_, err := api.FindTransactions(FindTransactionsQuery{Bundles: Hashes{"asdf"}})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidHash))
		})
	})

	Context("tag query", func() {

		It("resolves to correct response", func() {
			hashes, err := api.FindTransactions(FindTransactionsQuery{Tags: FindTransactionsByTags})
			Expect(err).ToNot(HaveOccurred())
			Expect(hashes).To(Equal(expect))
		})

		It("returns an error for invalid tags", func() {
			_, err := api.FindTransactions(FindTransactionsQuery{Tags: []Trytes{"asdf"}})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTag))
		})
	})

	Context("approvees query", func() {

		It("resolves to correct response", func() {
			hashes, err := api.FindTransactions(FindTransactionsQuery{Approvees: FindTransactionsByApprovees})
			Expect(err).ToNot(HaveOccurred())
			Expect(hashes).To(Equal(expect))
		})

		It("returns an error for invalid approvees", func() {
			_, err := api.FindTransactions(FindTransactionsQuery{Approvees: Hashes{"asdf"}})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTransactionHash))
		})
	})

})
