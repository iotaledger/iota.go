package iotago

// WorkScore defines the type of work score used to denote the computation costs of processing an object.
type WorkScore uint64

type WorkScoreFactor byte

type WorkScoreStructure struct {
	FactorData          WorkScoreFactor
	FactorInput         WorkScoreFactor
	FactorAllotment     WorkScoreFactor
	FactorMissingParent WorkScoreFactor

	WorkScoreOutput           WorkScore
	WorkScoreStaking          WorkScore
	WorkScoreBlockIssuer      WorkScore
	WorkScoreEd25519Signature WorkScore
	WorkScoreNativeToken      WorkScore
	WorkScoreMaxParents       WorkScore
}

type ProcessableObject interface {
	// WorkScore returns the cost this object has in terms of computation
	// requirements for a node to process it. These costs attempt to encapsulate all processing steps
	// carried out on this object throughout its life in the node.
	WorkScore(workScoreStructure *WorkScoreStructure) WorkScore
}

// Multiply multiplies in with this factor.
func (factor WorkScoreFactor) Multiply(in int) WorkScore {
	return WorkScore(factor) * WorkScore(in)
}
