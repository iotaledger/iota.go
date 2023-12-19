package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

type RewardInput struct {
	// The index of the transaction input for which to claim rewards.
	Index uint16 `serix:""`
}

func (r *RewardInput) Clone() ContextInput {
	return &RewardInput{
		Index: r.Index,
	}
}

func (r *RewardInput) Type() ContextInputType {
	return ContextInputReward
}

func (r *RewardInput) IsReadOnly() bool {
	return true
}

func (r *RewardInput) Size() int {
	// ContextInputType + Slot
	return serializer.OneByte + serializer.UInt16ByteSize
}

func (r *RewardInput) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	// context inputs require invocation of informations in the node, so requires extra work.
	return workScoreParameters.ContextInput, nil
}
