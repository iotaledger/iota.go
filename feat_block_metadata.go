package iotago

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// MaxMetadataLength defines the max. length of the data within a MetadataFeatureBlock.
	// TODO: replace with actual value (TBD)
	MaxMetadataLength = MessageBinSerializedMaxSize
)

var (
	// ErrMetadataFeatureBlockEmpty gets returned when a MetadataFeatureBlock is empty.
	ErrMetadataFeatureBlockEmpty = errors.New("metadata feature block is empty")
	// ErrMetadataFeatureBlockDataExceedsMaxLength gets returned when a MetadataFeatureBlock's data exceeds MaxMetadataLength.
	ErrMetadataFeatureBlockDataExceedsMaxLength = errors.New("metadata feature block data exceeds max length")
)

// MetadataFeatureBlock is a feature block which simply holds binary data to be freely
// interpreted by higher layer applications.
type MetadataFeatureBlock struct {
	Data []byte
}

func (s *MetadataFeatureBlock) ValidDataSize() error {
	switch {
	case len(s.Data) == 0:
		return ErrMetadataFeatureBlockEmpty
	case len(s.Data) > MaxMetadataLength:
		return ErrMetadataFeatureBlockDataExceedsMaxLength
	}
	return nil
}

func (s *MetadataFeatureBlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := serializer.CheckTypeByte(data, FeatureBlockMetadata); err != nil {
					return fmt.Errorf("unable to deserialize metadata feature block: %w", err)
				}
			}
			return nil
		}).
		Skip(serializer.SmallTypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip metadata feature block type during deserialization: %w", err)
		}).
		ReadVariableByteSlice(&s.Data, serializer.SeriLengthPrefixTypeAsUint32, func(err error) error {
			return fmt.Errorf("unable to deserialize data for metadata feature block: %w", err)
		}, MaxMetadataLength).
		AbortIf(func(err error) error {
			return s.ValidDataSize()
		}).
		Done()
}

func (s *MetadataFeatureBlock) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				return s.ValidDataSize()
			}
			return nil
		}).
		WriteVariableByteSlice(s.Data, serializer.SeriLengthPrefixTypeAsUint32, func(err error) error {
			return fmt.Errorf("unable to serialize metadata feature block data: %w", err)
		}).
		Serialize()
}

func (s *MetadataFeatureBlock) MarshalJSON() ([]byte, error) {
	jMetadataFeatBlock := &jsonMetadataFeatureBlock{}
	jMetadataFeatBlock.Type = int(FeatureBlockMetadata)
	jMetadataFeatBlock.Data = hex.EncodeToString(s.Data)
	return json.Marshal(jMetadataFeatBlock)
}

func (s *MetadataFeatureBlock) UnmarshalJSON(bytes []byte) error {
	jMetadataFeatBlock := &jsonMetadataFeatureBlock{}
	if err := json.Unmarshal(bytes, jMetadataFeatBlock); err != nil {
		return err
	}
	seri, err := jMetadataFeatBlock.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*MetadataFeatureBlock)
	return nil
}

// jsonMetadataFeatureBlock defines the json representation of a MetadataFeatureBlock.
type jsonMetadataFeatureBlock struct {
	Type int    `json:"type"`
	Data string `json:"data"`
}

func (j *jsonMetadataFeatureBlock) ToSerializable() (serializer.Serializable, error) {
	dataBytes, err := hex.DecodeString(j.Data)
	if err != nil {
		return nil, fmt.Errorf("unable to decode data from JSON for metadata feature block: %w", err)
	}
	return &MetadataFeatureBlock{Data: dataBytes}, nil
}
