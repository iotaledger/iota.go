package iotago

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// MaxIndexationTagLength defines the max. length of an indexation tag.
	MaxIndexationTagLength = 64
)

var (
	// ErrIndexationFeatureBlockEmpty gets returned when an IndexationFeatureBlock is empty.
	ErrIndexationFeatureBlockEmpty = errors.New("indexation feature block data is empty")
	// ErrIndexationFeatureBlockTagExceedsMaxLength gets returned when an IndexationFeatureBlock tag exceeds MaxIndexationTagLength.
	ErrIndexationFeatureBlockTagExceedsMaxLength = errors.New("indexation feature block tag exceeds max length")
)

// IndexationFeatureBlock is a feature block which allows to additionally tag an output by a user defined value.
type IndexationFeatureBlock struct {
	Tag []byte
}

func (s *IndexationFeatureBlock) VByteCost(costStruct *RentStructure, f VByteCostFunc) uint64 {
	if f != nil {
		return f(costStruct)
	}
	return costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.OneByte) +
		costStruct.VBFactorKey.With(costStruct.VBFactorData).Multiply(uint64(len(s.Tag)))
}

func (s *IndexationFeatureBlock) Equal(other FeatureBlock) bool {
	otherBlock, is := other.(*IndexationFeatureBlock)
	if !is {
		return false
	}

	return bytes.Equal(s.Tag, otherBlock.Tag)
}

func (s *IndexationFeatureBlock) Type() FeatureBlockType {
	return FeatureBlockIndexation
}

func (s *IndexationFeatureBlock) ValidTagSize() error {
	switch {
	case len(s.Tag) == 0:
		return ErrIndexationFeatureBlockEmpty
	case len(s.Tag) > MaxIndexationTagLength:
		return ErrIndexationFeatureBlockTagExceedsMaxLength
	}
	return nil
}

func (s *IndexationFeatureBlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(FeatureBlockIndexation), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize indexation feature block: %w", err)
		}).
		ReadVariableByteSlice(&s.Tag, serializer.SeriLengthPrefixTypeAsByte, func(err error) error {
			return fmt.Errorf("unable to deserialize tag for indexation feature block: %w", err)
		}, MaxIndexationTagLength).
		WithValidation(deSeriMode, func(_ []byte, err error) error { return s.ValidTagSize() }).
		Done()
}

func (s *IndexationFeatureBlock) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WithValidation(deSeriMode, func(_ []byte, err error) error { return s.ValidTagSize() }).
		WriteNum(byte(FeatureBlockIndexation), func(err error) error {
			return fmt.Errorf("unable to serialize indexation feature block type ID: %w", err)
		}).
		WriteVariableByteSlice(s.Tag, serializer.SeriLengthPrefixTypeAsByte, func(err error) error {
			return fmt.Errorf("unable to serialize indexation feature block data: %w", err)
		}).
		Serialize()
}

func (s *IndexationFeatureBlock) MarshalJSON() ([]byte, error) {
	jIndexationFeatBlock := &jsonIndexationFeatureBlock{}
	jIndexationFeatBlock.Type = int(FeatureBlockIndexation)
	jIndexationFeatBlock.Tag = hex.EncodeToString(s.Tag)
	return json.Marshal(jIndexationFeatBlock)
}

func (s *IndexationFeatureBlock) UnmarshalJSON(bytes []byte) error {
	jIndexationFeatBlock := &jsonIndexationFeatureBlock{}
	if err := json.Unmarshal(bytes, jIndexationFeatBlock); err != nil {
		return err
	}
	seri, err := jIndexationFeatBlock.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*IndexationFeatureBlock)
	return nil
}

// jsonIndexationFeatureBlock defines the json representation of an IndexationFeatureBlock.
type jsonIndexationFeatureBlock struct {
	Type int    `json:"type"`
	Tag  string `json:"tag"`
}

func (j *jsonIndexationFeatureBlock) ToSerializable() (serializer.Serializable, error) {
	dataBytes, err := hex.DecodeString(j.Tag)
	if err != nil {
		return nil, fmt.Errorf("unable to decode tag from JSON for indexation feature block: %w", err)
	}
	return &IndexationFeatureBlock{Tag: dataBytes}, nil
}
