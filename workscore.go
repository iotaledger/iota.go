package iotago

// WorkScore defines the type of work score used to denote the computation costs of processing an object.
type WorkScore uint64

type WorkScoreFactor uint16

type WorkScoreStructure struct {
	WorkScores WorkScores       `serix:"0,mapKey=workScores"`
	Factors    WorkScoreFactors `serix:"1,mapKey=factors"`

	MinStrongParentsThreshold byte `serix:"2,mapKey=missingParentsThreshold"`
}

type WorkScoreFactors struct {
	Data          WorkScoreFactor `serix:"0,mapKey=data"`
	Input         WorkScoreFactor `serix:"1,mapKey=input"`
	ContextInput  WorkScoreFactor `serix:"2,mapKey=contextInput"`
	Allotment     WorkScoreFactor `serix:"3,mapKey=allotment"`
	MissingParent WorkScoreFactor `serix:"4,mapKey=missingParent"`
}

type WorkScores struct {
	Output           WorkScore `serix:"0,mapKey=output"`
	Staking          WorkScore `serix:"1,mapKey=staking"`
	BlockIssuer      WorkScore `serix:"2,mapKey=blockIssuer"`
	Ed25519Signature WorkScore `serix:"3,mapKey=ed25519Signature"`
	NativeToken      WorkScore `serix:"4,mapKey=nativeToken"`
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
