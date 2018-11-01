package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	_ "github.com/iotaledger/iota.go/api/integration/gocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetNeighbors()", func() {

	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	It("resolves to correct response", func() {
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
