package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

// ReturnFeatureBlock is a feature block which defines
// the amount of tokens which must be sent back to the sender identity, when the output in which it occurs in,
// is consumed by the receiver. This block must have a companion SenderFeatureBlock occurring in the same output
// from which the sender identity can be extracted from.
// If a transaction consumes multiple outputs which have a ReturnFeatureBlock, then on the output side at least
// the sum of all occurring ReturnFeatureBlock(s) on the input side must be deposited to the designated origin sender.
type ReturnFeatureBlock struct {
	Amount uint64
}

func (s *ReturnFeatureBlock) Equal(other FeatureBlock) bool {
	otherBlock, is := other.(*ReturnFeatureBlock)
	if !is {
		return false
	}

	return s.Amount == otherBlock.Amount
}

func (s *ReturnFeatureBlock) Type() FeatureBlockType {
	return FeatureBlockReturn
}

func (s *ReturnFeatureBlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(FeatureBlockReturn), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize return feature block: %w", err)
		}).
		ReadNum(&s.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for return feature block: %w", err)
		}).
		Done()
}

func (s *ReturnFeatureBlock) Serialize(_ serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(FeatureBlockReturn), func(err error) error {
			return fmt.Errorf("unable to serialize return feature block type ID: %w", err)
		}).
		WriteNum(s.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize return feature block amount: %w", err)
		}).
		Serialize()
}

func (s *ReturnFeatureBlock) MarshalJSON() ([]byte, error) {
	jReturnFeatBlock := &jsonReturnFeatureBlock{Amount: int(s.Amount)}
	jReturnFeatBlock.Type = int(FeatureBlockReturn)
	return json.Marshal(jReturnFeatBlock)
}

func (s *ReturnFeatureBlock) UnmarshalJSON(bytes []byte) error {
	jReturnFeatBlock := &jsonReturnFeatureBlock{}
	if err := json.Unmarshal(bytes, jReturnFeatBlock); err != nil {
		return err
	}
	seri, err := jReturnFeatBlock.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*ReturnFeatureBlock)
	return nil
}

// jsonReturnFeatureBlock defines the json representation of a ReturnFeatureBlock.
type jsonReturnFeatureBlock struct {
	Type   int `json:"type"`
	Amount int `json:"amount"`
}

func (j *jsonReturnFeatureBlock) ToSerializable() (serializer.Serializable, error) {
	return &ReturnFeatureBlock{Amount: uint64(j.Amount)}, nil
}
