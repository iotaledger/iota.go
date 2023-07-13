package iotago

import (
	"github.com/iotaledger/iota.go/v4/util"
)

type BlockIssuanceCreditInput struct {
	AccountID AccountID `serix:"0,mapKey=accountId"`
}

func (b *BlockIssuanceCreditInput) Type() ContextInputType {
	return ContextInputBlockIssuanceCredit
}

func (b *BlockIssuanceCreditInput) Size() int {
	return util.NumByteLen(byte(ContextInputBlockIssuanceCredit)) + AccountIDLength
}
