package iotago

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

const (
	// MaxMetadataLength defines the max length of the data within a MetadataFeature.
	MaxMetadataLength = 8192
)

var (
	// ErrMetadataFeatureEmpty gets returned when a MetadataFeature is empty.
	ErrMetadataFeatureEmpty = errors.New("metadata feature is empty")
	// ErrMetadataFeatureDataExceedsMaxLength gets returned when a MetadataFeature's data exceeds MaxMetadataLength.
	ErrMetadataFeatureDataExceedsMaxLength = errors.New("metadata feature data exceeds max length")
)

// MetadataFeature is a feature which simply holds binary data to be freely
// interpreted by higher layer applications.
type MetadataFeature struct {
	Data []byte
}

func (s *MetadataFeature) Clone() Feature {
	return &MetadataFeature{Data: append([]byte(nil), s.Data...)}
}

func (s *MetadataFeature) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(uint64(serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize + len(s.Data)))
}

func (s *MetadataFeature) Equal(other Feature) bool {
	otherFeat, is := other.(*MetadataFeature)
	if !is {
		return false
	}

	return bytes.Equal(s.Data, otherFeat.Data)
}

func (s *MetadataFeature) Type() FeatureType {
	return FeatureMetadata
}

func (s *MetadataFeature) ValidDataSize() error {
	switch {
	case len(s.Data) == 0:
		return ErrMetadataFeatureEmpty
	case len(s.Data) > MaxMetadataLength:
		return ErrMetadataFeatureDataExceedsMaxLength
	}
	return nil
}

func (s *MetadataFeature) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(FeatureMetadata), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize metadata feature: %w", err)
		}).
		ReadVariableByteSlice(&s.Data, serializer.SeriLengthPrefixTypeAsUint16, func(err error) error {
			return fmt.Errorf("unable to deserialize data for metadata feature: %w", err)
		}, MaxMetadataLength).
		WithValidation(deSeriMode, func(_ []byte, err error) error { return s.ValidDataSize() }).
		Done()
}

func (s *MetadataFeature) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WithValidation(deSeriMode, func(_ []byte, err error) error { return s.ValidDataSize() }).
		WriteNum(byte(FeatureMetadata), func(err error) error {
			return fmt.Errorf("unable to serialize metadata feature type ID: %w", err)
		}).
		WriteVariableByteSlice(s.Data, serializer.SeriLengthPrefixTypeAsUint16, func(err error) error {
			return fmt.Errorf("unable to serialize metadata feature data: %w", err)
		}).
		Serialize()
}

func (s *MetadataFeature) Size() int {
	// data length prefix as uint16 = 2 bytes
	return util.NumByteLen(byte(FeatureMetadata)) + serializer.UInt16ByteSize + len(s.Data)
}

func (s *MetadataFeature) MarshalJSON() ([]byte, error) {
	jMetadataFeat := &jsonMetadataFeature{}
	jMetadataFeat.Type = int(FeatureMetadata)
	jMetadataFeat.Data = EncodeHex(s.Data)
	return json.Marshal(jMetadataFeat)
}

func (s *MetadataFeature) UnmarshalJSON(bytes []byte) error {
	jMetadataFeat := &jsonMetadataFeature{}
	if err := json.Unmarshal(bytes, jMetadataFeat); err != nil {
		return err
	}
	seri, err := jMetadataFeat.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*MetadataFeature)
	return nil
}

// jsonMetadataFeature defines the json representation of a MetadataFeature.
type jsonMetadataFeature struct {
	Type int    `json:"type"`
	Data string `json:"data"`
}

func (j *jsonMetadataFeature) ToSerializable() (serializer.Serializable, error) {
	dataBytes, err := DecodeHex(j.Data)
	if err != nil {
		return nil, fmt.Errorf("unable to decode data from JSON for metadata feature: %w", err)
	}
	return &MetadataFeature{Data: dataBytes}, nil
}
