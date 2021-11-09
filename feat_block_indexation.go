package iotago

import (
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

func (s *IndexationFeatureBlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(FeatureBlockIndexation), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize indexation feature block: %w", err)
		}).
		ReadVariableByteSlice(&s.Tag, serializer.SeriLengthPrefixTypeAsUint32, func(err error) error {
			return fmt.Errorf("unable to deserialize tag for indexation feature block: %w", err)
		}, MaxIndexationTagLength).
		AbortIf(func(err error) error { return s.ValidTagSize() }).
		Done()
}

func (s *IndexationFeatureBlock) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				return s.ValidTagSize()
			}
			return nil
		}).
		WriteNum(byte(FeatureBlockIndexation), func(err error) error {
			return fmt.Errorf("unable to serialize indexation feature block type ID: %w", err)
		}).
		WriteVariableByteSlice(s.Tag, serializer.SeriLengthPrefixTypeAsUint32, func(err error) error {
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
