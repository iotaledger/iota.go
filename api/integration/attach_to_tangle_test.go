package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("AttachToTangle()", func() {

	var api *API
	var revTrytes, reqTrytes []Trytes
	var trunk, branch Trytes

	BeforeEach(func() {
		a, err := ComposeAPI(HttpClientSettings{}, nil)
		if err != nil {
			panic(err)
		}
		api = a

		trunk = Bundle[len(Bundle)-1].TrunkTransaction
		branch = Bundle[len(Bundle)-1].BranchTransaction

		revTrytes = make([]Trytes, len(BundleTrytes))
		reqTrytes = make([]Trytes, len(BundleTrytes))
		copy(revTrytes, BundleTrytes)
		copy(reqTrytes, BundleTrytes)
		for i, j := 0, len(revTrytes)-1; i < j; i, j = i+1, j-1 {
			revTrytes[i], revTrytes[j] = revTrytes[j], revTrytes[i]
			reqTrytes[i], reqTrytes[j] = reqTrytes[j], reqTrytes[i]
		}
		gock.New(DefaultLocalIRIURI).
			Post("/").
			MatchType("json").
			JSON(AttachToTangleCommand{
				Command:            AttachToTangleCmd,
				TrunkTransaction:   trunk,
				BranchTransaction:  branch,
				Trytes:             revTrytes,
				MinWeightMagnitude: 14,
			}).
			Reply(200).
			JSON(AttachToTangleResponse{Trytes: BundleTrytes})
	})

	Context("call", func() {

		It("resolve to the correct response", func() {
			defer gock.Flush()
			trytes, err := api.AttachToTangle(trunk, branch, 14, reqTrytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(trytes).To(Equal(BundleTrytes))
		})

		It("does not mutate origin trytes", func() {
			defer gock.Flush()
			reqTrytesCopy := make([]Trytes, len(reqTrytes))
			copy(reqTrytesCopy, reqTrytes)
			_, err := api.AttachToTangle(trunk, branch, 14, reqTrytes)
			Expect(err).ToNot(HaveOccurred())
			Expect(reqTrytes).To(Equal(reqTrytesCopy))
		})
	})

	Context("invalid input", func() {
		invalidTrytes := "balalaika"
		It("returns an error for invalid trunk transaction", func() {
			_, err := api.AttachToTangle(invalidTrytes, branch, 14, reqTrytes)
			Expect(err).To(Equal(ErrInvalidTrunkTransaction))
		})

		It("returns an error for invalid branch transaction", func() {
			_, err := api.AttachToTangle(trunk, invalidTrytes, 14, reqTrytes)
			Expect(err).To(Equal(ErrInvalidBranchTransaction))
		})

		It("returns an error for invalid trytes", func() {
			_, err := api.AttachToTangle(trunk, branch, 14, []Trytes{invalidTrytes})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTransactionTrytes))
		})

	})

})
