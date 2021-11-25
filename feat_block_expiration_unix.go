package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// ExpirationUnixFeatureBlock is a feature block which puts a time constraint on whether the receiver or sender identity
// can consume an output depending on the latest confirmed milestone's timestamp T:
//	- only the receiver can consume the output, if T is before than the one defined in the timelock
//	- only the sender can consume the output, if T is at the same time or after the one defined in the timelock
// As this feature block needs a sender identity for its functionality, this block must have a companion SenderFeatureBlock present in the output.
type ExpirationUnixFeatureBlock struct {
	// UnixTime is the second resolution unix time.
	UnixTime uint64
}

func (s *ExpirationUnixFeatureBlock) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize + serializer.UInt64ByteSize)
}

func (s *ExpirationUnixFeatureBlock) Equal(other FeatureBlock) bool {
	otherBlock, is := other.(*ExpirationUnixFeatureBlock)
	if !is {
		return false
	}

	return s.UnixTime == otherBlock.UnixTime
}

func (s *ExpirationUnixFeatureBlock) Type() FeatureBlockType {
	return FeatureBlockExpirationUnix
}

func (s *ExpirationUnixFeatureBlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(FeatureBlockExpirationUnix), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize expiration unix feature block: %w", err)
		}).
		ReadNum(&s.UnixTime, func(err error) error {
			return fmt.Errorf("unable to deserialize unix time for expiration unix feature block: %w", err)
		}).
		Done()
}

func (s *ExpirationUnixFeatureBlock) Serialize(_ serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(FeatureBlockExpirationUnix), func(err error) error {
			return fmt.Errorf("unable to serialize expiration unix feature block type ID: %w", err)
		}).
		WriteNum(s.UnixTime, func(err error) error {
			return fmt.Errorf("unable to serialize expiration unix feature block unix time: %w", err)
		}).
		Serialize()
}

func (s *ExpirationUnixFeatureBlock) MarshalJSON() ([]byte, error) {
	jExpirationUnixFeatBlock := &jsonExpirationUnixFeatureBlock{UnixTime: int(s.UnixTime)}
	jExpirationUnixFeatBlock.Type = int(FeatureBlockExpirationUnix)
	return json.Marshal(jExpirationUnixFeatBlock)
}

func (s *ExpirationUnixFeatureBlock) UnmarshalJSON(bytes []byte) error {
	jExpirationMilestoneFeatBlock := &jsonExpirationUnixFeatureBlock{}
	if err := json.Unmarshal(bytes, jExpirationMilestoneFeatBlock); err != nil {
		return err
	}
	seri, err := jExpirationMilestoneFeatBlock.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*ExpirationUnixFeatureBlock)
	return nil
}

// jsonExpirationUnixFeatureBlock defines the json representation of a ExpirationUnixFeatureBlock.
type jsonExpirationUnixFeatureBlock struct {
	Type     int `json:"type"`
	UnixTime int `json:"unixTime"`
}

func (j *jsonExpirationUnixFeatureBlock) ToSerializable() (serializer.Serializable, error) {
	return &ExpirationUnixFeatureBlock{UnixTime: uint64(j.UnixTime)}, nil
}
