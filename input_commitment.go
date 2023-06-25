package iotago

import (
	"github.com/iotaledger/iota.go/v4/util"
)

type CommitmentInput struct {
	CommitmentID CommitmentID `serix:"0,mapKey=commitmentId"`
}

func (c *CommitmentInput) Type() InputType {
	return InputCommitment
}

func (c *CommitmentInput) Size() int {
	return util.NumByteLen(byte(InputCommitment)) + SlotIdentifierLength
}
