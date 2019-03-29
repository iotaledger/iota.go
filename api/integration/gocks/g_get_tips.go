package gocks

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/trinary"
	"gopkg.in/h2non/gock.v1"
	"strings"
)

func init() {
	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(GetTipsCommand{Command: Command{GetTipsCmd}}).
		Reply(200).
		JSON(GetTipsResponse{Hashes: Hashes{strings.Repeat("T", 81), strings.Repeat("U", 81)}})
}
