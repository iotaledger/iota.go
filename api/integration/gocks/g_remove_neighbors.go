package gocks

import (
	. "github.com/iotaledger/iota.go/api"
	"gopkg.in/h2non/gock.v1"
)

var NeighborURI = "tcp://example.com:14600"

func init() {
	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(RemoveNeighborsCommand{
			Command: Command{RemoveNeighborsCmd},
			URIs:    []string{NeighborURI},
		}).
		Reply(200).
		JSON(RemoveNeighborsResponse{
			Duration: 10, RemovedNeighbors: 1,
		})
}
