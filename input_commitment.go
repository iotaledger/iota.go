package iotago

import (
	"bytes"
	"cmp"

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

func (c *CommitmentInput) Compare(other ContextInput) int {
	typeCompare := cmp.Compare(c.Type(), other.Type())
	if typeCompare != 0 {
		return typeCompare
	}

	otherCommitmentInput := other.(*CommitmentInput)
	commitmentIDCompare := bytes.Compare(c.CommitmentID[:], otherCommitmentInput.CommitmentID[:])
	if commitmentIDCompare != 0 {
		return commitmentIDCompare
	}

	return 0
}
