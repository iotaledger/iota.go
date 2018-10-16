package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("GetNeighbors()", func() {

	var api *API
	BeforeEach(func() {
		a, err := ComposeAPI(HttpClientSettings{}, nil)
		if err != nil {
			panic(err)
		}
		api = a
	})

	It("resolves to correct response", func() {
		defer gock.Flush()
		gock.New(DefaultLocalIRIURI).
			Post("/").
			MatchType("json").
			JSON(GetNeighborsCommand{Command: GetNeighborsCmd}).
			Reply(200).
			JSON(GetNeighborsResponse{Neighbors: Neighbors{
				{
					Address:                     "tcp://example.com:14600",
					NumberOfAllTransactions:     9001,
					NumberOfInvalidTransactions: 0,
					NumberOfNewTransactions:     100,
				},
			}})

		neighbors, err := api.GetNeighbors()
		Expect(err).ToNot(HaveOccurred())
		Expect(neighbors[0]).To(Equal(Neighbor{
			Address:                     "tcp://example.com:14600",
			NumberOfAllTransactions:     9001,
			NumberOfInvalidTransactions: 0,
			NumberOfNewTransactions:     100,
		}))
	})

})
