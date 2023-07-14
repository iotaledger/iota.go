package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
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

func (c *Commitment) ID() (CommitmentID, error) {
	data, err := internalEncode(c)
	if err != nil {
		return CommitmentID{}, ierrors.Errorf("can't compute commitment ID: %w", err)
	}
	return SlotIdentifierRepresentingData(c.Index, data), nil
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

func (c *Commitment) Bytes() ([]byte, error) {
	return internalEncode(c)
}

func (c *Commitment) FromBytes(bytes []byte) (int, error) {
	return internalDecode(bytes, c)
}

func (c *Commitment) String() string {
	return fmt.Sprintf("Commitment{\n\tIndex: %d\n\tPrevID: %s\n\tRootsID: %s\n\tCumulativeWeight: %d\n}",
		c.Index, c.PrevID, c.RootsID, c.CumulativeWeight)
}

func (c *Commitment) Size() int {
	return 2*serializer.UInt64ByteSize +
		SlotIdentifierLength +
		IdentifierLength

}
