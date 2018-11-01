package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/gocks"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("StoreTransactions()", func() {

	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	Context("call", func() {
		It("resolves to correct response", func() {
			trytes2, err := api.StoreTransactions(TrytesToStore)
			Expect(err).ToNot(HaveOccurred())
			Expect(trytes2[0]).To(Equal(TrytesToStore))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid trytes", func() {
			_, err := api.StoreTransactions("")
			Expect(errors.Cause(err)).To(Equal(ErrInvalidAttachedTrytes))
		})
	})

})
