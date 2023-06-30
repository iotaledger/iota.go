package iotago

import (
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrWrongEpochIndex gets returned when a wrong epoch index was passed.
	ErrWrongEpochIndex = errors.New("wrong epoch index")
)

const EpochIndexLength = serializer.UInt64ByteSize

// EpochIndex is the index of an epoch.
type EpochIndex uint64

func EpochIndexFromBytes(b []byte) (EpochIndex, int, error) {
	if len(b) < EpochIndexLength {
		return 0, 0, errors.New("invalid epoch index size")
	}

	return EpochIndex(binary.LittleEndian.Uint64(b)), EpochIndexLength, nil
}

func (i EpochIndex) Bytes() ([]byte, error) {
	bytes := make([]byte, serializer.UInt64ByteSize)
	binary.LittleEndian.PutUint64(bytes[:], uint64(i))

	return bytes, nil
}

func (i EpochIndex) MustBytes() []byte {
	return lo.PanicOnErr(i.Bytes())
}

func (i EpochIndex) String() string {
	return fmt.Sprintf("EpochIndex(%d)", i)
}
