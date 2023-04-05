package iotago

import (
	"encoding/binary"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
)

const (
	CommitmentIDLength = marshalutil.Int64Size + IdentifierLength
)

type CommitmentID struct {
	SlotIndex  SlotIndex  `serix:"0"`
	Identifier Identifier `serix:"1"`
}

func NewCommitmentID(index SlotIndex, idBytes [32]byte) (newCommitmentID CommitmentID) {
	newCommitmentID.SlotIndex = index
	copy(newCommitmentID.Identifier[:], idBytes[:])

	return
}

func (b CommitmentID) Index() SlotIndex {
	return b.SlotIndex
}

// ToHex returns a human-readable version of the ID in hex.
func (b CommitmentID) ToHex() string {
	encoded := [CommitmentIDLength]byte{}
	binary.LittleEndian.PutUint64(encoded[0:], uint64(b.SlotIndex))
	copy(encoded[marshalutil.Int64Size:], b.Identifier[:])

	return EncodeHex(encoded[:])
}
