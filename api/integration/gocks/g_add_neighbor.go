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
		JSON(AddNeighborsCommand{Command: Command{AddNeighborsCmd}, URIs: []string{"tcp://example.com:14600"}}).
		Reply(200).
		JSON(AddNeighborsResponse{AddedNeighbors: 1, Duration: 7})
}
