package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/consts"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("RemoveNeighbors()", func() {

	var api *API
	BeforeEach(func() {
		a, err := ComposeAPI(HttpClientSettings{}, nil)
		if err != nil {
			panic(err)
		}
		api = a
	})

	Context("call", func() {

		It("resolves to correct response", func() {
			neighborToRemove := "tcp://example.com:14600"
			defer gock.Flush()
			gock.New(DefaultLocalIRIURI).
				Post("/").
				MatchType("json").
				JSON(RemoveNeighborsCommand{
					Command: RemoveNeighborsCmd,
					URIs:    []string{neighborToRemove},
				}).
				Reply(200).
				JSON(RemoveNeighborsResponse{
					Duration: 10, RemovedNeighbors: 1,
				})

			removed, err := api.RemoveNeighbors(neighborToRemove)
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
