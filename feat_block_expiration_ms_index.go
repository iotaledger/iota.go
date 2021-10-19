package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

// ExpirationMilestoneIndexFeatureBlock is a feature block which puts a time constraint on whether the receiver or sender identity
// can consume an output depending on the latest confirmed milestone index X:
//	- only the receiver can consume the output, if X is smaller than the one defined in the timelock
//	- only the sender can consume the output, if X is bigger-equal than the one defined in the timelock
// As this feature block needs a sender identity for its functionality, this block must have a companion SenderFeatureBlock present in the output.
type ExpirationMilestoneIndexFeatureBlock struct {
	MilestoneIndex uint32
}

func (s *ExpirationMilestoneIndexFeatureBlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := serializer.CheckTypeByte(data, FeatureBlockExpirationMilestoneIndex); err != nil {
					return fmt.Errorf("unable to deserialize expiration milestone index feature block: %w", err)
				}
			}
			return nil
		}).
		Skip(serializer.SmallTypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip expiration milestone index feature block type during deserialization: %w", err)
		}).
		ReadNum(&s.MilestoneIndex, func(err error) error {
			return fmt.Errorf("unable to deserialize milestone index for expiration milestone index feature block: %w", err)
		}).
		Done()
}

func (s *ExpirationMilestoneIndexFeatureBlock) Serialize(_ serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(s.MilestoneIndex, func(err error) error {
			return fmt.Errorf("unable to serialize expiration milestone index feature block milestone index: %w", err)
		}).
		Serialize()
}

func (s *ExpirationMilestoneIndexFeatureBlock) MarshalJSON() ([]byte, error) {
	jExpirationMilestoneFeatBlock := &jsonExpirationMilestoneIndexFeatureBlock{MilestoneIndex: int(s.MilestoneIndex)}
	jExpirationMilestoneFeatBlock.Type = int(FeatureBlockExpirationMilestoneIndex)
	return json.Marshal(jExpirationMilestoneFeatBlock)
}

func (s *ExpirationMilestoneIndexFeatureBlock) UnmarshalJSON(bytes []byte) error {
	jExpirationMilestoneFeatBlock := &jsonExpirationMilestoneIndexFeatureBlock{}
	if err := json.Unmarshal(bytes, jExpirationMilestoneFeatBlock); err != nil {
		return err
	}
	seri, err := jExpirationMilestoneFeatBlock.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*ExpirationMilestoneIndexFeatureBlock)
	return nil
}

// jsonExpirationMilestoneIndexFeatureBlock defines the json representation of a ExpirationMilestoneIndexFeatureBlock.
type jsonExpirationMilestoneIndexFeatureBlock struct {
	Type           int `json:"type"`
	MilestoneIndex int `json:"milestoneIndex"`
}

func (j *jsonExpirationMilestoneIndexFeatureBlock) ToSerializable() (serializer.Serializable, error) {
	return &ExpirationMilestoneIndexFeatureBlock{MilestoneIndex: uint32(j.MilestoneIndex)}, nil
}
