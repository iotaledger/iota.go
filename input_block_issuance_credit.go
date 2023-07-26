package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

type BlockIssuanceCreditInput struct {
	AccountID AccountID `serix:"0,mapKey=accountId"`
}

func (b *BlockIssuanceCreditInput) Type() ContextInputType {
	return ContextInputBlockIssuanceCredit
}

func (b *BlockIssuanceCreditInput) Size() int {
	// ContextInputType + AccountID
	return serializer.OneByte + AccountIDLength
}

func (b *BlockIssuanceCreditInput) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	workScoreBytes, err := workScoreStructure.DataByte.Multiply(b.Size())
	if err != nil {
		return 0, err
	}

	// context inputs require invocation of informations in the node, so requires extra work.
	return workScoreBytes.Add(workScoreStructure.ContextInput)
}
