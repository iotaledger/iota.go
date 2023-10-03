package iotago

import (
	"encoding/binary"

	"github.com/iotaledger/hive.go/serializer/v2"
)

type RewardInput struct {
	// The index of the transaction input for which to claim rewards.
	Index uint16 `serix:"0,mapKey=index"`
}

func (r *RewardInput) Clone() Input {
	return &RewardInput{
		Index: r.Index,
	}
}

func (r *RewardInput) ReferencedStateID() Identifier {
	return r.StateID()
}

func (r *RewardInput) StateID() Identifier {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, r.Index)
	return IdentifierFromData(buf)
}

func (r *RewardInput) Type() StateType {
	return InputReward
}

func (r *RewardInput) IsReadOnly() bool {
	return true
}

func (r *RewardInput) Size() int {
	// ContextInputType + Slot
	return serializer.OneByte + serializer.UInt16ByteSize
}

func (r *RewardInput) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	// context inputs require invocation of informations in the node, so requires extra work.
	return workScoreStructure.ContextInput, nil
}
