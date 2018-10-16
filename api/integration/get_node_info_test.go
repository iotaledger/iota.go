package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("GetNodeInfo()", func() {

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
			JSON(GetNodeInfoCommand{Command: GetNodeInfoCmd}).
			Reply(200).
			JSON(GetNodeInfoResponse{
				LatestMilestoneIndex:               9001,
				LatestSolidSubtangleMilestoneIndex: 9001,
			})

		info, err := api.GetNodeInfo()
		Expect(err).ToNot(HaveOccurred())
		Expect(*info).To(Equal(GetNodeInfoResponse{
			LatestMilestoneIndex:               9001,
			LatestSolidSubtangleMilestoneIndex: 9001,
		}))
	})

})
