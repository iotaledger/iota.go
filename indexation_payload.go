package iota

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	IndexationPayloadID uint32 = 2
	// type bytes + index prefix + one char + data length
	IndexationPayloadMinSize = TypeDenotationByteSize + UInt16ByteSize + OneByte + UInt32ByteSize
)

// IndexationPayload is a payload which holds an index and associated data.
type IndexationPayload struct {
	Index string `json:"index"`
	Data  []byte `json:"data"`
}

func (u *IndexationPayload) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(IndexationPayloadMinSize, len(data)); err != nil {
			return 0, err
		}
		if err := checkType(data, IndexationPayloadID); err != nil {
			return 0, fmt.Errorf("unable to deserialize indexation payload: %w", err)
		}
	}

	data = data[TypeDenotationByteSize:]
	index, indexBytesRead, err := ReadStringFromBytes(data)
	if err != nil {
		return 0, err
	}
	u.Index = index
	data = data[indexBytesRead:]

	// read data length
	dataLength := binary.LittleEndian.Uint32(data)
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check data length
	}

	data = data[ByteArrayLengthByteSize:]
	bytesAvailable := uint32(len(data)) - dataLength
	if bytesAvailable < 0 {
		return 0, fmt.Errorf("%w: indexation payload length denotes too many bytes (%d bytes)", ErrDeserializationNotEnoughData, dataLength)
	}

	u.Data = make([]byte, dataLength)
	copy(u.Data, data[:dataLength])

	return TypeDenotationByteSize + indexBytesRead + ByteArrayLengthByteSize + int(dataLength), nil
}

func (u *IndexationPayload) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check data length
	}

	var b bytes.Buffer
	if err := binary.Write(&b, binary.LittleEndian, IndexationPayloadID); err != nil {
		return nil, err
	}

	strLen := uint16(len(u.Index))
	if err := binary.Write(&b, binary.LittleEndian, strLen); err != nil {
		return nil, err
	}

	if _, err := b.Write([]byte(u.Index)); err != nil {
		return nil, err
	}

	if err := binary.Write(&b, binary.LittleEndian, uint32(len(u.Data))); err != nil {
		return nil, err
	}

	if _, err := b.Write(u.Data); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
