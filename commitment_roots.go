package iotago

import (
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
)

type Roots struct {
	TangleRoot        [32]byte `serix:"0"`
	StateMutationRoot [32]byte `serix:"1"`
	ActivityRoot      [32]byte `serix:"4"`
	StateRoot         [32]byte `serix:"2"`
	ManaRoot          [32]byte `serix:"3"`
}

func NewRoots(tangleRoot, stateMutationRoot, activityRoot, stateRoot, manaRoot [32]byte) *Roots {
	return &Roots{
		TangleRoot:        tangleRoot,
		StateMutationRoot: stateMutationRoot,
		ActivityRoot:      activityRoot,
		StateRoot:         stateRoot,
		ManaRoot:          manaRoot,
	}
}

func (r *Roots) ID() (id types.Identifier) {
	branch1Hashed := blake2b.Sum256(byteutils.ConcatBytes(r.TangleRoot[:], r.StateMutationRoot[:]))
	branch2Hashed := blake2b.Sum256(byteutils.ConcatBytes(r.StateRoot[:], r.ManaRoot[:]))
	rootHashed := blake2b.Sum256(byteutils.ConcatBytes(branch1Hashed[:], branch2Hashed[:]))

	return rootHashed
}
