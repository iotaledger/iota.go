package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	"strings"

	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("GetBalances()", func() {

	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	Context("call", func() {

		It("resolves to correct response", func() {
			balances, err := api.GetBalances(SampleAddressesWithChecksum, 100)
			Expect(err).ToNot(HaveOccurred())
			Expect(*balances).To(Equal(Balances{
				Balances:       []uint64{99, 0, 1},
				Milestone:      strings.Repeat("M", 81),
				MilestoneIndex: 1,
			}))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid addresses", func() {
			_, err := api.GetBalances(Hashes{"balalaika"}, 100)
			Expect(errors.Cause(err)).To(Equal(ErrInvalidHash))
		})

		It("returns an error for invalid threshold", func() {
			_, err := api.GetBalances(Hashes{SampleAddressesWithChecksum[0]}, 101)
			Expect(errors.Cause(err)).To(Equal(ErrInvalidThreshold))
		})
	})

})
