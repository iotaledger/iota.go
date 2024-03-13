package iotago

// ChainOutput is a type of Output which represents a chain of state transitions.
type ChainOutput interface {
	Output
	// ChainID returns the ChainID to which this Output belongs to.
	ChainID() ChainID
}

// ChainOutputImmutable is a type of Output which represents a chain of state transitions with immutable features.
type ChainOutputImmutable interface {
	ChainOutput
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

// ChainOutputSet is a map of ChainID to ChainOutput.
type ChainOutputSet map[ChainID]ChainOutput
