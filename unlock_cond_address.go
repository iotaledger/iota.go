package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

var addrUnlockCondAddrGuard = &serializer.SerializableGuard{
	ReadGuard:  addrReadGuard(allAddressTypeSet),
	WriteGuard: addrWriteGuard(allAddressTypeSet),
}

// AddressUnlockCondition is an UnlockCondition defining an identity which has to be unlocked.
type AddressUnlockCondition struct {
	Address Address
}

func (s *AddressUnlockCondition) Clone() UnlockCondition {
	return &AddressUnlockCondition{Address: s.Address.Clone()}
}

func (s *AddressUnlockCondition) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize) +
		s.Address.VByteCost(costStruct, nil)
}

func (s *AddressUnlockCondition) Equal(other UnlockCondition) bool {
	otherUnlockCond, is := other.(*AddressUnlockCondition)
	if !is {
		return false
	}

	return s.Address.Equal(otherUnlockCond.Address)
}

func (s *AddressUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionAddress
}

func (s *AddressUnlockCondition) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(UnlockConditionAddress), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize address unlock condition: %w", err)
		}).
		ReadObject(&s.Address, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, addrUnlockCondAddrGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize address for address unlock condition: %w", err)
		}).
		Done()
}

func (s *AddressUnlockCondition) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(UnlockConditionAddress), func(err error) error {
			return fmt.Errorf("unable to serialize address unlock condition type ID: %w", err)
		}).
		WriteObject(s.Address, deSeriMode, deSeriCtx, addrUnlockCondAddrGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize address unlock condition address: %w", err)
		}).
		Serialize()
}

func (s *AddressUnlockCondition) Size() int {
	return util.NumByteLen(byte(UnlockConditionAddress)) + s.Address.Size()
}

func (s *AddressUnlockCondition) MarshalJSON() ([]byte, error) {
	jUnlockCond := &jsonAddressUnlockCondition{}
	jUnlockCond.Type = int(UnlockConditionAddress)
	var err error
	jUnlockCond.Address, err = addressToJSONRawMsg(s.Address)
	if err != nil {
		return nil, err
	}
	return json.Marshal(jUnlockCond)
}

func (s *AddressUnlockCondition) UnmarshalJSON(bytes []byte) error {
	jUnlockCond := &jsonAddressUnlockCondition{}
	if err := json.Unmarshal(bytes, jUnlockCond); err != nil {
		return err
	}
	seri, err := jUnlockCond.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*AddressUnlockCondition)
	return nil
}

// jsonAddressUnlockCondition defines the json representation of an AddressUnlockCondition.
type jsonAddressUnlockCondition struct {
	Type    int              `json:"type"`
	Address *json.RawMessage `json:"address"`
}

func (j *jsonAddressUnlockCondition) ToSerializable() (serializer.Serializable, error) {
	unlockCond := &AddressUnlockCondition{}

	var err error
	unlockCond.Address, err = addressFromJSONRawMsg(j.Address)
	if err != nil {
		return nil, err
	}
	return unlockCond, nil
}
