package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("GetBundlesFromAddresses()", func() {
	api, err := ComposeAPI(HTTPClientSettings{})
	if err != nil {
		panic(err)
	}

	Context("call", func() {
		It("resolves to correct response", func() {
			addresses := make(Hashes, len(SampleAddressesWithChecksum))
			copy(addresses, SampleAddressesWithChecksum)
			bndls, err := api.GetBundlesFromAddresses(addresses, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(bndls)).To(Equal(2))
			Expect(bndls).To(Equal(Transfers))
		})

	})

	Context("invalid input", func() {
		It("returns an error for invalid trytes", func() {
			_, err := api.SendTrytes([]Trytes{"asdf"}, 3, 14)
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTransactionTrytes))
		})
	})

})
