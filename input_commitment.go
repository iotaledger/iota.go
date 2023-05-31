package iotago

import (
	"github.com/iotaledger/iota.go/v4/util"
)

type CommitmentInput struct {
	AccountID
	CommitmentID
}

func (b *CommitmentInput) Type() InputType {
	return InputCommitment
}

func (b *CommitmentInput) Size() int {
	return util.NumByteLen(byte(InputCommitment)) + SlotIdentifierLength
}
