package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

// TimelockUnlockCondition is an unlock condition which puts a time constraint on an output depending
// on the latest confirmed milestone's timestamp T:
//   - the output can only be consumed, if T is bigger than the one defined in the condition.
type TimelockUnlockCondition struct {
	// The unix time in second resolution until which the timelock applies (inclusive).
	UnixTime uint32
}

func (s *TimelockUnlockCondition) Clone() UnlockCondition {
	return &TimelockUnlockCondition{
		UnixTime: s.UnixTime,
	}
}

func (s *TimelockUnlockCondition) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize + serializer.UInt32ByteSize)
}

func (s *TimelockUnlockCondition) Equal(other UnlockCondition) bool {
	otherCond, is := other.(*TimelockUnlockCondition)
	if !is {
		return false
	}

	switch {
	case s.UnixTime != otherCond.UnixTime:
		return false
	}

	return true
}

func (s *TimelockUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionTimelock
}

func (s *TimelockUnlockCondition) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(UnlockConditionTimelock), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize timelock unlock condition: %w", err)
		}).
		ReadNum(&s.UnixTime, func(err error) error {
			return fmt.Errorf("unable to deserialize unix time for timelock unlock condition: %w", err)
		}).
		Done()
}

func (s *TimelockUnlockCondition) Serialize(_ serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(UnlockConditionTimelock), func(err error) error {
			return fmt.Errorf("unable to serialize timelock unlock condition type ID: %w", err)
		}).
		WriteNum(s.UnixTime, func(err error) error {
			return fmt.Errorf("unable to serialize timelock unlock condition unix time: %w", err)
		}).
		Serialize()
}

func (s *TimelockUnlockCondition) Size() int {
	return util.NumByteLen(byte(UnlockConditionTimelock)) +
		util.NumByteLen(s.UnixTime)
}

func (s *TimelockUnlockCondition) MarshalJSON() ([]byte, error) {
	jTimelockUnlockCond := &jsonTimelockUnlockCondition{
		UnixTime: int(s.UnixTime),
	}
	jTimelockUnlockCond.Type = int(UnlockConditionTimelock)

	return json.Marshal(jTimelockUnlockCond)
}

func (s *TimelockUnlockCondition) UnmarshalJSON(bytes []byte) error {
	jTimelockUnlockCond := &jsonTimelockUnlockCondition{}
	if err := json.Unmarshal(bytes, jTimelockUnlockCond); err != nil {
		return err
	}
	seri, err := jTimelockUnlockCond.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*TimelockUnlockCondition)

	return nil
}

// jsonTimelockUnlockCondition defines the json representation of a TimelockUnlockCondition.
type jsonTimelockUnlockCondition struct {
	Type     int `json:"type"`
	UnixTime int `json:"unixTime"`
}

func (j *jsonTimelockUnlockCondition) ToSerializable() (serializer.Serializable, error) {
	return &TimelockUnlockCondition{
		UnixTime: uint32(j.UnixTime),
	}, nil
}
