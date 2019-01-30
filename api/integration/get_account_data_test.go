package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("GetAccountData()", func() {

	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	accountData := AccountData{
		Addresses: SampleAddressesWithChecksum,
		Transfers: Transfers,
		Inputs: []Input{
			{
				Address:  SampleAddressesWithChecksum[2],
				Balance:  1,
				KeyIndex: 2,
				Security: SecurityLevelMedium,
			},
		},
		LatestAddress: SampleAddressesWithChecksum[2],
		Transactions:  nil, // txs addresses (which are 9s) never matched the seed's addresses
		Balance:       1,
	}

	Context("call", func() {
		It("resolves to correct account data", func() {
			ad, err := api.GetAccountData(Seed, GetAccountDataOptions{Start: 0})
			Expect(err).ToNot(HaveOccurred())
			Expect(*ad).To(Equal(accountData))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid seed", func() {
			_, err := api.GetAccountData("asdf", GetAccountDataOptions{Start: 0})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidSeed))
		})

		It("returns an error for invalid start end options", func() {
			var end uint64 = 9
			_, err := api.GetAccountData(Seed, GetAccountDataOptions{Start: 10, End: &end})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidStartEndOptions))
		})
	})

})
