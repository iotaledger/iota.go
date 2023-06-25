package iotago

import (
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrWrongEpochIndex gets returned when a wrong epoch index was passed.
	ErrWrongEpochIndex = errors.New("wrong epoch index")
)

// EpochIndex is the index of an epoch.
type EpochIndex uint64

func EpochIndexFromBytes(b []byte) (EpochIndex, error) {
	if len(b) != serializer.UInt64ByteSize {
		return 0, errors.New("invalid epoch index size")
	}

	return EpochIndex(binary.LittleEndian.Uint64(b)), nil
}

func (i EpochIndex) Bytes() []byte {
	bytes := make([]byte, serializer.UInt64ByteSize)
	binary.LittleEndian.PutUint64(bytes[:], uint64(i))

	return bytes
}

func (i EpochIndex) String() string {
	return fmt.Sprintf("EpochIndex(%d)", i)
}
