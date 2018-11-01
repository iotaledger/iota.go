package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("GetBundle()", func() {
	api, err := ComposeAPI(HTTPClientSettings{})
	if err != nil {
		panic(err)
	}

	Context("call", func() {
		It("resolves to correct response", func() {
			bndl, err := api.GetBundle(Bundle[0].Hash)
			Expect(err).ToNot(HaveOccurred())
			Expect(bndl).To(Equal(Bundle))
		})

		It("resolves to correct response for single transaction bundle", func() {
			bndl, err := api.GetBundle(BundleWithZeroValue[0].Hash)
			Expect(err).ToNot(HaveOccurred())
			Expect(bndl).To(Equal(BundleWithZeroValue))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid transaction hash", func() {
			_, err := api.GetBundle("asdf")
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTransactionHash))
		})
	})

})
