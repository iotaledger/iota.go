package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("GetTransfers()", func() {

	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	Context("call", func() {
		It("resolves to correct response", func() {
			trnsfs, err := api.GetTransfers(Seed, GetTransfersOptions{Start: 0, InclusionStates: true})
			Expect(err).ToNot(HaveOccurred())
			Expect(trnsfs).To(Equal(Transfers))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid seed", func() {
			_, err := api.GetTransfers("asdf", GetTransfersOptions{Start: 0})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidSeed))
		})

		It("returns an error for invalid start end option", func() {
			var end uint64 = 9
			_, err := api.GetTransfers("asdf", GetTransfersOptions{Start: 0, End: &end})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidSeed))
		})
	})

})
