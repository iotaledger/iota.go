package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/gocks"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("RemoveNeighbors()", func() {

	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	Context("call", func() {
		It("resolves to correct response", func() {
			removed, err := api.RemoveNeighbors(NeighborURI)
			Expect(err).ToNot(HaveOccurred())
			Expect(removed).To(Equal(int64(1)))
		})
	})

	Context("invalid input", func() {
		It("returns an error for invalid uris", func() {
			_, err := api.RemoveNeighbors("")
			Expect(errors.Cause(err)).To(Equal(ErrInvalidURI))
		})
	})

})
