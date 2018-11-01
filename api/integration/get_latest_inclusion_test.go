package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("GetLatestInclusion()", func() {

	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	Context("call", func() {
		It("resolves to correct response", func() {
			states, err := api.GetLatestInclusion(DefaultHashes())
			Expect(err).ToNot(HaveOccurred())
			Expect(states[0]).To(BeTrue())
			Expect(states[1]).To(BeFalse())
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid transaction hashes", func() {
			_, err := api.GetLatestInclusion(Hashes{"balalaika"})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTransactionHash))
		})
	})

})
