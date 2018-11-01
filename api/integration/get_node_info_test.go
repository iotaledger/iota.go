package integration_test

import (
	. "github.com/iotaledger/iota.go/api"
	_ "github.com/iotaledger/iota.go/api/integration/gocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strings"
)

var _ = Describe("GetNodeInfo()", func() {
	api, err := ComposeAPI(HTTPClientSettings{}, nil)
	if err != nil {
		panic(err)
	}

	It("resolves to correct response", func() {
		info, err := api.GetNodeInfo()
		Expect(err).ToNot(HaveOccurred())
		Expect(*info).To(Equal(GetNodeInfoResponse{
			AppName:                            "IRI",
			AppVersion:                         "",
			Duration:                           100,
			JREAvailableProcessors:             4,
			JREFreeMemory:                      13020403,
			JREMaxMemory:                       1241331231,
			JRETotalMemory:                     4245234332,
			LatestMilestone:                    strings.Repeat("M", 81),
			LatestMilestoneIndex:               1,
			LatestSolidSubtangleMilestone:      strings.Repeat("M", 81),
			LatestSolidSubtangleMilestoneIndex: 1,
			Neighbors:                          5,
			PacketsQueueSize:                   23,
			Time:                               213213214,
			Tips:                               123,
			TransactionsToRequest:              10,
		}))
	})

})
