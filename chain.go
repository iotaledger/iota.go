package iotago

import "fmt"

// ChainConstrainedOutput is a type of Output which represents a chain of state transitions.
type ChainConstrainedOutput interface {
	Output
	// Chain returns the ChainID to which this Output belongs to.
	Chain() ChainID
	// ValidateStateTransition runs the state transition validation function with next.
	// Next is nil if transType is ChainTransitionTypeGenesis or ChainTransitionTypeDestroy.
	ValidateStateTransition(transType ChainTransitionType, next ChainConstrainedOutput, semValCtx *SemanticValidationContext) error

	// ImmutableFeatureSet returns the immutable FeatureSet this output contains.
	ImmutableFeatureSet() FeatureSet
}

// ChainTransitionType defines the type of transition a ChainConstrainedOutput is doing.
type ChainTransitionType byte

const (
	// ChainTransitionTypeGenesis indicates that the chain is in its genesis, aka it is new.
	ChainTransitionTypeGenesis ChainTransitionType = iota
	// ChainTransitionTypeStateChange indicates that the chain is state transitioning.
	ChainTransitionTypeStateChange
	// ChainTransitionTypeDestroy indicates that the chain is being destroyed.
	ChainTransitionTypeDestroy
)

// ChainConstrainedOutputs is a slice of ChainConstrainedOutput.
type ChainConstrainedOutputs []ChainConstrainedOutput

// ChainConstrainedOutputsSet is a map of ChainID to ChainConstrainedOutput.
type ChainConstrainedOutputsSet map[ChainID]ChainConstrainedOutput

// Includes checks whether all chains included in other exist in this set.
func (set ChainConstrainedOutputsSet) Includes(other ChainConstrainedOutputsSet) error {
	for chainID := range other {
		if _, has := set[chainID]; !has {
			return fmt.Errorf("%w: %s missing in source", ErrChainMissing, chainID)
		}
	}
	return nil
}

// Merge merges other with this set in a new set.
// Returns an error if a chain isn't unique across both sets.
func (set ChainConstrainedOutputsSet) Merge(other ChainConstrainedOutputsSet) (ChainConstrainedOutputsSet, error) {
	newSet := make(ChainConstrainedOutputsSet)
	for k, v := range set {
		newSet[k] = v
	}
	for k, v := range other {
		if _, has := newSet[k]; has {
			return nil, fmt.Errorf("%w: chain %s exists in both sets", ErrNonUniqueChainConstrainedOutputs, k)
		}
		newSet[k] = v
	}
	return newSet, nil
}

// IsIssuerOnOutputUnlocked checks whether the issuer in an IssuerFeature of this new ChainConstrainedOutput has been unlocked.
// This function is a no-op if the chain output does not contain an IssuerFeature.
func IsIssuerOnOutputUnlocked(output ChainConstrainedOutput, unlockedIdents UnlockedIdentities) error {
	immFeats := output.ImmutableFeatureSet()
	if len(immFeats) == 0 {
		return nil
	}

	issuerFeat := immFeats.IssuerFeature()
	if issuerFeat == nil {
		return nil
	}
	if _, isUnlocked := unlockedIdents[issuerFeat.Address.Key()]; !isUnlocked {
		return ErrIssuerFeatureNotUnlocked
	}
	return nil
}
