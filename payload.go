package iota

import (
	"encoding/binary"
	"fmt"
)

// ParsePayload parses a payload out of the given data.
// It returns the amount of bytes read from data. If the payload length is 0, then
// the returned Serializable is nil.
func ParsePayload(data []byte, deSeriMode DeSerializationMode) (Serializable, int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if len(data) < PayloadLengthByteSize {
			return nil, 0, fmt.Errorf("%w: data is smaller than payload length denotation", ErrDeserializationNotEnoughData)
		}
	}

	// read length
	payloadLength := binary.LittleEndian.Uint32(data)
	data = data[PayloadLengthByteSize:]

	if payloadLength == 0 {
		return nil, PayloadLengthByteSize, nil
	}

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check max payload length

		if len(data) < MinPayloadByteSize {
			return nil, 0, fmt.Errorf("%w: payload data is smaller than min. required length %d", ErrDeserializationNotEnoughData, MinPayloadByteSize)
		}

		if len(data) < int(payloadLength) {
			return nil, 0, fmt.Errorf("%w: payload length denotes more bytes than are available", ErrDeserializationNotEnoughData)
		}
	}

	payload, err := PayloadSelector(binary.LittleEndian.Uint32(data))
	if err != nil {
		return nil, 0, err
	}

	payloadBytesConsumed, err := payload.Deserialize(data, deSeriMode)
	if err != nil {
		return nil, 0, err
	}

	if payloadBytesConsumed != int(payloadLength) {
		return nil, 0, fmt.Errorf("%w: denoted payload length (%d) doesn't equal the size of deserialized payload (%d)", ErrInvalidBytes, payloadLength, payloadBytesConsumed)
	}

	return payload, UInt32ByteSize + payloadBytesConsumed, nil
}
