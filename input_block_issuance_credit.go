package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

type BlockIssuanceCreditInput struct {
	AccountID AccountID `serix:""`
}

func (b *BlockIssuanceCreditInput) Clone() ContextInput {
	return &BlockIssuanceCreditInput{
		AccountID: b.AccountID,
	}
}

func (b *BlockIssuanceCreditInput) Type() ContextInputType {
	return ContextInputBlockIssuanceCredit
}

func (b *BlockIssuanceCreditInput) IsReadOnly() bool {
	return true
}

func (b *BlockIssuanceCreditInput) Size() int {
	// ContextInputType + AccountID
	return serializer.OneByte + AccountIDLength
}

func (b *BlockIssuanceCreditInput) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	// context inputs require invocation of informations in the node, so requires extra work.
	return workScoreParameters.ContextInput, nil
}
