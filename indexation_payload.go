package iota

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	// Defines the indexation payload's ID.
	IndexationPayloadID uint32 = 2
	// type bytes + index prefix + one char + data length
	IndexationPayloadMinSize = TypeDenotationByteSize + UInt16ByteSize + OneByte + UInt32ByteSize
)

// IndexationPayload is a payload which holds an index and associated data.
type IndexationPayload struct {
	// The index to use to index the enclosing message and data.
	Index string `json:"index"`
	// The data within the payload.
	Data []byte `json:"data"`
}

func (u *IndexationPayload) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(IndexationPayloadMinSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid indexation payload bytes: %w", err)
		}
		if err := checkType(data, IndexationPayloadID); err != nil {
			return 0, fmt.Errorf("unable to deserialize indexation payload: %w", err)
		}
	}

	data = data[TypeDenotationByteSize:]
	index, indexBytesRead, err := ReadStringFromBytes(data)
	if err != nil {
		return 0, fmt.Errorf("%w: unable to deserialize indexation payload index", err)
	}
	u.Index = index
	data = data[indexBytesRead:]

	if len(data) < UInt32ByteSize {
		return 0, fmt.Errorf("%w: unable to deserialize indexation payload data length", ErrDeserializationNotEnoughData)
	}

	// read data length
	dataLength := binary.LittleEndian.Uint32(data)
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check data length
	}

	data = data[ByteArrayLengthByteSize:]
	if uint32(len(data)) < dataLength {
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
		return nil, fmt.Errorf("%w: unable to serialize indexation payload ID", err)
	}

	strLen := uint16(len(u.Index))
	if err := binary.Write(&b, binary.LittleEndian, strLen); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize indexation payload index length", err)
	}

	if _, err := b.Write([]byte(u.Index)); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize indexation payload index", err)
	}

	if err := binary.Write(&b, binary.LittleEndian, uint32(len(u.Data))); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize indexation payload data length", err)
	}

	if _, err := b.Write(u.Data); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize indexation payload data", err)
	}

	return b.Bytes(), nil
}

func (u *IndexationPayload) MarshalJSON() ([]byte, error) {
	jsonIndexPayload := &jsonindexationpayload{}
	jsonIndexPayload.Type = int(IndexationPayloadID)
	jsonIndexPayload.Index = u.Index
	jsonIndexPayload.Data = hex.EncodeToString(u.Data)
	return json.Marshal(jsonIndexPayload)
}

func (u *IndexationPayload) UnmarshalJSON(bytes []byte) error {
	jsonIndexPayload := &jsonindexationpayload{}
	if err := json.Unmarshal(bytes, jsonIndexPayload); err != nil {
		return err
	}
	seri, err := jsonIndexPayload.ToSerializable()
	if err != nil {
		return err
	}
	*u = *seri.(*IndexationPayload)
	return nil
}

// jsonindexationpayload defines the json representation of an IndexationPayload.
type jsonindexationpayload struct {
	Type  int    `json:"type"`
	Index string `json:"index"`
	Data  string `json:"data"`
}

func (j *jsonindexationpayload) ToSerializable() (Serializable, error) {
	dataBytes, err := hex.DecodeString(j.Data)
	if err != nil {
		return nil, fmt.Errorf("unable to decode data from JSON for indexation payload: %w", err)
	}

	payload := &IndexationPayload{Index: j.Index}
	payload.Data = dataBytes
	return payload, nil
}
