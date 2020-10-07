package iota

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var (
	// Returned for bytes which are invalid for deserialization.
	ErrInvalidBytes = errors.New("invalid bytes")
	// Returned when a denoted type for a given object is mismatched.
	// For example, while trying to deserialize a signature unlock block, a reference unlock block is seen.
	ErrDeserializationTypeMismatch = errors.New("data type is invalid for deserialization")
	// Returned for unknown payload types.
	ErrUnknownPayloadType = errors.New("unknown payload type")
	// Returned for unknown address types.
	ErrUnknownAddrType = errors.New("unknown address type")
	// Returned for unknown input types.
	ErrUnknownInputType = errors.New("unknown input type")
	// Returned for unknown output types.
	ErrUnknownOutputType = errors.New("unknown output type")
	// Returned for unknown transaction essence types.
	ErrUnknownTransactionEssenceType = errors.New("unknown transaction essence type")
	// Returned for unknown unlock blocks.
	ErrUnknownUnlockBlockType = errors.New("unknown unlock block type")
	// Returned for unknown signature types.
	ErrUnknownSignatureType = errors.New("unknown signature type")
	// Returned if there is not enough data available to deserialize a given object.
	ErrDeserializationNotEnoughData = errors.New("not enough data for deserialization")
	// Returned if not all bytes were consumed during deserialization of a given type.
	ErrDeserializationNotAllConsumed = errors.New("not all data has been consumed but should have been")
	// Returned when WOTS objects are tried to be de/serialized.
	ErrWOTSNotImplemented = errors.New("unfortunately WOTS is not yet implemented")
)

// checkType checks that the denoted type equals the shouldType.
func checkType(data []byte, shouldType uint32) error {
	actualType := binary.LittleEndian.Uint32(data)
	if actualType != shouldType {
		return fmt.Errorf("%w: type denotation must be %d but is %d", ErrDeserializationTypeMismatch, shouldType, actualType)
	}
	return nil
}

// checkTypeByte checks that the denoted type byte equals the shouldType.
func checkTypeByte(data []byte, shouldType byte) error {
	if data == nil || len(data) == 0 {
		return fmt.Errorf("%w: can't check type byte", ErrDeserializationNotEnoughData)
	}
	if data[0] != shouldType {
		return fmt.Errorf("%w: type denotation must be %d but is %d", ErrDeserializationTypeMismatch, shouldType, data[0])
	}
	return nil
}

// checkExactByteLength checks that the given length equals exact.
func checkExactByteLength(exact int, length int) error {
	if length != exact {
		return fmt.Errorf("%w: data must be at exact %d bytes long but is %d", ErrInvalidBytes, exact, length)
	}
	return nil
}

// checkByteLengthRange checks that length is within min and max.
func checkByteLengthRange(min int, max int, length int) error {
	if err := checkMinByteLength(min, length); err != nil {
		return err
	}
	if err := checkMaxByteLength(max, length); err != nil {
		return err
	}
	return nil
}

// checkMinByteLength checks that length is min. min.
func checkMinByteLength(min int, length int) error {
	if length < min {
		return fmt.Errorf("%w: data must be at least %d bytes long but is %d", ErrDeserializationNotEnoughData, min, length)
	}
	return nil
}

// checkMaxByteLength checks that length is max. max.
func checkMaxByteLength(max int, length int) error {
	if length > max {
		return fmt.Errorf("%w: data must be max %d bytes long but is %d", ErrInvalidBytes, max, length)
	}
	return nil
}
