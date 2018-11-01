package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	"github.com/iotaledger/iota.go/bundle"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("TraverseBundle()", func() {
	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	Context("call", func() {
		It("resolves to correct response", func() {
			tail := Bundle[0].Hash
			bndl, err := api.TraverseBundle(tail, bundle.Bundle{})
			Expect(err).ToNot(HaveOccurred())
			Expect(bndl).To(Equal(Bundle))
		})

		It("resolves to correct single transaction bundle", func() {
			tail := BundleWithZeroValue[0].Hash
			bndl, err := api.TraverseBundle(tail, bundle.Bundle{})
			Expect(err).ToNot(HaveOccurred())
			Expect(bndl).To(Equal(BundleWithZeroValue))
		})

	})

	Context("invalid input", func() {
		It("returns an error for invalid hash", func() {
			_, err := api.TraverseBundle("asdf", bundle.Bundle{})
			Expect(errors.Cause(err)).To(Equal(ErrInvalidTransactionHash))
		})
	})

})
