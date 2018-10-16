package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("InterruptAttachToTangle()", func() {

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
			defer gock.Flush()
			gock.New(DefaultLocalIRIURI).
				Post("/").
				MatchType("json").
				JSON(InterruptAttachToTangleCommand{Command: InterruptAttachToTangleCmd}).
				Reply(200)

			err := api.InterruptAttachToTangle()
			Expect(err).ToNot(HaveOccurred())
		})
	})

})
