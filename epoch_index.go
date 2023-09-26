package iotago

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrWrongEpochIndex gets returned when a wrong epoch index was passed.
	ErrWrongEpochIndex = ierrors.New("wrong epoch index")
)

const EpochIndexLength = serializer.UInt32ByteSize

const MaxEpochIndex = EpochIndex(math.MaxUint32)

// EpochIndex is the index of an epoch.
type EpochIndex uint32

func EpochIndexFromBytes(b []byte) (EpochIndex, int, error) {
	if len(b) < EpochIndexLength {
		return 0, 0, ierrors.New("invalid epoch index size")
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
