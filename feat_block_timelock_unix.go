package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

// TimelockUnixFeatureBlock is a feature block which puts a time constraint on an output depending
// on the latest confirmed milestone's timestamp T:
//	- the output can only be consumed, if T is after the one defined in the timelock
type TimelockUnixFeatureBlock struct {
	// UnixTime is the second resolution unix time.
	UnixTime uint64
}

func (s *TimelockUnixFeatureBlock) Equal(other FeatureBlock) bool {
	otherBlock, is := other.(*TimelockUnixFeatureBlock)
	if !is {
		return false
	}

	return s.UnixTime == otherBlock.UnixTime
}

func (s *TimelockUnixFeatureBlock) Type() FeatureBlockType {
	return FeatureBlockTimelockUnix
}

func (s *TimelockUnixFeatureBlock) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(FeatureBlockTimelockUnix), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize timelock unix feature block: %w", err)
		}).
		ReadNum(&s.UnixTime, func(err error) error {
			return fmt.Errorf("unable to deserialize unix time for timelock unix feature block: %w", err)
		}).
		Done()
}

func (s *TimelockUnixFeatureBlock) Serialize(_ serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(FeatureBlockTimelockUnix), func(err error) error {
			return fmt.Errorf("unable to serialize timelock unix feature block type ID: %w", err)
		}).
		WriteNum(s.UnixTime, func(err error) error {
			return fmt.Errorf("unable to serialize timelock unix feature block unix time: %w", err)
		}).
		Serialize()
}

func (s *TimelockUnixFeatureBlock) MarshalJSON() ([]byte, error) {
	jTimelockUnixFeatBlock := &jsonTimelockUnixFeatureBlock{UnixTime: int(s.UnixTime)}
	jTimelockUnixFeatBlock.Type = int(FeatureBlockTimelockUnix)
	return json.Marshal(jTimelockUnixFeatBlock)
}

func (s *TimelockUnixFeatureBlock) UnmarshalJSON(bytes []byte) error {
	jTimelockMilestoneFeatBlock := &jsonTimelockUnixFeatureBlock{}
	if err := json.Unmarshal(bytes, jTimelockMilestoneFeatBlock); err != nil {
		return err
	}
	seri, err := jTimelockMilestoneFeatBlock.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*TimelockUnixFeatureBlock)
	return nil
}

// jsonTimelockUnixFeatureBlock defines the json representation of a TimelockUnixFeatureBlock.
type jsonTimelockUnixFeatureBlock struct {
	Type     int `json:"type"`
	UnixTime int `json:"unixTime"`
}

func (j *jsonTimelockUnixFeatureBlock) ToSerializable() (serializer.Serializable, error) {
	return &TimelockUnixFeatureBlock{UnixTime: uint64(j.UnixTime)}, nil
}
