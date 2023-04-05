package iotago

import (
	"fmt"

	"golang.org/x/crypto/blake2b"
)

type Commitment struct {
	Index            SlotIndex    `serix:"0,mapKey=index"`
	PrevID           CommitmentID `serix:"1,mapKey=prevID"`
	RootsID          Identifier   `serix:"2,mapKey=rootsID"`
	CumulativeWeight int64        `serix:"3,mapKey=cumulativeWeight"`
}

func New(index SlotIndex, prevID CommitmentID, rootsID Identifier, cumulativeWeight int64) *Commitment {
	return &Commitment{
		Index:            index,
		PrevID:           prevID,
		RootsID:          rootsID,
		CumulativeWeight: cumulativeWeight,
	}
}

func NewEmptyCommitment() *Commitment {
	return &Commitment{}
}

func (c *Commitment) ID() (CommitmentID, error) {
	data, err := internalEncode(c)
	if err != nil {
		return CommitmentID{}, fmt.Errorf("can't compute commitment ID: %w", err)
	}
	return NewCommitmentID(c.Index, blake2b.Sum256(data)), nil
}

func (c *Commitment) MustID() CommitmentID {
	id, err := c.ID()
	if err != nil {
		panic(err)
	}

	return id
}

func (c *Commitment) Equals(other *Commitment) bool {
	return c.MustID() == other.MustID() &&
		c.PrevID == other.PrevID &&
		c.Index == other.Index &&
		c.RootsID == other.RootsID &&
		c.CumulativeWeight == other.CumulativeWeight
}
