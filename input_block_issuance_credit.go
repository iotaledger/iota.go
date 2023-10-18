package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

type BlockIssuanceCreditInput struct {
	AccountID AccountID `serix:"0,mapKey=accountId"`
}

func (b *BlockIssuanceCreditInput) Clone() Input {
	return &BlockIssuanceCreditInput{
		AccountID: b.AccountID,
	}
}

func (b *BlockIssuanceCreditInput) ReferencedStateID() Identifier {
	return b.StateID()
}

func (b *BlockIssuanceCreditInput) StateID() Identifier {
	return IdentifierFromData(b.AccountID[:])
}

func (b *BlockIssuanceCreditInput) Type() StateType {
	return InputBlockIssuanceCredit
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
