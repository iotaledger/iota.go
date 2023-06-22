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
			Output:       chainOutput,
			CreationTime: input.CreationTime,
		}
	}
	return set
}

// ChainInputSet is a map of ChainID to ChainOutputWithCreationTime.
type ChainInputSet map[iotago.ChainID]*ChainOutputWithCreationTime

type BICInputSet map[iotago.AccountID]BlockIssuanceCredit

// A map of either DelegationID or AccountID to their mana reward amount.
type RewardsInputSet map[iotago.ChainID]uint64

type BlockIssuanceCredit struct {
	AccountID    iotago.AccountID
	CommitmentID iotago.CommitmentID
	Value        int64
}

func (b BlockIssuanceCredit) Negative() bool {
	return b.Value < 0
}

type CommitmentInputSet map[iotago.CommitmentID]*iotago.Commitment

type ResolvedInputs struct {
	InputSet
	BICInputSet
	CommitmentInputSet
	RewardsInputSet
}
