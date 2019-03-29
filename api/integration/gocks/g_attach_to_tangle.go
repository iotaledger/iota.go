package gocks

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/transaction"
	. "github.com/iotaledger/iota.go/trinary"
	"gopkg.in/h2non/gock.v1"
	"time"
)

func init() {
	reqTrytes := make([]Trytes, len(BundleTrytes))
	copy(reqTrytes, BundleTrytes)
	for i, j := 0, len(reqTrytes)-1; i < j; i, j = i+1, j-1 {
		reqTrytes[i], reqTrytes[j] = reqTrytes[j], reqTrytes[i]
	}

	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(AttachToTangleCommand{
			Command:            Command{AttachToTangleCmd},
			TrunkTransaction:   TrunkTx,
			BranchTransaction:  BranchTx,
			Trytes:             reqTrytes,
			MinWeightMagnitude: 14,
		}).
		Reply(200).
		JSON(AttachToTangleResponse{Trytes: BundleTrytes})

	ts := uint64(time.Now().UnixNano() / int64(time.Second))
	entries, err := bundle.TransfersToBundleEntries(ts, bundle.EmptyTransfer)
	if err != nil {
		panic(err)
	}

	emptyTxTrytes := transaction.MustTransactionToTrytes(&(bundle.AddEntry(bundle.Bundle{}, entries[0])[0]))

	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(AttachToTangleCommand{
			Command:            Command{AttachToTangleCmd},
			TrunkTransaction:   TrunkTx,
			BranchTransaction:  BranchTx,
			Trytes:             []Trytes{emptyTxTrytes},
			MinWeightMagnitude: 14,
		}).
		Reply(200).
		JSON(AttachToTangleResponse{Trytes: BundleTrytes})
}
