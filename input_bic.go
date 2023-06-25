package iotago

import (
	"github.com/iotaledger/iota.go/v4/util"
)

type BICInput struct {
	AccountID AccountID `serix:"0,mapKey=accountId"`
}

func (b *BICInput) Type() InputType {
	return InputBlockIssuanceCredit
}

func (b *BICInput) Size() int {
	return util.NumByteLen(byte(InputBlockIssuanceCredit)) + AccountIDLength
}
