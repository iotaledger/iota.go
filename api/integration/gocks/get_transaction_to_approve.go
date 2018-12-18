package gocks

import (
	. "github.com/iotaledger/iota.go/api"
	. "github.com/iotaledger/iota.go/api/integration/samples"
	"gopkg.in/h2non/gock.v1"
	"strings"
)

var TrunkTx = Bundle[len(Bundle)-1].TrunkTransaction
var BranchTx = Bundle[len(Bundle)-1].BranchTransaction

func init() {
	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(GetTransactionsToApproveCommand{
			Command: Command{GetTransactionsToApproveCmd},
			Depth:   3,
		}).
		Reply(200).
		JSON(GetTransactionsToApproveResponse{
			Duration: 100, TransactionsToApprove: TransactionsToApprove{
				TrunkTransaction:  TrunkTx,
				BranchTransaction: BranchTx,
			},
		})

	gock.New(DefaultLocalIRIURI).
		Persist().
		Post("/").
		MatchType("json").
		JSON(GetTransactionsToApproveCommand{
			Command: Command{GetTransactionsToApproveCmd},
			Depth:   3, Reference: strings.Repeat("R", 81),
		}).
		Reply(200).
		JSON(GetTransactionsToApproveResponse{
			Duration: 100,
			TransactionsToApprove: TransactionsToApprove{
				TrunkTransaction:  BranchTx, // don't get confused
				BranchTransaction: TrunkTx,
			},
		})

}
