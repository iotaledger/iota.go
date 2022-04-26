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
			return fmt.Errorf("%w: %s missing in source", ErrChainMissing, chainID)
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
			return nil, fmt.Errorf("%w: chain %s exists in both sets", ErrNonUniqueChainConstrainedOutputs, k)
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

	// ImmutableFeatureBlocks returns the immutable FeatureBlocks this output contains.
	ImmutableFeatureBlocks() FeatureBlocks
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

// IsIssuerOnOutputUnlocked checks whether the issuer in an IssuerFeatureBlock of this new ChainConstrainedOutput has been unlocked.
// This function is a no-op if the chain output does not contain an IssuerFeatureBlock.
func IsIssuerOnOutputUnlocked(output ChainConstrainedOutput, unlockedIdents UnlockedIdentities) error {
	immFeatblocks := output.ImmutableFeatureBlocks()
	if immFeatblocks == nil || len(immFeatblocks) == 0 {
		return nil
	}

	issuerFeatureBlock := immFeatblocks.MustSet().IssuerFeatureBlock()
	if issuerFeatureBlock == nil {
		return nil
	}
	if _, isUnlocked := unlockedIdents[issuerFeatureBlock.Address.Key()]; !isUnlocked {
		return ErrIssuerFeatureBlockNotUnlocked
	}
	return nil
}

// FeatureBlockSetTransitionValidationFunc checks whether the FeatureBlocks transition from in to out is valid.
type FeatureBlockSetTransitionValidationFunc func(inSet FeatureBlocksSet, outSet FeatureBlocksSet) error

// FeatureBlockUnchanged checks whether the specified FeatureBlock type is unchanged between in and out.
// Unchanged also means that the block's existence is unchanged between both sets.
func FeatureBlockUnchanged(featBlockType FeatureBlockType, inBlockSet FeatureBlocksSet, outBlockSet FeatureBlocksSet) error {
	in, inHas := inBlockSet[featBlockType]
	out, outHas := outBlockSet[featBlockType]

	switch {
	case outHas && !inHas:
		return fmt.Errorf("%w: %s in next state but not in previous", ErrInvalidFeatureBlockTransition, featBlockType)
	case !outHas && inHas:
		return fmt.Errorf("%w: %s in current state but not in next", ErrInvalidFeatureBlockTransition, featBlockType)
	}

	// not in both sets
	if in == nil {
		return nil
	}

	if !in.Equal(out) {
		return fmt.Errorf("%w: %s changed, in %v / out %v", ErrInvalidFeatureBlockTransition, featBlockType, in, out)
	}

	return nil
}
