package api_test

import (
	"fmt"
	. "github.com/iotaledger/iota.go/api"
	"github.com/iotaledger/iota.go/trinary"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gopkg.in/h2non/gock.v1"
	"strings"
)

type fakereqres struct {
	Command
	Val int `json:"val"`
}

var _ = Describe("Quorum", func() {

	const nodesCount = 4
	nodes := make([]string, nodesCount)
	for i := 0; i < nodesCount; i++ {
		nodes[i] = fmt.Sprintf("http:/%d", i)
	}

	Context("Voting", func() {
		It("throws an error when quorum couldn't be reached 0%", func() {
			provider, _ := NewQuorumHTTPClient(QuorumHTTPClientSettings{
				Nodes:     nodes,
				Threshold: 1,
			})
			defer gock.Flush()
			// every node gives a different answer
			for i, node := range nodes {
				gock.New(node).
					Post("/").
					MatchType("json").
					JSON(fakereqres{Val: 0}).
					Reply(200).
					JSON(fakereqres{Val: i})
			}
			err := provider.Send(&fakereqres{Val: 0}, nil)
			Expect(errors.Cause(err)).To(Equal(ErrQuorumNotReached))
		})

		It("throws an error when quorum couldn't be reached 50%", func() {
			provider, _ := NewQuorumHTTPClient(QuorumHTTPClientSettings{
				Nodes:     nodes,
				Threshold: 1,
			})
			defer gock.Flush()
			// 50%
			for i, node := range nodes {
				var res int
				if i%2 == 0 {
					res = 1
				} else {
					res = 0
				}
				gock.New(node).
					Post("/").
					MatchType("json").
					JSON(fakereqres{Val: 0}).
					Reply(200).
					JSON(fakereqres{Val: res})
			}

			err := provider.Send(&fakereqres{Val: 0}, nil)
			Expect(errors.Cause(err)).To(Equal(ErrQuorumNotReached))
		})

		It("returns the optional defined value when quorum couldn't be reached", func() {
			defVal := true
			provider, _ := NewQuorumHTTPClient(QuorumHTTPClientSettings{
				Nodes:     nodes[:2],
				Threshold: 1,
				Defaults: &QuorumDefaults{
					WereAddressesSpentFrom: &defVal,
				},
			})

			req := &WereAddressesSpentFromCommand{
				Addresses: trinary.Hashes{"bla", "alb", "lab"},
				Command:   Command{WereAddressesSpentFromCmd},
			}
			defer gock.Flush()
			// 50%, 100% not reached
			for i := 0; i < 2; i++ {
				var answer bool
				if i == 1 {
					answer = true
				}
				gock.New(nodes[i]).
					Post("/").
					MatchType("json").
					JSON(req).
					Reply(200).
					JSON(WereAddressesSpentFromResponse{States: []bool{answer}})
			}
			res := &WereAddressesSpentFromResponse{}
			err := provider.Send(req, res)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.States[0]).To(Equal(true))
		})

		It("returns the response when a quorum was reached", func() {
			provider, _ := NewQuorumHTTPClient(QuorumHTTPClientSettings{
				Nodes:     nodes,
				Threshold: 1,
			})
			const resVal = 1
			defer gock.Flush()
			for _, node := range nodes {
				gock.New(node).
					Post("/").
					MatchType("json").
					JSON(fakereqres{Val: 0}).
					Reply(200).
					JSON(fakereqres{Val: resVal})
			}
			res := &fakereqres{}
			err := provider.Send(&fakereqres{Val: 0}, res)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Val).To(Equal(resVal))
		})

		It("returns the response when a quorum was reached (at threshold 75%)", func() {
			provider, _ := NewQuorumHTTPClient(QuorumHTTPClientSettings{
				Nodes:     nodes,
				Threshold: 0.75,
			})
			defer gock.Flush()
			// one node gives another answer
			for i, node := range nodes {
				var resVal int
				if i == len(nodes)-1 {
					resVal = 1
				}
				gock.New(node).
					Post("/").
					MatchType("json").
					JSON(fakereqres{Val: 0}).
					Reply(200).
					JSON(fakereqres{Val: resVal})
			}
			res := &fakereqres{}
			err := provider.Send(&fakereqres{Val: 0}, res)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.Val).To(Equal(0))
		})

		It("returns the error response when the quorum forms it", func() {
			provider, _ := NewQuorumHTTPClient(QuorumHTTPClientSettings{
				Nodes:     nodes,
				Threshold: 1,
			})
			type errorresp struct {
				Error string `json:"error"`
			}
			const errorMsg = "Command [getBanana] is unknown"
			const resVal = 1
			defer gock.Flush()
			for _, node := range nodes {
				gock.New(node).
					Post("/").
					MatchType("json").
					JSON(fakereqres{Val: 0}).
					Reply(400).
					JSON(errorresp{errorMsg})
			}
			res := &fakereqres{}
			err := provider.Send(&fakereqres{Val: 0}, res)
			Expect(err).To(HaveOccurred())
			reqErr, ok := err.(*ErrRequestError)
			Expect(ok).To(BeTrue())
			Expect(reqErr.ErrorMessage).To(Equal(errorMsg))
		})
	})

	Context("LatestSolidSubtangleMilestone", func() {
		nodeInfoRes := GetNodeInfoResponse{
			AppName:                            "IRI",
			AppVersion:                         "",
			Duration:                           100,
			JREAvailableProcessors:             4,
			JREFreeMemory:                      13020403,
			JREMaxMemory:                       1241331231,
			JRETotalMemory:                     4245234332,
			LatestMilestone:                    strings.Repeat("X", 81),
			LatestMilestoneIndex:               3,
			LatestSolidSubtangleMilestone:      strings.Repeat("M", 81),
			LatestSolidSubtangleMilestoneIndex: 1,
			Neighbors:                          5,
			PacketsQueueSize:                   23,
			Time:                               213213214,
			Tips:                               123,
			TransactionsToRequest:              10,
		}

		It("returns the correct solid subtangle milestone", func() {
			provider, _ := NewQuorumHTTPClient(QuorumHTTPClientSettings{
				Nodes:                      nodes,
				Threshold:                  0.75,
				MaxSubtangleMilestoneDelta: 1,
			})
			defer gock.Flush()
			req := &GetLatestSolidSubtangleMilestoneCommand{Command: Command{GetNodeInfoCmd}}
			// we create a delta of 1
			for i, node := range nodes {
				resCopy := nodeInfoRes
				if i%2 == 0 {
					resCopy.LatestSolidSubtangleMilestoneIndex = 2
					resCopy.LatestSolidSubtangleMilestone = strings.Repeat("N", 81)
				}
				gock.New(node).
					Post("/").
					MatchType("json").
					JSON(req).
					Reply(200).
					JSON(resCopy)
			}
			res := &GetLatestSolidSubtangleMilestoneResponse{}
			err := provider.Send(req, res)
			Expect(err).ToNot(HaveOccurred())
			Expect(res.LatestSolidSubtangleMilestoneIndex).To(Equal(int64(1)))
			Expect(res.LatestSolidSubtangleMilestone).To(Equal(nodeInfoRes.LatestSolidSubtangleMilestone))
		})

		It("returns an error when MaxSubtangleMilestoneDelta is exceeded", func() {
			provider, _ := NewQuorumHTTPClient(QuorumHTTPClientSettings{
				Nodes:                      nodes,
				Threshold:                  0.75,
				MaxSubtangleMilestoneDelta: 2,
			})
			defer gock.Flush()
			req := &GetLatestSolidSubtangleMilestoneCommand{Command: Command{GetNodeInfoCmd}}
			// delta of 3
			for i, node := range nodes {
				resCopy := nodeInfoRes
				if i%2 == 0 {
					resCopy.LatestSolidSubtangleMilestoneIndex = 4
					resCopy.LatestSolidSubtangleMilestone = strings.Repeat("N", 81)
				}
				gock.New(node).
					Post("/").
					MatchType("json").
					JSON(req).
					Reply(200).
					JSON(resCopy)
			}
			res := &GetLatestSolidSubtangleMilestoneResponse{}
			err := provider.Send(req, res)
			Expect(errors.Cause(err)).To(Equal(ErrExceededMaxSubtangleMilestoneDelta))
		})

		It("returns an error when faulty status codes are over no response tolerance", func() {
			provider, _ := NewQuorumHTTPClient(QuorumHTTPClientSettings{
				Nodes:                      nodes,
				Threshold:                  0.75,
				MaxSubtangleMilestoneDelta: 2,
			})
			defer gock.Flush()
			req := &GetLatestSolidSubtangleMilestoneCommand{Command: Command{GetNodeInfoCmd}}
			// delta of 3
			for _, node := range nodes {
				gock.New(node).
					Post("/").
					MatchType("json").
					JSON(req).
					Reply(404).
					JSON(nodeInfoRes)
			}
			res := &GetLatestSolidSubtangleMilestoneResponse{}
			err := provider.Send(req, res)
			Expect(errors.Cause(err)).To(Equal(ErrExceededNoResponseTolerance))
		})
	})

})
