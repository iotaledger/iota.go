package vm

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v4"
)

type OutputWithCreationTime struct {
	Output       iotago.Output
	CreationTime iotago.SlotIndex
}

// InputSet is a map of OutputID to OutputWithCreationTime.
type InputSet map[iotago.OutputID]OutputWithCreationTime

func (inputSet InputSet) OutputSet() iotago.OutputSet {
	outputs := make(iotago.OutputSet, len(inputSet))
	for outputID := range inputSet {
		outputs[outputID] = inputSet[outputID].Output
	}

	return outputs
}

type ChainOutputWithCreationTime struct {
	ChainID      iotago.ChainID
	Output       iotago.ChainOutput
	CreationTime iotago.SlotIndex
}

// ChainInputSet returns a ChainInputSet for all ChainOutputs in the InputSet.
func (inputSet InputSet) ChainInputSet() ChainInputSet {
	set := make(ChainInputSet)
	for utxoInputID, input := range inputSet {
		chainOutput, is := input.Output.(iotago.ChainOutput)
		if !is {
			continue
		}

		chainID := chainOutput.Chain()
		if chainID.Empty() {
			if utxoIDChainID, is := chainOutput.Chain().(iotago.UTXOIDChainID); is {
				chainID = utxoIDChainID.FromOutputID(utxoInputID)
			}
		}

		if chainID.Empty() {
			panic(fmt.Sprintf("output of type %s has empty chain ID but is not utxo dependable", chainOutput.Type()))
		}

		set[chainID] = &ChainOutputWithCreationTime{
			ChainID:      chainID,
			Output:       chainOutput,
			CreationTime: input.CreationTime,
		}
	}

	return set
}

// ChainInputSet is a map of ChainID to ChainOutputWithCreationTime.
type ChainInputSet map[iotago.ChainID]*ChainOutputWithCreationTime

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
