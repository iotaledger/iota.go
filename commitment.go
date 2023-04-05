package iotago

import (
	"context"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ds/types"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	"github.com/iotaledger/iota.go/v4/slot"
)

const (
	CommitmentSize = marshalutil.Int64Size + CommitmentIDLength + 32 + marshalutil.Uint64Size
)

type Commitment struct {
	Index            slot.Index `serix:"0"`
	PrevID           ID         `serix:"1"`
	RootsID          [32]byte   `serix:"2"`
	CumulativeWeight int64      `serix:"3"`
}

func New(index slot.Index, prevID ID, rootsID types.Identifier, cumulativeWeight int64) *Commitment {
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

// FromBytes deserializes a ID from a byte slice.
func (c *Commitment) FromBytes(serialized []byte) (consumedBytes int, err error) {
	return serix.DefaultAPI.Decode(context.Background(), serialized, c, serix.WithValidation())
}

// Length returns the byte length of a serialized ID.
func (c Commitment) Length() int {
	return marshalutil.Int64Size + CommitmentIDLength
}

// Bytes returns a serialized version of the ID.
func (c Commitment) Bytes() (serialized []byte, err error) {
	return serix.DefaultAPI.Encode(context.Background(), c, serix.WithValidation())
}

func (c *Commitment) ID() (id ID) {
	return NewID(c.Index, blake2b.Sum256(lo.PanicOnErr(c.Bytes())))
}

func (c *Commitment) Equals(other *Commitment) bool {
	return c.ID() == other.ID() &&
		c.PrevID == other.PrevID &&
		c.Index == other.Index &&
		c.RootsID == other.RootsID &&
		c.CumulativeWeight == other.CumulativeWeight
}
