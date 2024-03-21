package iotago

import (
	"crypto"
	"fmt"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/iota.go/v4/merklehasher"
)

type Roots struct {
	TangleRoot             Identifier `serix:""`
	StateMutationRoot      Identifier `serix:""`
	StateRoot              Identifier `serix:""`
	AccountRoot            Identifier `serix:""`
	AttestationsRoot       Identifier `serix:""`
	CommitteeRoot          Identifier `serix:""`
	RewardsRoot            Identifier `serix:""`
	ProtocolParametersHash Identifier `serix:""`
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
			merklehasher.NewHasher[Identifier](crypto.BLAKE2b_256).HashValues(r.values()),
		),
	)
}

func (r *Roots) AttestationsProof() *merklehasher.Proof[Identifier] {
	// We can ignore the error because Identifier.Bytes() will never return an error
	return lo.PanicOnErr(merklehasher.NewHasher[Identifier](crypto.BLAKE2b_256).ComputeProofForIndex(r.values(), 4))
}

func (r *Roots) TangleProof() *merklehasher.Proof[Identifier] {
	// We can ignore the error because Identifier.Bytes() will never return an error
	return lo.PanicOnErr(merklehasher.NewHasher[Identifier](crypto.BLAKE2b_256).ComputeProofForIndex(r.values(), 0))
}

func (r *Roots) MutationProof() *merklehasher.Proof[Identifier] {
	// We can ignore the error because Identifier.Bytes() will never return an error
	return lo.PanicOnErr(merklehasher.NewHasher[Identifier](crypto.BLAKE2b_256).ComputeProofForIndex(r.values(), 1))
}

func VerifyProof(proof *merklehasher.Proof[Identifier], proofedRoot Identifier, treeRoot Identifier) bool {
	// We can ignore the error because Identifier.Bytes() will never return an error
	if !lo.PanicOnErr(proof.ContainsValue(proofedRoot, merklehasher.NewHasher[Identifier](crypto.BLAKE2b_256))) {
		return false
	}

	return treeRoot == Identifier(proof.Hash(merklehasher.NewHasher[Identifier](crypto.BLAKE2b_256)))
}

func (r *Roots) String() string {
	return fmt.Sprintf(
		"Roots(%s): TangleRoot: %s, StateMutationRoot: %s, StateRoot: %s, AccountRoot: %s, AttestationsRoot: %s, CommitteeRoot: %s, RewardsRoot: %s, ProtocolParametersHash: %s", r.ID(), r.TangleRoot, r.StateMutationRoot, r.StateRoot, r.AccountRoot, r.AttestationsRoot, r.CommitteeRoot, r.RewardsRoot, r.ProtocolParametersHash)
}
