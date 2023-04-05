package iotago

import (
	"bytes"
	"context"
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	"github.com/iotaledger/iota.go/v4/slot"
)

const (
	CommitmentIDLength = marshalutil.Int64Size + 32
)

type ID struct {
	SlotIndex  slot.Index `serix:"0"`
	Identifier [32]byte   `serix:"1"`
}

func NewID(index slot.Index, idBytes [32]byte) (newCommitmentID ID) {
	newCommitmentID.SlotIndex = index
	copy(newCommitmentID.Identifier[:], idBytes[:])

	return
}

func (b ID) Index() slot.Index {
	return b.SlotIndex
}

func (b ID) EncodeJSON() (any, error) {
	return b.String(), nil
}

func (b *ID) DecodeJSON(val any) error {
	serialized, ok := val.(string)
	if !ok {
		return errors.New("incorrect type")
	}

	decoded, err := hexutil.Decode(serialized)
	if err != nil {
		return err
	}

	copy(b.Identifier[:], decoded[marshalutil.Uint64Size:])
	b.SlotIndex = slot.Index(binary.LittleEndian.Uint64(decoded[:marshalutil.Uint64Size]))

	return nil
}

// FromBytes deserializes a ID from a byte slice.
func (b *ID) FromBytes(serialized []byte) (consumedBytes int, err error) {
	return serix.DefaultAPI.Decode(context.Background(), serialized, b, serix.WithValidation())
}

// Bytes returns a serialized version of the ID.
func (b ID) Bytes() (serialized []byte, err error) {
	return serix.DefaultAPI.Encode(context.Background(), b, serix.WithValidation())
}

// String returns a human-readable version of the ID.
func (b ID) String() string {
	encoded := [CommitmentIDLength]byte{}
	binary.LittleEndian.PutUint64(encoded[0:], uint64(b.SlotIndex))
	copy(encoded[8:], b.Identifier[:])

	return hexutil.Encode(encoded[:])
}

// CompareTo does a lexicographical comparison to another blockID.
// Returns 0 if equal, -1 if smaller, or 1 if larger than other.
// Passing nil as other will result in a panic.
func (b ID) CompareTo(other ID) int {
	return bytes.Compare(lo.PanicOnErr(b.Bytes()), lo.PanicOnErr(other.Bytes()))
}
