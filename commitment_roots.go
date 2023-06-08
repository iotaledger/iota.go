package iotago

import (
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
)

type Roots struct {
	TangleRoot        Identifier `serix:"0"`
	StateMutationRoot Identifier `serix:"1"`
	StateRoot         Identifier `serix:"2"`
	ManaRoot          Identifier `serix:"3"`
	AttestationsRoot  Identifier `serix:"4"`
}

func NewRoots(tangleRoot, stateMutationRoot, activityRoot, stateRoot, manaRoot Identifier) *Roots {
	return &Roots{
		TangleRoot:        tangleRoot,
		StateMutationRoot: stateMutationRoot,
		AttestationsRoot:  activityRoot,
		StateRoot:         stateRoot,
		ManaRoot:          manaRoot,
	}
}

func (r *Roots) ID() (id Identifier) {
	h01 := blake2b.Sum256(byteutils.ConcatBytes(r.TangleRoot[:], r.StateMutationRoot[:]))
	h23 := blake2b.Sum256(byteutils.ConcatBytes(r.StateRoot[:], r.ManaRoot[:]))
	h0123 := blake2b.Sum256(byteutils.ConcatBytes(h01[:], h23[:]))

	h45 := blake2b.Sum256(byteutils.ConcatBytes(r.AttestationsRoot[:], emptyIdentifier[:]))
	h67 := blake2b.Sum256(byteutils.ConcatBytes(emptyIdentifier[:], emptyIdentifier[:]))
	h4567 := blake2b.Sum256(byteutils.ConcatBytes(h45[:], h67[:]))

	rootHashed := blake2b.Sum256(byteutils.ConcatBytes(h0123[:], h4567[:]))

	return rootHashed
}

func (r *Roots) AttestationsProof() Identifier {
	h01 := blake2b.Sum256(byteutils.ConcatBytes(r.TangleRoot[:], r.StateMutationRoot[:]))
	h23 := blake2b.Sum256(byteutils.ConcatBytes(r.StateRoot[:], r.ManaRoot[:]))
	h0123 := blake2b.Sum256(byteutils.ConcatBytes(h01[:], h23[:]))

	return h0123
}

func VerifyRootsAttestationsProof(rootsID, h0123, attestationsRoot Identifier) bool {
	h45 := blake2b.Sum256(byteutils.ConcatBytes(attestationsRoot[:], emptyIdentifier[:]))
	h67 := blake2b.Sum256(byteutils.ConcatBytes(emptyIdentifier[:], emptyIdentifier[:]))
	h4567 := blake2b.Sum256(byteutils.ConcatBytes(h45[:], h67[:]))

	computedRootsID := blake2b.Sum256(byteutils.ConcatBytes(h0123[:], h4567[:]))

	return rootsID == computedRootsID
}
