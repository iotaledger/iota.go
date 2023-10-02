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

type WorkScoreStructure struct {
	// DataByte accounts for the network traffic per byte.
	DataByte WorkScore `serix:"0,mapKey=dataByte"`
	// Block accounts for work done to process a block in the node software.
	Block WorkScore `serix:"1,mapKey=block"`
	// Input accounts for loading the UTXO from the database and performing the mana calculations.
	Input WorkScore `serix:"2,mapKey=input"`
	// ContextInput accounts for loading and checking the context input.
	ContextInput WorkScore `serix:"3,mapKey=contextInput"`
	// Output accounts for storing the UTXO in the database.
	Output WorkScore `serix:"4,mapKey=output"`
	// NativeToken accounts for calculations done with native tokens.
	NativeToken WorkScore `serix:"5,mapKey=nativeToken"`
	// Staking accounts for the existence of a staking feature in the output.
	// The node might need to update the staking vector.
	Staking WorkScore `serix:"6,mapKey=staking"`
	// BlockIssuer accounts for the existence of a block issuer feature in the output.
	// The node might need to update the available public keys that are allowed to issue blocks.
	BlockIssuer WorkScore `serix:"7,mapKey=blockIssuer"`
	// Allotment accounts for accessing the account based ledger to transform the mana to block issuance credits.
	Allotment WorkScore `serix:"8,mapKey=allotment"`
	// SignatureEd25519 accounts for an Ed25519 signature check.
	SignatureEd25519 WorkScore `serix:"9,mapKey=signatureEd25519"`
}

func (w WorkScoreStructure) Equals(other WorkScoreStructure) bool {
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
func (w WorkScoreStructure) MaxBlockWork() (WorkScore, error) {
	var maxBlockWork WorkScore
	// max basic block payload size data factor
	dataFactorBytes, err := w.DataByte.Multiply(MaxPayloadSize)
	if err != nil {
		return 0, err
	}
	maxBlockWork += dataFactorBytes
	// block factor
	maxBlockWork += w.Block
	// inputs factor for max number of inputs
	inputsFactor, err := w.Input.Multiply(MaxInputsCount)
	if err != nil {
		return 0, err
	}
	maxBlockWork += inputsFactor
	// context inputs factor for max number of inputs
	contextInputsFactor, err := w.ContextInput.Multiply(MaxContextInputsCount)
	if err != nil {
		return 0, err
	}
	maxBlockWork += contextInputsFactor
	// outputs factor for max number of outputs
	outputsFactor, err := w.Output.Multiply(MaxOutputsCount)
	if err != nil {
		return 0, err
	}
	maxBlockWork += outputsFactor
	// native tokens factor for max number of outputs
	nativeTokensFactor, err := w.NativeToken.Multiply(MaxOutputsCount)
	if err != nil {
		return 0, err
	}
	maxBlockWork += nativeTokensFactor
	// staking factor for max number of outputs each with a staking feature
	stakingFactor, err := w.Staking.Multiply(MaxOutputsCount)
	if err != nil {
		return 0, err
	}
	maxBlockWork += stakingFactor
	// block issuer factor for max number of outputs each with a block issuer feature
	blockIssuerFactor, err := w.BlockIssuer.Multiply(MaxOutputsCount)
	if err != nil {
		return 0, err
	}
	maxBlockWork += blockIssuerFactor
	// allotments factor for max number of allotments
	allotmentsFactor, err := w.Allotment.Multiply(MaxAllotmentCount)
	if err != nil {
		return 0, err
	}
	maxBlockWork += allotmentsFactor
	// signature for block and max number of inputs
	signatureFactor, err := w.SignatureEd25519.Multiply(1 + MaxInputsCount)
	if err != nil {
		return 0, err
	}
	maxBlockWork += signatureFactor

	return maxBlockWork, nil
}

type ProcessableObject interface {
	// WorkScore returns the cost this object has in terms of computation
	// requirements for a node to process it. These costs attempt to encapsulate all processing steps
	// carried out on this object throughout its life in the node.
	WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error)
}
