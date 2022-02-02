package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	govAddrUnlockCondAddrGuard = &serializer.SerializableGuard{
		ReadGuard:  addrReadGuard(allAddressTypeSet),
		WriteGuard: addrWriteGuard(allAddressTypeSet),
	}
)

// GovernorAddressUnlockCondition is an UnlockCondition defining the governor identity for an AliasOutput.
type GovernorAddressUnlockCondition struct {
	Address Address
}

func (s *GovernorAddressUnlockCondition) Clone() UnlockCondition {
	return &GovernorAddressUnlockCondition{Address: s.Address.Clone()}
}

func (s *GovernorAddressUnlockCondition) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize) +
		s.Address.VByteCost(costStruct, nil)
}

func (s *GovernorAddressUnlockCondition) Equal(other UnlockCondition) bool {
	otherUnlockCond, is := other.(*GovernorAddressUnlockCondition)
	if !is {
		return false
	}

	return s.Address.Equal(otherUnlockCond.Address)
}

func (s *GovernorAddressUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionGovernorAddress
}

func (s *GovernorAddressUnlockCondition) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(UnlockConditionGovernorAddress), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize governor address unlock condition: %w", err)
		}).
		ReadObject(&s.Address, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, govAddrUnlockCondAddrGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize address for governor address unlock condition: %w", err)
		}).
		Done()
}

func (s *GovernorAddressUnlockCondition) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(UnlockConditionGovernorAddress), func(err error) error {
			return fmt.Errorf("unable to serialize governor address unlock condition type ID: %w", err)
		}).
		WriteObject(s.Address, deSeriMode, deSeriCtx, govAddrUnlockCondAddrGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize governor address unlock condition address: %w", err)
		}).
		Serialize()
}

func (s *GovernorAddressUnlockCondition) MarshalJSON() ([]byte, error) {
	jUnlockCond := &jsonGovernorAddressUnlockCondition{}
	jUnlockCond.Type = int(UnlockConditionGovernorAddress)
	var err error
	jUnlockCond.Address, err = addressToJSONRawMsg(s.Address)
	if err != nil {
		return nil, err
	}
	return json.Marshal(jUnlockCond)
}

func (s *GovernorAddressUnlockCondition) UnmarshalJSON(bytes []byte) error {
	jUnlockCond := &jsonGovernorAddressUnlockCondition{}
	if err := json.Unmarshal(bytes, jUnlockCond); err != nil {
		return err
	}
	seri, err := jUnlockCond.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*GovernorAddressUnlockCondition)
	return nil
}

// jsonGovernorAddressUnlockCondition defines the json representation of a GovernorAddressUnlockCondition.
type jsonGovernorAddressUnlockCondition struct {
	Type    int              `json:"type"`
	Address *json.RawMessage `json:"address"`
}

func (j *jsonGovernorAddressUnlockCondition) ToSerializable() (serializer.Serializable, error) {
	unlockCond := &GovernorAddressUnlockCondition{}

	var err error
	unlockCond.Address, err = addressFromJSONRawMsg(j.Address)
	if err != nil {
		return nil, err
	}
	return unlockCond, nil
}
