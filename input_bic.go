package iotago

import (
	"github.com/iotaledger/iota.go/v4/util"
)

type BICInput struct {
	AccountID    AccountID    `serix:"0,mapKey=accountId"`
	CommitmentID CommitmentID `serix:"1,mapKey=commitmentId"`
}

func (b *BICInput) Type() InputType {
	return InputBlockIssuanceCredit
}

func (b *BICInput) Size() int {
	return util.NumByteLen(byte(InputBlockIssuanceCredit)) + AccountIDLength + SlotIdentifierLength
}
