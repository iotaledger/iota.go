package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"

	"golang.org/x/crypto/blake2b"
)

const (
	// TreasuryInputBytesLength is the length of a TreasuryInput.
	TreasuryInputBytesLength = blake2b.Size256
	// TreasuryInputSerializedBytesSize is the size of a serialized TreasuryInput with its type denoting byte.
	TreasuryInputSerializedBytesSize = serializer.SmallTypeDenotationByteSize + TreasuryInputBytesLength
)

// TreasuryInput is an input which references a milestone which generated a TreasuryOutput.
type TreasuryInput [32]byte

func (ti *TreasuryInput) Decode(b []byte) (int, error) {
	copy(ti[:], b)
	return TreasuryInputSerializedBytesSize - 1, nil
}

func (ti *TreasuryInput) Encode() ([]byte, error) {
	var b [TreasuryInputSerializedBytesSize - 1]byte
	copy(b[:], ti[:])
	return b[:], nil
}

func (ti *TreasuryInput) Type() InputType {
	return InputTreasury
}

func (ti *TreasuryInput) Clone() *TreasuryInput {
	p := *ti
	return &p
}

func (ti *TreasuryInput) Size() int {
	return TreasuryInputSerializedBytesSize
}
