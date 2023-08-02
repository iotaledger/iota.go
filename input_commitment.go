package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

type CommitmentInput struct {
	CommitmentID CommitmentID `serix:"0,mapKey=commitmentId"`
}

func (c *CommitmentInput) Type() InputType {
	return InputCommitment
}

func (c *CommitmentInput) Size() int {
	// ContextInputType + CommitmentID
	return serializer.OneByte + CommitmentIDLength
}

func (c *CommitmentInput) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	workScoreBytes, err := workScoreStructure.DataByte.Multiply(c.Size())
	if err != nil {
		return 0, err
	}

	// context inputs require invocation of informations in the node, so requires extra work.
	return workScoreBytes.Add(workScoreStructure.ContextInput)
}
