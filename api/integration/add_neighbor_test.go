package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	_ "github.com/iotaledger/iota.go/api/integration/gocks"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("AddNeighbors()", func() {

	var api *API

	BeforeEach(func() {
		a, err := ComposeAPI(HTTPClientSettings{}, nil)
		if err != nil {
			panic(err)
		}
		api = a
	})

	It("resolves to the correct response", func() {
		added, err := api.AddNeighbors("tcp://example.com:14600")
		Expect(err).ToNot(HaveOccurred())
		Expect(added).To(Equal(int64(1)))
	})

	It("returns an error for invalid uris", func() {
		_, err := api.AddNeighbors("example.com")
		Expect(errors.Cause(err)).To(Equal(ErrInvalidURI))
	})

	It("returns an error for empty uris", func() {
		_, err := api.AddNeighbors()
		Expect(errors.Cause(err)).To(Equal(ErrInvalidURI))
	})

})
