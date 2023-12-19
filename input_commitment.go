package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

type CommitmentInput struct {
	CommitmentID CommitmentID `serix:""`
}

func (c *CommitmentInput) Clone() ContextInput {
	return &CommitmentInput{
		CommitmentID: c.CommitmentID,
	}
}

func (c *CommitmentInput) Type() ContextInputType {
	return ContextInputCommitment
}

func (c *CommitmentInput) IsReadOnly() bool {
	return true
}

func (c *CommitmentInput) Size() int {
	// ContextInputType + CommitmentID
	return serializer.OneByte + CommitmentIDLength
}

func (c *CommitmentInput) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	// context inputs require invocation of informations in the node, so requires extra work.
	return workScoreParameters.ContextInput, nil
}
