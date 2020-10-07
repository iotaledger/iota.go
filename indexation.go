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
	IndexationPayloadTypeID uint32 = 2
	// type bytes + index prefix + one char + data length
	IndexationBinSerializedMinSize = TypeDenotationByteSize + UInt16ByteSize + OneByte + UInt32ByteSize
)

// Indexation is a payload which holds an index and associated data.
type Indexation struct {
	// The index to use to index the enclosing message and data.
	Index string `json:"index"`
	// The data within the payload.
	Data []byte `json:"data"`
}

func (u *Indexation) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(IndexationBinSerializedMinSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid indexation bytes: %w", err)
		}
		if err := checkType(data, IndexationPayloadTypeID); err != nil {
			return 0, fmt.Errorf("unable to deserialize indexation: %w", err)
		}
	}

	data = data[TypeDenotationByteSize:]
	index, indexBytesRead, err := ReadStringFromBytes(data)
	if err != nil {
		return 0, fmt.Errorf("%w: unable to deserialize indexation index", err)
	}
	u.Index = index
	data = data[indexBytesRead:]

	if len(data) < UInt32ByteSize {
		return 0, fmt.Errorf("%w: unable to deserialize indexation data length", ErrDeserializationNotEnoughData)
	}

	// read data length
	dataLength := binary.LittleEndian.Uint32(data)
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check data length
	}

	data = data[ByteArrayLengthByteSize:]
	if uint32(len(data)) < dataLength {
		return 0, fmt.Errorf("%w: indexation length denotes too many bytes (%d bytes)", ErrDeserializationNotEnoughData, dataLength)
	}

	u.Data = make([]byte, dataLength)
	copy(u.Data, data[:dataLength])

	return TypeDenotationByteSize + indexBytesRead + ByteArrayLengthByteSize + int(dataLength), nil
}

func (u *Indexation) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check data length
	}

	var b bytes.Buffer
	if err := binary.Write(&b, binary.LittleEndian, IndexationPayloadTypeID); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize indexation payload ID", err)
	}

	strLen := uint16(len(u.Index))
	if err := binary.Write(&b, binary.LittleEndian, strLen); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize indexation index length", err)
	}

	if _, err := b.Write([]byte(u.Index)); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize indexation index", err)
	}

	if err := binary.Write(&b, binary.LittleEndian, uint32(len(u.Data))); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize indexation data length", err)
	}

	if _, err := b.Write(u.Data); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize indexation data", err)
	}

	return b.Bytes(), nil
}

func (u *Indexation) MarshalJSON() ([]byte, error) {
	jsonIndexPayload := &jsonindexation{}
	jsonIndexPayload.Type = int(IndexationPayloadTypeID)
	jsonIndexPayload.Index = u.Index
	jsonIndexPayload.Data = hex.EncodeToString(u.Data)
	return json.Marshal(jsonIndexPayload)
}

func (u *Indexation) UnmarshalJSON(bytes []byte) error {
	jsonIndexPayload := &jsonindexation{}
	if err := json.Unmarshal(bytes, jsonIndexPayload); err != nil {
		return err
	}
	seri, err := jsonIndexPayload.ToSerializable()
	if err != nil {
		return err
	}
	*u = *seri.(*Indexation)
	return nil
}

// jsonindexation defines the json representation of an Indexation.
type jsonindexation struct {
	Type  int    `json:"type"`
	Index string `json:"index"`
	Data  string `json:"data"`
}

func (j *jsonindexation) ToSerializable() (Serializable, error) {
	dataBytes, err := hex.DecodeString(j.Data)
	if err != nil {
		return nil, fmt.Errorf("unable to decode data from JSON for indexation: %w", err)
	}

	payload := &Indexation{Index: j.Index}
	payload.Data = dataBytes
	return payload, nil
}
