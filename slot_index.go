package iotago

import (
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// SlotIndex is the ID of a slot.
type SlotIndex uint64

func SlotIndexFromBytes(b []byte) (SlotIndex, error) {
	if len(b) != serializer.UInt64ByteSize {
		return 0, errors.New("invalid slot index size")
	}

	return SlotIndex(binary.LittleEndian.Uint64(b)), nil
}

func (i SlotIndex) Bytes() []byte {
	bytes := make([]byte, serializer.UInt64ByteSize)
	binary.LittleEndian.PutUint64(bytes[:], uint64(i))

	return bytes
}

func (i SlotIndex) String() string {
	return fmt.Sprintf("SlotIndex(%d)", i)
}
