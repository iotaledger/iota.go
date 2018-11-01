package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/gocks"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("AttachToTangle()", func() {

	api, err := ComposeAPI(HTTPClientSettings{})
	if err != nil {
		panic(err)
	}

	var reqTrytes []Trytes
	BeforeEach(func() {
		reqTrytes = make([]Trytes, len(BundleTrytes))
		copy(reqTrytes, BundleTrytes)
		for i, j := 0, len(reqTrytes)-1; i < j; i, j = i+1, j-1 {
			reqTrytes[i], reqTrytes[j] = reqTrytes[j], reqTrytes[i]
		}
	})

	Context("call", func() {

		It("resolve to the correct response", func() {
			trytes, err := api.AttachToTangle(TrunkTx, BranchTx, 14, reqTrytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(trytes).To(Equal(BundleTrytes))
		})

		It("does not mutate origin trytes", func() {
			reqTrytesCopy := make([]Trytes, len(reqTrytes))
			copy(reqTrytesCopy, reqTrytes)
			_, err := api.AttachToTangle(TrunkTx, BranchTx, 14, reqTrytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(reqTrytes).To(Equal(reqTrytesCopy))
		})
	})

	Context("invalid input", func() {
		invalidTrytes := "balalaika"
		It("returns an error for invalid trunk transaction", func() {
			_, err := api.AttachToTangle(invalidTrytes, BranchTx, 14, reqTrytes)
			Expect(err).To(Equal(ErrInvalidTrunkTransaction))
		})

		It("returns an error for invalid branch transaction", func() {
			_, err := api.AttachToTangle(TrunkTx, invalidTrytes, 14, reqTrytes)
			Expect(err).To(Equal(ErrInvalidBranchTransaction))
		})

		It("returns an error for invalid trytes", func() {
			_, err := api.AttachToTangle(TrunkTx, BranchTx, 14, []Trytes{invalidTrytes})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTransactionTrytes))
		})

	})

})
