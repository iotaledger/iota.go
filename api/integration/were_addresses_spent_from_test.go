package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("WereAddressesSpentFrom()", func() {

	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	Context("call", func() {
		It("resolves to correct response", func() {
			spent, err := api.WereAddressesSpentFrom(SampleAddressesWithChecksum...)
			Expect(err).ToNot(HaveOccurred())
			Expect(spent[0]).To(BeTrue())
			Expect(spent[1]).To(BeFalse())
			Expect(spent[2]).To(BeFalse())
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
