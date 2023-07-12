package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/util"
)

type RewardInput struct {
	// The index of the transaction input for which to claim rewards.
	Index uint16 `serix:"0,mapKey=index"`
}

func (r *RewardInput) Type() ContextInputType {
	return ContextInputReward
}

func (r *RewardInput) Size() int {
	return util.NumByteLen(byte(ContextInputReward)) + serializer.UInt16ByteSize
}
