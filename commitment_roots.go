package iotago

import (
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
)

type Roots struct {
	TangleRoot        Identifier `serix:"0"`
	StateMutationRoot Identifier `serix:"1"`
	ActivityRoot      Identifier `serix:"2"`
	StateRoot         Identifier `serix:"3"`
	ManaRoot          Identifier `serix:"4"`
}

func NewRoots(tangleRoot, stateMutationRoot, activityRoot, stateRoot, manaRoot Identifier) *Roots {
	return &Roots{
		TangleRoot:        tangleRoot,
		StateMutationRoot: stateMutationRoot,
		ActivityRoot:      activityRoot,
		StateRoot:         stateRoot,
		ManaRoot:          manaRoot,
	}
}

func (r *Roots) ID() (id Identifier) {
	branch1Hashed := blake2b.Sum256(byteutils.ConcatBytes(r.TangleRoot[:], r.StateMutationRoot[:]))
	branch2Hashed := blake2b.Sum256(byteutils.ConcatBytes(r.StateRoot[:], r.ManaRoot[:]))
	rootHashed := blake2b.Sum256(byteutils.ConcatBytes(branch1Hashed[:], branch2Hashed[:]))

	//TODO: hash ActivityRoot?
	return rootHashed
}
