package gocks

import (
	. "github.com/iotaledger/iota.go/api"
	"gopkg.in/h2non/gock.v1"
	"strings"
)

func init() {
	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(GetNodeInfoCommand{Command: Command{GetNodeInfoCmd}}).
		Reply(200).
		JSON(GetNodeInfoResponse{
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
		})
}
