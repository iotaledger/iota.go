package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/gocks"
	. "github.com/iotaledger/iota.go/consts"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("GetTransactionsToApprove()", func() {
	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	Context("call", func() {

		It("resolves to correct response", func() {
			gtta, err := api.GetTransactionsToApprove(3)
			Expect(err).ToNot(HaveOccurred())
			Expect(*gtta).To(Equal(TransactionsToApprove{
				TrunkTransaction:  TrunkTx,
				BranchTransaction: BranchTx,
			}))
		})

		It("resolves to correct response with reference option", func() {
			gtta, err := api.GetTransactionsToApprove(3, strings.Repeat("R", 81))
			Expect(err).ToNot(HaveOccurred())
			Expect(*gtta).To(Equal(TransactionsToApprove{
				TrunkTransaction:  BranchTx,
				BranchTransaction: TrunkTx,
			}))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid reference", func() {
			_, err := api.GetTransactionsToApprove(3, "")
			Expect(errors.Cause(err)).To(Equal(ErrInvalidReferenceHash))
		})
	})

})
