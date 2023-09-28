package vm

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v4"
)

// InputSet is a map of OutputID to Output.
type InputSet map[iotago.OutputID]iotago.Output

func (inputSet InputSet) OutputSet() iotago.OutputSet {
	outputs := make(iotago.OutputSet, len(inputSet))
	for outputID := range inputSet {
		outputs[outputID] = inputSet[outputID]
	}

	return outputs
}

type ChainOutput struct {
	ChainID  iotago.ChainID
	OutputID iotago.OutputID
	Output   iotago.ChainOutput
}

// ChainInputSet returns a ChainInputSet for all ChainOutputs in the InputSet.
func (inputSet InputSet) ChainInputSet() ChainInputSet {
	set := make(ChainInputSet)
	for utxoInputID, input := range inputSet {
		chainOutput, is := input.(iotago.ChainOutput)
		if !is {
			continue
		}

		chainID := chainOutput.ChainID()
		if chainID.Empty() {
			if utxoIDChainID, is := chainID.(iotago.UTXOIDChainID); is {
				chainID = utxoIDChainID.FromOutputID(utxoInputID)
			}
		}

		if chainID.Empty() {
			panic(fmt.Sprintf("output of type %s has empty chain ID but is not utxo dependable", chainOutput.Type()))
		}

		set[chainID] = &ChainOutput{
			ChainID:  chainID,
			OutputID: utxoInputID,
			Output:   chainOutput,
		}
	}

	return set
}

// ChainInputSet is a map of ChainID to ChainOutput.
type ChainInputSet map[iotago.ChainID]*ChainOutput

type BlockIssuanceCreditInputSet map[iotago.AccountID]iotago.BlockIssuanceCredits

// A map of either DelegationID or AccountID to their mana reward amount.
type RewardsInputSet map[iotago.ChainID]iotago.Mana

//nolint:revive // the VM at the beginning makes it more clear
type VMCommitmentInput *iotago.Commitment

type ResolvedInputs struct {
	InputSet
	BlockIssuanceCreditInputSet
	CommitmentInput VMCommitmentInput
	RewardsInputSet
}

type ImplicitAccountOutput struct {
	*iotago.BasicOutput `serix:"0,nest=false"`
}

func (o *ImplicitAccountOutput) ChainID() iotago.ChainID {
	return iotago.EmptyAccountID()
}
