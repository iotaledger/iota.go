package iotago

import (
	"crypto"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/iota.go/v4/merklehasher"
)

type Roots struct {
	TangleRoot        Identifier `serix:"0"`
	StateMutationRoot Identifier `serix:"1"`
	StateRoot         Identifier `serix:"2"`
	AccountRoot       Identifier `serix:"4"`
	AttestationsRoot  Identifier `serix:"5"`
}

func NewRoots(tangleRoot, stateMutationRoot, attestationsRoot, stateRoot, accountRoot Identifier) *Roots {
	return &Roots{
		TangleRoot:        tangleRoot,
		StateMutationRoot: stateMutationRoot,
		StateRoot:         stateRoot,
		AccountRoot:       accountRoot,
		AttestationsRoot:  attestationsRoot,
	}
}

func (r *Roots) values() []Identifier {
	return []Identifier{
		r.TangleRoot,
		r.StateMutationRoot,
		r.StateRoot,
		r.AccountRoot,
		r.AttestationsRoot,
	}
}

func (r *Roots) ID() (id Identifier) {
	// We can ignore the error because Identifier.Bytes() will never return an error
	return Identifier(
		lo.PanicOnErr(
			merklehasher.NewHasher[Identifier](crypto.BLAKE2b_256).HashValues(r.values()),
		),
	)
}

func (r *Roots) AttestationsProof() *merklehasher.Proof[Identifier] {
	// We can ignore the error because Identifier.Bytes() will never return an error
	return lo.PanicOnErr(merklehasher.NewHasher[Identifier](crypto.BLAKE2b_256).ComputeProofForIndex(r.values(), 4))

}

func VerifyProof(proof *merklehasher.Proof[Identifier], proofedRoot Identifier, treeRoot Identifier) bool {
	// We can ignore the error because Identifier.Bytes() will never return an error
	if !lo.PanicOnErr(proof.ContainsValue(proofedRoot)) {
		return false
	}

	return treeRoot == Identifier(proof.Hash(merklehasher.NewHasher[Identifier](crypto.BLAKE2b_256)))
}
