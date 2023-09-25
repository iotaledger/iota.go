package iotago

import (
	"encoding/binary"
	"fmt"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
)

const SlotIndexLength = serializer.UInt32ByteSize

// SlotIndex is the ID of a slot.
type SlotIndex uint32

func SlotIndexFromBytes(b []byte) (SlotIndex, int, error) {
	if len(b) < SlotIndexLength {
		return 0, 0, ierrors.New("invalid slot index size")
	}

	return SlotIndex(binary.LittleEndian.Uint32(b)), SlotIndexLength, nil
}

func (i SlotIndex) Bytes() ([]byte, error) {
	bytes := make([]byte, SlotIndexLength)
	binary.LittleEndian.PutUint32(bytes, uint32(i))

	return bytes, nil
}

func (i SlotIndex) MustBytes() []byte {
	return lo.PanicOnErr(i.Bytes())
}

func (i SlotIndex) String() string {
	return fmt.Sprintf("SlotIndex(%d)", i)
}
