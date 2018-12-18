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
		JSON(InterruptAttachToTangleCommand{Command: Command{InterruptAttachToTangleCmd}}).
		Reply(200)
}
