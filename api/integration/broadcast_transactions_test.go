package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	"github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("BroadcastTransactions()", func() {
	var api *API
	a, err := ComposeAPI(HttpClientSettings{}, nil)
	if err != nil {
		panic(err)
	}
	api = a

	Context("call", func() {
		It("resolves to correct response", func() {
			defer gock.Flush()

			gock.New(DefaultLocalIRIURI).
				Post("/").
				MatchType("json").
				JSON(BroadcastTransactionsCommand{Command: BroadcastTransactionsCmd, Trytes: BundleTrytes}).
				Reply(200)

			broadcastedTrytes, err := api.BroadcastTransactions(BundleTrytes...)
			Expect(err).ToNot(HaveOccurred())
			Expect(broadcastedTrytes).To(Equal(BundleTrytes))
		})

		It("returns an error for invalid trytes", func() {
			_, err := api.BroadcastTransactions("balalaika")
			Expect(errors.Cause(err)).To(Equal(consts.ErrInvalidAttachedTrytes))
		})
	})

})
