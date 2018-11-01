package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("ReplayBundle()", func() {
	api, err := ComposeAPI(HTTPClientSettings{})
	if err != nil {
		panic(err)
	}

	Context("call", func() {
		It("resolves to correct response", func() {
			bndl, err := api.ReplayBundle(Bundle[0].Hash, 3, 14)
			Expect(err).ToNot(HaveOccurred())
			Expect(bndl).To(Equal(Bundle))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid transaction hash", func() {
			_, err := api.ReplayBundle("asdf", 3, 14)
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTransactionHash))
		})
	})

})
