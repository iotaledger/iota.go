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
	// MaxTagLength defines the max. length of an tag feature block tag.
	MaxTagLength = 64
)

var (
	// ErrTagFeatureBlockEmpty gets returned when an TagFeatureBlock is empty.
	ErrTagFeatureBlockEmpty = errors.New("tag feature block data is empty")
	// ErrTagFeatureBlockTagExceedsMaxLength gets returned when an TagFeatureBlock tag exceeds MaxTagLength.
	ErrTagFeatureBlockTagExceedsMaxLength = errors.New("tag feature block tag exceeds max length")
)

// TagFeatureBlock is a feature block which allows to additionally tag an output by a user defined value.
type TagFeatureBlock struct {
	Tag []byte
}

func (s *TagFeatureBlock) Clone() FeatureBlock {
	return &TagFeatureBlock{Tag: append([]byte(nil), s.Tag...)}
}

func (s *TagFeatureBlock) VBytes(costStruct *RentStructure, f VByteCostFunc) uint64 {
	if f != nil {
		return f(costStruct)
	}
	return costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize + serializer.OneByte + uint64(len(s.Tag)))
}

func (s *TagFeatureBlock) Equal(other FeatureBlock) bool {
	otherBlock, is := other.(*TagFeatureBlock)
	if !is {
		return false
	}

	return bytes.Equal(s.Tag, otherBlock.Tag)
}

func (s *TagFeatureBlock) Type() FeatureBlockType {
	return FeatureBlockTag
}

func (s *TagFeatureBlock) ValidTagSize() error {
	switch {
	case len(s.Tag) == 0:
		return ErrTagFeatureBlockEmpty
	case len(s.Tag) > MaxTagLength:
		return ErrTagFeatureBlockTagExceedsMaxLength
	}
	return nil
}

func (s *TagFeatureBlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(FeatureBlockTag), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize tag feature block: %w", err)
		}).
		ReadVariableByteSlice(&s.Tag, serializer.SeriLengthPrefixTypeAsByte, func(err error) error {
			return fmt.Errorf("unable to deserialize tag for tag feature block: %w", err)
		}, MaxTagLength).
		WithValidation(deSeriMode, func(_ []byte, err error) error { return s.ValidTagSize() }).
		Done()
}

func (s *TagFeatureBlock) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WithValidation(deSeriMode, func(_ []byte, err error) error { return s.ValidTagSize() }).
		WriteNum(byte(FeatureBlockTag), func(err error) error {
			return fmt.Errorf("unable to serialize tag feature block type ID: %w", err)
		}).
		WriteVariableByteSlice(s.Tag, serializer.SeriLengthPrefixTypeAsByte, func(err error) error {
			return fmt.Errorf("unable to serialize tag feature block tag: %w", err)
		}).
		Serialize()
}

func (s *TagFeatureBlock) Size() int {
	// tag length prefix = 1 byte
	return util.NumByteLen(byte(FeatureBlockSender)) + serializer.OneByte + len(s.Tag)
}

func (s *TagFeatureBlock) MarshalJSON() ([]byte, error) {
	jTagFeatBlock := &jsonTagFeatureBlock{}
	jTagFeatBlock.Type = int(FeatureBlockTag)
	jTagFeatBlock.Tag = EncodeHex(s.Tag)
	return json.Marshal(jTagFeatBlock)
}

func (s *TagFeatureBlock) UnmarshalJSON(bytes []byte) error {
	jTagFeatBlock := &jsonTagFeatureBlock{}
	if err := json.Unmarshal(bytes, jTagFeatBlock); err != nil {
		return err
	}
	seri, err := jTagFeatBlock.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*TagFeatureBlock)
	return nil
}

// jsonTagFeatureBlock defines the json representation of an TagFeatureBlock.
type jsonTagFeatureBlock struct {
	Type int    `json:"type"`
	Tag  string `json:"tag"`
}

func (j *jsonTagFeatureBlock) ToSerializable() (serializer.Serializable, error) {
	dataBytes, err := DecodeHex(j.Tag)
	if err != nil {
		return nil, fmt.Errorf("unable to decode tag from JSON for tag feature block: %w", err)
	}
	return &TagFeatureBlock{Tag: dataBytes}, nil
}
