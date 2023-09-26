package iotago

import (
	"context"
	"fmt"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	CommitmentIDLength = SlotIdentifierLength
)

type CommitmentID = SlotIdentifier

var EmptyCommitmentID = CommitmentID{}

type Commitment struct {
	ProtocolVersion Version `serix:"0,mapKey=protocolVersion"`
	// TODO: rename to Slot?
	Index                SlotIndex    `serix:"1,mapKey=index"`
	PreviousCommitmentID CommitmentID `serix:"2,mapKey=previousCommitmentId"`
	RootsID              Identifier   `serix:"3,mapKey=rootsId"`
	CumulativeWeight     uint64       `serix:"4,mapKey=cumulativeWeight"`
	ReferenceManaCost    Mana         `serix:"5,mapKey=referenceManaCost"`
}

func NewCommitment(version Version, slot SlotIndex, prevID CommitmentID, rootsID Identifier, cumulativeWeight uint64, rmc Mana) *Commitment {
	return &Commitment{
		ProtocolVersion:      version,
		Index:                slot,
		PreviousCommitmentID: prevID,
		RootsID:              rootsID,
		CumulativeWeight:     cumulativeWeight,
		ReferenceManaCost:    rmc,
	}
}

func NewEmptyCommitment(version Version) *Commitment {
	return &Commitment{
		ProtocolVersion: version,
	}
}

func (c *Commitment) ID() (CommitmentID, error) {
	data, err := CommonSerixAPI().Encode(context.TODO(), c)
	if err != nil {
		return CommitmentID{}, ierrors.Errorf("can't compute commitment ID: %w", err)
	}

	return SlotIdentifierRepresentingData(c.Index, data), nil
}

func (c *Commitment) StateID() Identifier {
	return IdentifierFromData(lo.PanicOnErr(c.MustID().Bytes()))
}

func (c *Commitment) Type() StateType {
	return InputCommitment
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
		c.ProtocolVersion == other.ProtocolVersion &&
		c.Index == other.Index &&
		c.PreviousCommitmentID == other.PreviousCommitmentID &&
		c.RootsID == other.RootsID &&
		c.CumulativeWeight == other.CumulativeWeight &&
		c.ReferenceManaCost == other.ReferenceManaCost
}

func (c *Commitment) String() string {
	return fmt.Sprintf("Commitment{\n\tIndex: %d\n\tPrevID: %s\n\tRootsID: %s\n\tCumulativeWeight: %d\n\tRMC: %d\n}",
		c.Index, c.PreviousCommitmentID, c.RootsID, c.CumulativeWeight, c.ReferenceManaCost)
}

func (c *Commitment) Size() int {
	return serializer.OneByte +
		SlotIndexLength +
		CommitmentIDLength +
		IdentifierLength +
		serializer.UInt64ByteSize +
		ManaSize
}
