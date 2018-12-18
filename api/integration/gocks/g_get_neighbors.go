package gocks

import (
	. "github.com/iotaledger/iota.go/api"
	"gopkg.in/h2non/gock.v1"
)

func init() {
	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(GetNeighborsCommand{Command: Command{GetNeighborsCmd}}).
		Reply(200).
		JSON(GetNeighborsResponse{Neighbors: Neighbors{
			{
				Address:                     "tcp://example.com:14600",
				NumberOfAllTransactions:     9001,
				NumberOfInvalidTransactions: 0,
				NumberOfNewTransactions:     100,
			},
		}})
}
