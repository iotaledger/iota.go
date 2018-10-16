package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/checksum"
	"github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("FindTransactions()", func() {

	var api *API
	BeforeEach(func() {
		a, err := ComposeAPI(HttpClientSettings{}, nil)
		if err != nil {
			panic(err)
		}
		api = a
	})

	hash := "UFKDPIQSIGJCKXWJZXAPPOWGSTCENJERGMUKJOWXQDUVNXRKXMEAJCTTZDEC9DUNXKUXEOBLULCBA9999"
	resp := "IVDAFVTTIKVFUQ9H9CNOUJKUJKXTVXRXLKHEVKAQCVGJNMWLBYQMJPQBAZGGSFTXLDDYGVDVPJOJ99999"

	Context("address query", func() {

		It("resolves to correct response", func() {
			defer gock.Flush()

			gock.New(DefaultLocalIRIURI).
				Post("/").
				MatchType("json").
				JSON(FindTransactionsCommand{
					Command:               FindTransactionsCmd,
					FindTransactionsQuery: FindTransactionsQuery{Addresses: Hashes{hash}},
				}).
				Reply(200).
				JSON(FindTransactionsResponse{Hashes: Hashes{resp}})

			hashes, err := api.FindTransactions(FindTransactionsQuery{Addresses: Hashes{hash}})
			Expect(err).ToNot(HaveOccurred())
			Expect(hashes[0]).To(Equal(resp))
		})

		It("removes the checksum from the query addresses", func() {
			defer gock.Flush()

			gock.New(DefaultLocalIRIURI).
				Post("/").
				MatchType("json").
				JSON(FindTransactionsCommand{
					Command:               FindTransactionsCmd,
					FindTransactionsQuery: FindTransactionsQuery{Addresses: Hashes{hash}},
				}).
				Reply(200).
				JSON(FindTransactionsResponse{Hashes: Hashes{resp}})

			hashWithChecksum, err := checksum.AddChecksum(hash, true, consts.AddressChecksumTrytesSize)
			Expect(err).ToNot(HaveOccurred())

			hashes, err := api.FindTransactions(FindTransactionsQuery{Addresses: Hashes{hashWithChecksum}})
			Expect(err).ToNot(HaveOccurred())
			Expect(hashes[0]).To(Equal(resp))
		})

		It("returns an error for invalid addresses", func() {
			_, err := api.FindTransactions(FindTransactionsQuery{Addresses: Hashes{"balalaika"}})
			Expect(err).To(HaveOccurred())
			_, err = api.FindTransactions(FindTransactionsQuery{Addresses: Hashes{}})
			Expect(err).To(HaveOccurred())
		})

	})

	Context("bundle query", func() {

		It("resolves to correct response", func() {
			defer gock.Flush()

			gock.New(DefaultLocalIRIURI).
				Post("/").
				MatchType("json").
				JSON(FindTransactionsCommand{
					Command:               FindTransactionsCmd,
					FindTransactionsQuery: FindTransactionsQuery{Bundles: Hashes{hash}},
				}).
				Reply(200).
				JSON(FindTransactionsResponse{Hashes: Hashes{resp}})

			hashes, err := api.FindTransactions(FindTransactionsQuery{Bundles: Hashes{hash}})
			Expect(err).ToNot(HaveOccurred())
			Expect(hashes[0]).To(Equal(resp))
		})

		It("returns an error for invalid bundles", func() {
			_, err := api.FindTransactions(FindTransactionsQuery{Bundles: Hashes{"asdf"}})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidHash))
		})
	})

	Context("tag query", func() {

		It("resolves to correct response", func() {
			defer gock.Flush()

			tag := "BENDER999BENDER99BENDER9999"
			gock.New(DefaultLocalIRIURI).
				Post("/").
				MatchType("json").
				JSON(FindTransactionsCommand{
					Command:               FindTransactionsCmd,
					FindTransactionsQuery: FindTransactionsQuery{Tags: []Trytes{tag}},
				}).
				Reply(200).
				JSON(FindTransactionsResponse{Hashes: Hashes{resp}})

			hashes, err := api.FindTransactions(FindTransactionsQuery{Tags: []Trytes{tag}})
			Expect(err).ToNot(HaveOccurred())
			Expect(hashes[0]).To(Equal(resp))
		})

		It("returns an error for invalid tags", func() {
			_, err := api.FindTransactions(FindTransactionsQuery{Tags: []Trytes{"asdf"}})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTag))
		})
	})

	Context("approvees query", func() {

		It("resolves to correct response", func() {
			defer gock.Flush()

			gock.New(DefaultLocalIRIURI).
				Post("/").
				MatchType("json").
				JSON(FindTransactionsCommand{
					Command:               FindTransactionsCmd,
					FindTransactionsQuery: FindTransactionsQuery{Approvees: Hashes{hash}},
				}).
				Reply(200).
				JSON(FindTransactionsResponse{Hashes: Hashes{resp}})

			hashes, err := api.FindTransactions(FindTransactionsQuery{Approvees: Hashes{hash}})
			Expect(err).ToNot(HaveOccurred())
			Expect(hashes[0]).To(Equal(resp))
		})

		It("returns an error for invalid approvees", func() {
			_, err := api.FindTransactions(FindTransactionsQuery{Approvees: Hashes{"asdf"}})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTransactionHash))
		})
	})

})
