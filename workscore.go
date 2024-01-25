package iotago

import (
	"github.com/iotaledger/hive.go/core/safemath"
	"github.com/iotaledger/hive.go/ierrors"
)

// WorkScore defines the type of work score used to denote the computation costs of processing an object.
type WorkScore uint32

// Add adds in to this workscore.
func (w WorkScore) Add(in ...WorkScore) (WorkScore, error) {
	var err error

	result := w
	for _, workScore := range in {
		result, err = safemath.SafeAdd(result, workScore)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to calculate WorkScore")
		}
	}

	return result, nil
}

// Multiply multiplies in with this workscore.
func (w WorkScore) Multiply(in int) (WorkScore, error) {
	result, err := safemath.SafeMul(w, WorkScore(in))
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to calculate WorkScore")
	}

	return result, nil
}

type WorkScoreParameters struct {
	// DataByte accounts for the network traffic per byte.
	DataByte WorkScore `serix:""`
	// Block accounts for work done to process a block in the node software (includes signature check for the block).
	Block WorkScore `serix:""`
	// Input accounts for loading the UTXO from the database and performing the mana calculations.
	Input WorkScore `serix:""`
	// ContextInput accounts for loading and checking the context input.
	ContextInput WorkScore `serix:""`
	// Output accounts for storing the UTXO in the database.
	Output WorkScore `serix:""`
	// NativeToken accounts for calculations done with native tokens.
	NativeToken WorkScore `serix:""`
	// Staking accounts for the existence of a staking feature in the output.
	// The node might need to update the staking vector.
	Staking WorkScore `serix:""`
	// BlockIssuer accounts for the existence of a block issuer feature in the output.
	// The node might need to update the available public keys that are allowed to issue blocks.
	BlockIssuer WorkScore `serix:""`
	// Allotment accounts for accessing the account based ledger to transform the mana to block issuance credits.
	Allotment WorkScore `serix:""`
	// SignatureEd25519 accounts for an Ed25519 signature check.
	SignatureEd25519 WorkScore `serix:""`
}

func (w WorkScoreParameters) Equals(other WorkScoreParameters) bool {
	return w.DataByte == other.DataByte &&
		w.Block == other.Block &&
		w.Input == other.Input &&
		w.ContextInput == other.ContextInput &&
		w.Output == other.Output &&
		w.NativeToken == other.NativeToken &&
		w.Staking == other.Staking &&
		w.BlockIssuer == other.BlockIssuer &&
		w.Allotment == other.Allotment &&
		w.SignatureEd25519 == other.SignatureEd25519
}

// MaxBlockWork is the maximum work score a block can have.
func (w WorkScoreParameters) MaxBlockWork() (WorkScore, error) {
	var innerErr error
	var maxBlockWork WorkScore

	addWorkScore := func(workScore WorkScore, amount int) {
		result, err := workScore.Multiply(amount)
		if err != nil {
			innerErr = err
		}

		maxBlockWork, err = maxBlockWork.Add(result)
		if err != nil {
			innerErr = err
		}
	}

	// max basic block payload size data factor
	addWorkScore(w.DataByte, MaxPayloadSize)

	// block factor
	addWorkScore(w.Block, 1)

	// inputs factor for max number of inputs
	addWorkScore(w.Input, MaxInputsCount)

	// context inputs factor for max number of inputs
	addWorkScore(w.ContextInput, MaxContextInputsCount)

	// outputs factor for max number of outputs
	addWorkScore(w.Output, MaxOutputsCount)

	// native tokens factor for max number of outputs
	addWorkScore(w.NativeToken, MaxOutputsCount)

	// staking factor for max number of outputs each with a staking feature
	addWorkScore(w.Staking, MaxOutputsCount)

	// block issuer factor for max number of outputs each with a block issuer feature
	addWorkScore(w.BlockIssuer, MaxOutputsCount)

	// allotments factor for max number of allotments
	addWorkScore(w.Allotment, MaxAllotmentCount)

	// signature check for max number of inputs each unlocked by a maximum sized mutli unlock
	// TODO: this is not correct because the signatures would not even fit in the tx.
	addWorkScore(w.SignatureEd25519, MaxInputsCount*10)

	if innerErr != nil {
		return 0, innerErr
	}

	return maxBlockWork, nil
}

type ProcessableObject interface {
	// WorkScore returns the cost this object has in terms of computation
	// requirements for a node to process it. These costs attempt to encapsulate all processing steps
	// carried out on this object throughout its life in the node.
	WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error)
}
