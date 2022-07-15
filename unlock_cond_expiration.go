package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

var (
	expUnlockCondAddrGuard = &serializer.SerializableGuard{
		ReadGuard:  addrReadGuard(allAddressTypeSet),
		WriteGuard: addrWriteGuard(allAddressTypeSet),
	}
)

// ExpirationUnlockCondition is an unlock condition which puts a time constraint on whether the receiver or return identity
// can consume an output depending on the latest confirmed milestone's timestamp T:
//	- only the receiver identity can consume the output, if T is before than the one defined in the condition.
//	- only the return identity can consume the output, if T is at the same time or after the one defined in the condition.
type ExpirationUnlockCondition struct {
	// The identity who is allowed to use the output after the expiration has happened.
	ReturnAddress Address
	// The unix time in second resolution at which the expiration happens.
	UnixTime uint32
}

func (s *ExpirationUnlockCondition) Clone() UnlockCondition {
	return &ExpirationUnlockCondition{
		ReturnAddress: s.ReturnAddress.Clone(),
		UnixTime:      s.UnixTime,
	}
}

func (s *ExpirationUnlockCondition) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt32ByteSize) +
		s.ReturnAddress.VBytes(rentStruct, nil)
}

func (s *ExpirationUnlockCondition) Equal(other UnlockCondition) bool {
	otherCond, is := other.(*ExpirationUnlockCondition)
	if !is {
		return false
	}

	switch {
	case !s.ReturnAddress.Equal(otherCond.ReturnAddress):
		return false
	case s.UnixTime != otherCond.UnixTime:
		return false
	}

	return true
}

func (s *ExpirationUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionExpiration
}

func (s *ExpirationUnlockCondition) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(UnlockConditionExpiration), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize expiration unlock condition: %w", err)
		}).
		ReadObject(&s.ReturnAddress, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, expUnlockCondAddrGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize return address for expiration unlock condition: %w", err)
		}).
		ReadNum(&s.UnixTime, func(err error) error {
			return fmt.Errorf("unable to deserialize unix time for expiration unlock condition: %w", err)
		}).
		Done()
}

func (s *ExpirationUnlockCondition) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(UnlockConditionExpiration), func(err error) error {
			return fmt.Errorf("unable to serialize expiration unlock condition type ID: %w", err)
		}).
		WriteObject(s.ReturnAddress, deSeriMode, deSeriCtx, expUnlockCondAddrGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize expiration unlock condition return address: %w", err)
		}).
		WriteNum(s.UnixTime, func(err error) error {
			return fmt.Errorf("unable to serialize expiration unlock condition unix time: %w", err)
		}).
		Serialize()
}

func (s *ExpirationUnlockCondition) Size() int {
	return util.NumByteLen(byte(UnlockConditionExpiration)) + s.ReturnAddress.Size() +
		+util.NumByteLen(s.UnixTime)
}

func (s *ExpirationUnlockCondition) MarshalJSON() ([]byte, error) {
	jExpUnlockCond := &jsonExpirationUnlockCondition{
		UnixTime: int(s.UnixTime),
	}
	jExpUnlockCond.Type = int(UnlockConditionExpiration)
	var err error
	jExpUnlockCond.ReturnAddress, err = addressToJSONRawMsg(s.ReturnAddress)
	if err != nil {
		return nil, err
	}
	return json.Marshal(jExpUnlockCond)
}

func (s *ExpirationUnlockCondition) UnmarshalJSON(bytes []byte) error {
	jExpUnlockCond := &jsonExpirationUnlockCondition{}
	if err := json.Unmarshal(bytes, jExpUnlockCond); err != nil {
		return err
	}
	seri, err := jExpUnlockCond.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*ExpirationUnlockCondition)
	return nil
}

// jsonExpirationUnlockCondition defines the json representation of an ExpirationUnlockCondition.
type jsonExpirationUnlockCondition struct {
	Type          int              `json:"type"`
	ReturnAddress *json.RawMessage `json:"returnAddress"`
	UnixTime      int              `json:"unixTime"`
}

func (j *jsonExpirationUnlockCondition) ToSerializable() (serializer.Serializable, error) {
	unlockCondExp := &ExpirationUnlockCondition{
		UnixTime: uint32(j.UnixTime),
	}

	var err error
	unlockCondExp.ReturnAddress, err = addressFromJSONRawMsg(j.ReturnAddress)
	if err != nil {
		return nil, err
	}
	return unlockCondExp, nil
}
