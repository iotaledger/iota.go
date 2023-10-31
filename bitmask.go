package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

var (
	// ErrBitmaskTrailingZeroBytes gets returned when the trailing bytes of a bitmask are zero.
	ErrBitmaskTrailingZeroBytes = ierrors.New("bitmask trailing bytes are zero")
)

func BitMaskHasBit(bm []byte, bit uint) bool {
	byteIndex := bit / 8
	if uint(len(bm)) <= byteIndex {
		return false
	}
	bitIndex := bit % 8

	return bm[byteIndex]&(1<<bitIndex) > 0
}

func BitMaskSetBit(bm []byte, bit uint) []byte {
	newBitmask := bm
	byteIndex := bit / 8
	for uint(len(newBitmask)) <= byteIndex {
		newBitmask = append(newBitmask, 0)
	}
	bitIndex := bit % 8
	newBitmask[byteIndex] |= 1 << bitIndex

	return newBitmask
}

// BitMaskNonTrailingZeroBytesValidatorFunc checks that the trailing bytes of the bitmask are not zero.
func BitMaskNonTrailingZeroBytesValidatorFunc(bm []byte) error {
	if len(bm) == 0 || bm[len(bm)-1] != 0 {
		return nil
	}

	return ierrors.Wrapf(ErrBitmaskTrailingZeroBytes, "bitmask: %s", hexutil.EncodeHex(bm))
}
