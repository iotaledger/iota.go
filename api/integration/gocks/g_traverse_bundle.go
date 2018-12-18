package gocks

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	. "github.com/iotaledger/iota.go/trinary"
	"gopkg.in/h2non/gock.v1"
)

func init() {
	for i := range Bundle {
		// create a getTrytes mock for each tx in sample bundle
		gock.New(DefaultLocalIRIURI).
			Persist().
			Post("/").
			MatchType("json").
			JSON(GetTrytesCommand{
				Command: Command{GetTrytesCmd},
				Hashes:  Hashes{Bundle[i].Hash},
			}).
			Reply(200).
			JSON(GetTrytesResponse{Trytes: []Trytes{BundleTrytes[i]}})
	}

	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(GetTrytesCommand{
			Command: Command{GetTrytesCmd},
			Hashes:  Hashes{BundleWithZeroValue[0].Hash},
		}).
		Reply(200).
		JSON(GetTrytesResponse{Trytes: []Trytes{BundleWithZeroValueTrytes[0]}})
}
