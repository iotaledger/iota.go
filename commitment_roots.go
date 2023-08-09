package iotago

import (
	"crypto"
	"fmt"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/iota.go/v4/merklehasher"
)

type Roots struct {
	TangleRoot             Identifier `serix:"0"`
	StateMutationRoot      Identifier `serix:"1"`
	StateRoot              Identifier `serix:"2"`
	AccountRoot            Identifier `serix:"4"`
	AttestationsRoot       Identifier `serix:"5"`
	CommitteeRoot          Identifier `serix:"6"`
	RewardsRoot            Identifier `serix:"7"`
	ProtocolParametersHash Identifier `serix:"8"`
}

func NewRoots(tangleRoot, stateMutationRoot, attestationsRoot, stateRoot, accountRoot, committeeRoot, rewardsRoot, protocolParametersHash Identifier) *Roots {
	return &Roots{
		TangleRoot:             tangleRoot,
		StateMutationRoot:      stateMutationRoot,
		StateRoot:              stateRoot,
		AccountRoot:            accountRoot,
		AttestationsRoot:       attestationsRoot,
		CommitteeRoot:          committeeRoot,
		RewardsRoot:            rewardsRoot,
		ProtocolParametersHash: protocolParametersHash,
	}
}

func (r *Roots) values() []Identifier {
	return []Identifier{
		r.TangleRoot,
		r.StateMutationRoot,
		r.StateRoot,
		r.AccountRoot,
		r.AttestationsRoot,
		r.CommitteeRoot,
		r.RewardsRoot,
		r.ProtocolParametersHash,
	}
}

func (r *Roots) ID() (id Identifier) {
	// We can ignore the error because Identifier.Bytes() will never return an error
	return Identifier(
		lo.PanicOnErr(
			//nolint:nosnakecase // false positive
			merklehasher.NewHasher[Identifier](crypto.BLAKE2b_256).HashValues(r.values()),
		),
	)
}

func (r *Roots) AttestationsProof() *merklehasher.Proof[Identifier] {
	// We can ignore the error because Identifier.Bytes() will never return an error
	//nolint:nosnakecase // false positive
	return lo.PanicOnErr(merklehasher.NewHasher[Identifier](crypto.BLAKE2b_256).ComputeProofForIndex(r.values(), 4))
}

func VerifyProof(proof *merklehasher.Proof[Identifier], proofedRoot Identifier, treeRoot Identifier) bool {
	// We can ignore the error because Identifier.Bytes() will never return an error
	if !lo.PanicOnErr(proof.ContainsValue(proofedRoot)) {
		return false
	}

	//nolint:nosnakecase // false positive
	return treeRoot == Identifier(proof.Hash(merklehasher.NewHasher[Identifier](crypto.BLAKE2b_256)))
}

func (r *Roots) String() string {
	return fmt.Sprintf(
		"Roots(%s): TangleRoot: %s, StateMutationRoot: %s, StateRoot: %s, AccountRoot: %s, AttestationsRoot: %s, CommitteeRoot: %s, RewardsRoot: %s, ProtocolParametersHash: %s", r.ID(), r.TangleRoot, r.StateMutationRoot, r.StateRoot, r.AccountRoot, r.AttestationsRoot, r.CommitteeRoot, r.RewardsRoot, r.ProtocolParametersHash)
}
