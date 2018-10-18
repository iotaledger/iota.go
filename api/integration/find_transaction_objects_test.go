package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	_ "github.com/iotaledger/iota.go/api/integration/gocks"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("FindTransactionObjects()", func() {
	api, err := ComposeAPI(HttpClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	Context("call", func() {
		It("resolves to correct response", func() {
			txs, err := api.FindTransactionObjects(FindTransactionsQuery{
				Addresses: Hashes{Bundle[0].Address},
			})
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
