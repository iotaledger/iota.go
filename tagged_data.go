package iotago

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// TaggedPayloadTagMaxLength defines the max length of the tag within a TaggedData payload.
	TaggedPayloadTagMaxLength = 64
	// TaggedPayloadTagMinLength defines the min length of the tag within a TaggedData payload.
	TaggedPayloadTagMinLength = 1
)

var (
	// ErrTaggedDataTagExceedsMaxSize gets returned when a TaggedData payload's tag exceeds TaggedPayloadTagMaxLength.
	ErrTaggedDataTagExceedsMaxSize = errors.New("tag exceeds max size")
	// ErrTaggedDataTagUnderMinSize gets returned when an TaggedData payload's tag is under TaggedPayloadTagMinLength.
	ErrTaggedDataTagUnderMinSize = errors.New("tag is below min size")
)

// TaggedData is a payload which holds a tag and associated data.
type TaggedData struct {
	// The tag to use to categorize the data.
	Tag []byte
	// The data within the payload.
	Data []byte
}

func (u *TaggedData) PayloadType() PayloadType {
	return PayloadTaggedData
}

func (u *TaggedData) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(PayloadTaggedData), serializer.TypeDenotationUint32, func(err error) error {
			return fmt.Errorf("unable to deserialize tagged data: %w", err)
		}).
		ReadVariableByteSlice(&u.Tag, serializer.SeriLengthPrefixTypeAsUint16, func(err error) error {
			return fmt.Errorf("unable to deserialize tagged data tag: %w", err)
		}, TaggedPayloadTagMaxLength).
		WithValidation(deSeriMode, func(_ []byte, err error) error {
			switch {
			case len(u.Tag) < TaggedPayloadTagMinLength:
				return fmt.Errorf("unable to deserialize tagged data tag: %w", ErrTaggedDataTagUnderMinSize)
			}
			return nil
		}).
		ReadVariableByteSlice(&u.Data, serializer.SeriLengthPrefixTypeAsUint32, func(err error) error {
			return fmt.Errorf("unable to deserialize tagged data data: %w", err)
		}, MessageBinSerializedMaxSize). // obviously can never be that size
		Done()
}

func (u *TaggedData) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WithValidation(deSeriMode, func(_ []byte, err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				switch {
				case len(u.Tag) > TaggedPayloadTagMaxLength:
					return fmt.Errorf("unable to serialize tagged data tag: %w", ErrTaggedDataTagExceedsMaxSize)
				case len(u.Tag) < TaggedPayloadTagMinLength:
					return fmt.Errorf("unable to serialize tagged data tag: %w", ErrTaggedDataTagUnderMinSize)
				}
				// we do not check the length of the data field as in any circumstance
				// the max size it can take up is dependent on how big the enclosing
				// parent object is
			}
			return nil
		}).
		WriteNum(PayloadTaggedData, func(err error) error {
			return fmt.Errorf("unable to serialize tagged data payload ID: %w", err)
		}).
		WriteVariableByteSlice(u.Tag, serializer.SeriLengthPrefixTypeAsUint16, func(err error) error {
			return fmt.Errorf("unable to serialize tagged data tag: %w", err)
		}).
		WriteVariableByteSlice(u.Data, serializer.SeriLengthPrefixTypeAsUint32, func(err error) error {
			return fmt.Errorf("unable to serialize tagged data data: %w", err)
		}).
		Serialize()
}

func (u *TaggedData) MarshalJSON() ([]byte, error) {
	jTaggedData := &jsonTaggedData{}
	jTaggedData.Type = int(PayloadTaggedData)
	jTaggedData.Tag = hex.EncodeToString(u.Tag)
	jTaggedData.Data = hex.EncodeToString(u.Data)
	return json.Marshal(jTaggedData)
}

func (u *TaggedData) UnmarshalJSON(bytes []byte) error {
	jTaggedData := &jsonTaggedData{}
	if err := json.Unmarshal(bytes, jTaggedData); err != nil {
		return err
	}
	seri, err := jTaggedData.ToSerializable()
	if err != nil {
		return err
	}
	*u = *seri.(*TaggedData)
	return nil
}

// jsonTaggedData defines the json representation of a TaggedData payload.
type jsonTaggedData struct {
	Type int    `json:"type"`
	Tag  string `json:"tag"`
	Data string `json:"data"`
}

func (j *jsonTaggedData) ToSerializable() (serializer.Serializable, error) {
	tagBytes, err := hex.DecodeString(j.Tag)
	if err != nil {
		return nil, fmt.Errorf("unable to decode tag from JSON for tagged data payload: %w", err)
	}

	dataBytes, err := hex.DecodeString(j.Data)
	if err != nil {
		return nil, fmt.Errorf("unable to decode data from JSON for tagged data payload: %w", err)
	}

	return &TaggedData{Tag: tagBytes, Data: dataBytes}, nil
}
