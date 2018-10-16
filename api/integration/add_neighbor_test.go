package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("AddNeighbors()", func() {

	var api *API

	BeforeEach(func() {
		a, err := ComposeAPI(HttpClientSettings{}, nil)
		if err != nil {
			panic(err)
		}
		api = a
	})

	BeforeEach(func() {
		gock.New(DefaultLocalIRIURI).
			Post("/").
			MatchType("json").
			JSON(AddNeighborsCommand{Command: AddNeighborsCmd, URIs: []string{"tcp://example.com:14600"}}).
			Reply(200).
			JSON(AddNeighborsResponse{AddedNeighbors: 1, Duration: 7})
	})

	It("resolves to the correct response", func() {
		defer gock.Flush()

		added, err := api.AddNeighbors("tcp://example.com:14600")
		Expect(err).ToNot(HaveOccurred())
		Expect(added).To(Equal(int64(1)))
	})

	It("returns an error for invalid uris", func() {
		defer gock.Flush()
		_, err := api.AddNeighbors("example.com")
		Expect(err).To(HaveOccurred())
	})

	It("returns an error for empty uris", func() {
		defer gock.Flush()
		_, err := api.AddNeighbors()
		Expect(err).To(HaveOccurred())
	})

})
