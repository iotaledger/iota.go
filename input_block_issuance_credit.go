package iotago

import (
	"bytes"
	"cmp"

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

func (b *BlockIssuanceCreditInput) Compare(other ContextInput) int {
	typeCompare := cmp.Compare(b.Type(), other.Type())
	if typeCompare != 0 {
		return typeCompare
	}

	// Causes any two BIC Inputs with the same account ID to be considered duplicates.
	//nolint:forcetypeassert // we can safely assume that this is a BlockIssuanceCreditInput
	otherBICInput := other.(*BlockIssuanceCreditInput)

	return bytes.Compare(b.AccountID[:], otherBICInput.AccountID[:])
}
