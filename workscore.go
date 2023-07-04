package iotago

// WorkScore defines the type of work score used to denote the computation costs of processing an object.
type WorkScore uint64

type WorkScoreFactor byte

type WorkScoreStructure struct {
	FactorData          WorkScoreFactor `serix:"0,mapKey=factorData"`
	FactorInput         WorkScoreFactor `serix:"1,mapKey=factorInput"`
	FactorAllotment     WorkScoreFactor `serix:"2,mapKey=factorAllotment"`
	FactorMissingParent WorkScoreFactor `serix:"3,mapKey=factorMissingParent"`

	WorkScoreOutput           WorkScore `serix:"4,mapKey=workScoreOutput"`
	WorkScoreStaking          WorkScore `serix:"5,mapKey=workScoreStaking"`
	WorkScoreBlockIssuer      WorkScore `serix:"6,mapKey=workScoreBlockIssuer"`
	WorkScoreEd25519Signature WorkScore `serix:"7,mapKey=workScoreEd25519Signature"`
	WorkScoreNativeToken      WorkScore `serix:"8,mapKey=workScoreNativeToken"`
	WorkScoreMaxParents       WorkScore `serix:"9,mapKey=workScoreMaxParents"`
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
