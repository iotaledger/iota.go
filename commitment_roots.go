package iotago

import (
	"crypto"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/iota.go/v4/merklehasher"
)

type Roots struct {
	TangleRoot        Identifier `serix:"0"`
	StateMutationRoot Identifier `serix:"1"`
	StateRoot         Identifier `serix:"2"`
	ManaRoot          Identifier `serix:"3"`
	AttestationsRoot  Identifier `serix:"4"`
}

func NewRoots(tangleRoot, stateMutationRoot, attestationsRoot, stateRoot, manaRoot Identifier) *Roots {
	return &Roots{
		TangleRoot:        tangleRoot,
		StateMutationRoot: stateMutationRoot,
		StateRoot:         stateRoot,
		ManaRoot:          manaRoot,
		AttestationsRoot:  attestationsRoot,
	}
}

func (r *Roots) values() []Identifier {
	return []Identifier{
		r.TangleRoot,
		r.StateMutationRoot,
		r.StateRoot,
		r.ManaRoot,
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
	proof := lo.PanicOnErr(merklehasher.NewHasher[Identifier](crypto.BLAKE2b_256).ComputeProofForIndex(r.values(), 4))
	fmt.Printf("proof: %s, attestationRoot: %s, root: %s\n", string(lo.PanicOnErr(json.Marshal(proof))), r.AttestationsRoot, r.ID())
	return proof
}

func VerifyProof(proof *merklehasher.Proof[Identifier], proofedRoot Identifier, treeRoot Identifier) bool {
	// We can ignore the error because Identifier.Bytes() will never return an error
	if !lo.PanicOnErr(proof.ContainsValue(proofedRoot)) {
		return false
	}

	return treeRoot == Identifier(proof.Hash(merklehasher.NewHasher[Identifier](crypto.BLAKE2b_256)))
}
