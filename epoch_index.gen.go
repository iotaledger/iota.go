package iotago

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	EpochIndexLength = serializer.UInt32ByteSize
	MaxEpochIndex    = EpochIndex(math.MaxUint32)
)

// EpochIndex is the index of an epoch.
type EpochIndex uint32

func EpochIndexFromBytes(b []byte) (EpochIndex, int, error) {
	if len(b) < EpochIndexLength {
		return 0, 0, ierrors.Errorf("invalid length for epoch index, expected at least %d bytes, got %d bytes", EpochIndexLength, len(b))
	}

	return EpochIndex(binary.LittleEndian.Uint32(b)), EpochIndexLength, nil
}

func (i EpochIndex) Bytes() ([]byte, error) {
	bytes := make([]byte, EpochIndexLength)
	binary.LittleEndian.PutUint32(bytes, uint32(i))

	return bytes, nil
}

func (i EpochIndex) MustBytes() []byte {
	return lo.PanicOnErr(i.Bytes())
}

func (i EpochIndex) String() string {
	return fmt.Sprintf("EpochIndex(%d)", i)
}
