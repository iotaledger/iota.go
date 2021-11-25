package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// DustDepositReturnFeatureBlock is a feature block which defines
// the amount of tokens which must be sent back to the sender identity, when the output in which it occurs in,
// is consumed. This block must have a companion SenderFeatureBlock occurring in the same output
// from which the sender identity can be extracted from.
// If a transaction consumes multiple outputs which have a DustDepositReturnFeatureBlock, then on the output side at least
// the sum of all occurring DustDepositReturnFeatureBlock(s) on the input side must be deposited to the designated origin sender.
type DustDepositReturnFeatureBlock struct {
	Amount uint64
}

func (s *DustDepositReturnFeatureBlock) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize + serializer.UInt64ByteSize)
}

func (s *DustDepositReturnFeatureBlock) Equal(other FeatureBlock) bool {
	otherBlock, is := other.(*DustDepositReturnFeatureBlock)
	if !is {
		return false
	}

	return s.Amount == otherBlock.Amount
}

func (s *DustDepositReturnFeatureBlock) Type() FeatureBlockType {
	return FeatureBlockDustDepositReturn
}

func (s *DustDepositReturnFeatureBlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(FeatureBlockDustDepositReturn), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize return feature block: %w", err)
		}).
		ReadNum(&s.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for return feature block: %w", err)
		}).
		Done()
}

func (s *DustDepositReturnFeatureBlock) Serialize(_ serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(FeatureBlockDustDepositReturn), func(err error) error {
			return fmt.Errorf("unable to serialize return feature block type ID: %w", err)
		}).
		WriteNum(s.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize return feature block amount: %w", err)
		}).
		Serialize()
}

func (s *DustDepositReturnFeatureBlock) MarshalJSON() ([]byte, error) {
	jReturnFeatBlock := &jsonReturnFeatureBlock{Amount: int(s.Amount)}
	jReturnFeatBlock.Type = int(FeatureBlockDustDepositReturn)
	return json.Marshal(jReturnFeatBlock)
}

func (s *DustDepositReturnFeatureBlock) UnmarshalJSON(bytes []byte) error {
	jReturnFeatBlock := &jsonReturnFeatureBlock{}
	if err := json.Unmarshal(bytes, jReturnFeatBlock); err != nil {
		return err
	}
	seri, err := jReturnFeatBlock.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*DustDepositReturnFeatureBlock)
	return nil
}

// jsonReturnFeatureBlock defines the json representation of a DustDepositReturnFeatureBlock.
type jsonReturnFeatureBlock struct {
	Type   int `json:"type"`
	Amount int `json:"amount"`
}

func (j *jsonReturnFeatureBlock) ToSerializable() (serializer.Serializable, error) {
	return &DustDepositReturnFeatureBlock{Amount: uint64(j.Amount)}, nil
}
