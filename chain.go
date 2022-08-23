package iotago

import "fmt"

// ChainConstrainedOutputs is a slice of ChainConstrainedOutput.
type ChainConstrainedOutputs []ChainConstrainedOutput

// ChainConstrainedOutputsSet is a map of ChainID to ChainConstrainedOutput.
type ChainConstrainedOutputsSet map[ChainID]ChainConstrainedOutput

// Includes checks whether all chains included in other exist in this set.
func (set ChainConstrainedOutputsSet) Includes(other ChainConstrainedOutputsSet) error {
	for chainID := range other {
		if _, has := set[chainID]; !has {
			return fmt.Errorf("%w: %s missing in source", ErrChainMissing, chainID.ToHex())
		}
	}
	return nil
}

// EveryTuple runs f for every key which exists in both this set and other.
func (set ChainConstrainedOutputsSet) EveryTuple(other ChainConstrainedOutputsSet, f func(in ChainConstrainedOutput, out ChainConstrainedOutput) error) error {
	for k, v := range set {
		v2, has := other[k]
		if !has {
			continue
		}
		if err := f(v, v2); err != nil {
			return err
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
			return nil, fmt.Errorf("%w: chain %s exists in both sets", ErrNonUniqueChainConstrainedOutputs, k.ToHex())
		}
		newSet[k] = v
	}
	return newSet, nil
}

// ChainConstrainedOutput is a type of Output which represents a chain of state transitions.
type ChainConstrainedOutput interface {
	Output
	// Chain returns the ChainID to which this Output belongs to.
	Chain() ChainID
	// ValidateStateTransition runs a StateTransitionValidationFunc with next.
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

// StateTransitionValidationFunc is a function which given the current and next chain state,
// validates the state transition.
type StateTransitionValidationFunc func(current ChainConstrainedOutput, next ChainConstrainedOutput) error

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

// FeatureSetTransitionValidationFunc checks whether the Features transition from in to out is valid.
type FeatureSetTransitionValidationFunc func(inSet FeatureSet, outSet FeatureSet) error

// FeatureUnchanged checks whether the specified Feature type is unchanged between in and out.
// Unchanged also means that the block's existence is unchanged between both sets.
func FeatureUnchanged(featType FeatureType, inFeatSet FeatureSet, outFeatSet FeatureSet) error {
	in, inHas := inFeatSet[featType]
	out, outHas := outFeatSet[featType]

	switch {
	case outHas && !inHas:
		return fmt.Errorf("%w: %s in next state but not in previous", ErrInvalidFeatureTransition, featType)
	case !outHas && inHas:
		return fmt.Errorf("%w: %s in current state but not in next", ErrInvalidFeatureTransition, featType)
	}

	// not in both sets
	if in == nil {
		return nil
	}

	if !in.Equal(out) {
		return fmt.Errorf("%w: %s changed, in %v / out %v", ErrInvalidFeatureTransition, featType, in, out)
	}

	return nil
}
