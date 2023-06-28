package iotago

import (
	"fmt"
)

const (
	CommitmentIDLength = SlotIdentifierLength
)

type CommitmentID = SlotIdentifier

type Commitment struct {
	Index            SlotIndex    `serix:"0,mapKey=index"`
	PrevID           CommitmentID `serix:"1,mapKey=prevID"`
	RootsID          Identifier   `serix:"2,mapKey=rootsID"`
	CumulativeWeight uint64       `serix:"3,mapKey=cumulativeWeight"`
}

func NewCommitment(index SlotIndex, prevID CommitmentID, rootsID Identifier, cumulativeWeight uint64) *Commitment {
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

func (c *Commitment) ID(api API) (CommitmentID, error) {
	data, err := api.Encode(c)
	if err != nil {
		return CommitmentID{}, fmt.Errorf("can't compute commitment ID: %w", err)
	}
	return SlotIdentifierRepresentingData(c.Index, data), nil
}

func (c *Commitment) MustID(api API) CommitmentID {
	id, err := c.ID(api)
	if err != nil {
		panic(err)
	}

	return id
}

func (c *Commitment) Equals(api API, other *Commitment) bool {
	return c.MustID(api) == other.MustID(api) &&
		c.PrevID == other.PrevID &&
		c.Index == other.Index &&
		c.RootsID == other.RootsID &&
		c.CumulativeWeight == other.CumulativeWeight
}

func (c *Commitment) Bytes(api API) ([]byte, error) {
	return api.Encode(c)
}

func (c *Commitment) FromBytes(api API, bytes []byte) (int, error) {
	return api.Decode(bytes, c)
}

func (c *Commitment) String() string {
	return fmt.Sprintf("Commitment{\n\tIndex: %d\n\tPrevID: %s\n\tRootsID: %s\n\tCumulativeWeight: %d\n}",
		c.Index, c.PrevID, c.RootsID, c.CumulativeWeight)
}
