package iotago

import (
	"fmt"
)

// ChainOutput is a type of Output which represents a chain of state transitions.
type ChainOutput interface {
	Output
	// Chain returns the ChainID to which this Output belongs to.
	Chain() ChainID
	// ImmutableFeatureSet returns the immutable FeatureSet this output contains.
	ImmutableFeatureSet() FeatureSet
}

// ChainTransitionType defines the type of transition a ChainOutput is doing.
type ChainTransitionType byte

const (
	// ChainTransitionTypeGenesis indicates that the chain is in its genesis, aka it is new.
	ChainTransitionTypeGenesis ChainTransitionType = iota
	// ChainTransitionTypeStateChange indicates that the chain is state transitioning.
	ChainTransitionTypeStateChange
	// ChainTransitionTypeDestroy indicates that the chain is being destroyed.
	ChainTransitionTypeDestroy
)

// ChainOutputs is a slice of ChainOutput.
type ChainOutputs []ChainOutput

// ChainOutputSet is a map of ChainID to ChainOutput.
type ChainOutputSet map[ChainID]ChainOutput

// Includes checks whether all chains included in other exist in this set.
func (set ChainOutputSet) Includes(other ChainOutputSet) error {
	for chainID := range other {
		if _, has := set[chainID]; !has {
			return fmt.Errorf("%w: %s missing in source", ErrChainMissing, chainID)
		}
	}
	return nil
}

// Merge merges other with this set in a new set.
// Returns an error if a chain isn't unique across both sets.
func (set ChainOutputSet) Merge(other ChainOutputSet) (ChainOutputSet, error) {
	newSet := make(ChainOutputSet)
	for k, v := range set {
		newSet[k] = v
	}
	for k, v := range other {
		if _, has := newSet[k]; has {
			return nil, fmt.Errorf("%w: chain %s exists in both sets", ErrNonUniqueChainOutputs, k)
		}
		newSet[k] = v
	}
	return newSet, nil
}
