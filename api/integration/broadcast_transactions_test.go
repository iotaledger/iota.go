package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("BroadcastTransactions()", func() {
	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	Context("call", func() {
		It("resolves to correct response", func() {
			broadcastedTrytes, err := api.BroadcastTransactions(BundleTrytes...)
			Expect(err).ToNot(HaveOccurred())
			Expect(broadcastedTrytes).To(Equal(BundleTrytes))
		})

		It("returns an error for invalid trytes", func() {
			_, err := api.BroadcastTransactions("balalaika")
			Expect(errors.Cause(err)).To(Equal(ErrInvalidAttachedTrytes))
		})
	})

})
