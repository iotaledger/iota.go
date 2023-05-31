package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/util"
)

type BICInput struct {
	AccountID
	CommitmentID
}

func (b *BICInput) Type() InputType {
	return InputBlockIssuanceCredit
}

func (b *BICInput) Size() int {
	return util.NumByteLen(byte(InputBlockIssuanceCredit)) + AccountIDLength + util.NumByteLen(SlotIndex(0)) + serializer.Int64ByteSize
}
