package util

import (
	"math/big"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// NumByteLen returns the number of bytes a given number will occupy when serialized.
func NumByteLen(n interface{}) int {
	switch n.(type) {
	case bool, int8, uint8, *bool, *int8, *uint8:
		return serializer.OneByte
	case int16, uint16, *int16, *uint16:
		return serializer.UInt16ByteSize
	case int32, uint32, *int32, *uint32:
		return serializer.UInt32ByteSize
	case int64, uint64, *int64, *uint64:
		return serializer.UInt64ByteSize
	case float32, *float32:
		return serializer.Float32ByteSize
	case float64, *float64:
		return serializer.Float64ByteSize
	case *big.Int, big.Int:
		return serializer.UInt256ByteSize
	}
	panic("NumByteLen type not supported")
}
