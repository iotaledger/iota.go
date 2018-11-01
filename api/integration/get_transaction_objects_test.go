package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("FindTransactions()", func() {
	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	Context("call", func() {
		It("resolves to correct response", func() {
			txs, err := api.GetTransactionObjects(Bundle[0].Hash)
			Expect(err).ToNot(HaveOccurred())
			Expect(txs[0]).To(Equal(Bundle[0]))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid hashes", func() {
			_, err = api.GetTransactionObjects("asdf")
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTransactionHash))
		})
	})
})
