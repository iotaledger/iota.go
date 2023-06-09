package iotago

import (
	"github.com/iotaledger/iota.go/v4/util"
)

type CommitmentInput struct {
	AccountID    AccountID    `serix:"0,mapKey=accountId"`
	CommitmentID CommitmentID `serix:"1,mapKey=commitmentId"`
}

func (b *CommitmentInput) Type() InputType {
	return InputCommitment
}

func (b *CommitmentInput) Size() int {
	return util.NumByteLen(byte(InputCommitment)) + AccountIDLength + SlotIdentifierLength
}
