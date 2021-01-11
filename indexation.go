package iota

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"unicode/utf8"
)

const (
	// Defines the indexation payload's ID.
	IndexationPayloadTypeID uint32 = 2
	// type bytes + index prefix + one char + data length
	IndexationBinSerializedMinSize = TypeDenotationByteSize + UInt16ByteSize + OneByte + UInt32ByteSize
	// Defines the max length of the index within an Indexation.
	IndexationIndexMaxLength = 64
)

var (
	// Returned when an Indexation contains a non UTF-8 index.
	ErrIndexationNonUTF8Index = errors.New("index is not valid utf-8")
	// Returned when an Indexation's index exceeds IndexationIndexMaxLength.
	ErrIndexationIndexExceedsMaxSize = errors.New("index exceeds max size")
)

// Indexation is a payload which holds an index and associated data.
type Indexation struct {
	// The index to use to index the enclosing message and data.
	Index string `json:"index"`
	// The data within the payload.
	Data []byte `json:"data"`
}

func (u *Indexation) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	return NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := checkMinByteLength(IndexationBinSerializedMinSize, len(data)); err != nil {
					return fmt.Errorf("invalid indexation bytes: %w", err)
				}
				if err := checkType(data, IndexationPayloadTypeID); err != nil {
					return fmt.Errorf("unable to deserialize indexation: %w", err)
				}
				// TODO: check data length
			}
			return nil
		}).
		Skip(TypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip indexation payload ID during deserialization: %w", err)
		}).
		ReadString(&u.Index, func(err error) error {
			return fmt.Errorf("unable to deserialize indexation index: %w", err)
		}, IndexationIndexMaxLength).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if !utf8.ValidString(u.Index) {
					return fmt.Errorf("unable to deserialize indexation index: %w", ErrIndexationNonUTF8Index)
				}
			}
			return nil
		}).
		ReadVariableByteSlice(&u.Data, SeriSliceLengthAsUint32, func(err error) error {
			return fmt.Errorf("unable to deserialize indexation data: %w", err)
		}).
		Done()
}

func (u *Indexation) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	return NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if len(u.Index) > IndexationIndexMaxLength {
					return fmt.Errorf("unable to deserialize indexation index: %w", ErrIndexationIndexExceedsMaxSize)
				}
				if !utf8.ValidString(u.Index) {
					return fmt.Errorf("unable to deserialize indexation index: %w", ErrIndexationNonUTF8Index)
				}
				// TODO: check data length
			}
			return nil
		}).
		WriteNum(IndexationPayloadTypeID, func(err error) error {
			return fmt.Errorf("unable to serialize indexation payload ID: %w", err)
		}).
		WriteString(u.Index, func(err error) error {
			return fmt.Errorf("unable to serialize indexation index: %w", err)
		}).
		WriteVariableByteSlice(u.Data, SeriSliceLengthAsUint32, func(err error) error {
			return fmt.Errorf("unable to serialize indexation data: %w", err)
		}).
		Serialize()
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
