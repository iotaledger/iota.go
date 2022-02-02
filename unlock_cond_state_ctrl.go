package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	stateCtrlUnlockCondAddrGuard = &serializer.SerializableGuard{
		ReadGuard:  addrReadGuard(allAddressTypeSet),
		WriteGuard: addrWriteGuard(allAddressTypeSet),
	}
)

// StateControllerAddressUnlockCondition is an UnlockCondition defining the state controller identity for an AliasOutput.
type StateControllerAddressUnlockCondition struct {
	Address Address
}

func (s *StateControllerAddressUnlockCondition) Clone() UnlockCondition {
	return &StateControllerAddressUnlockCondition{Address: s.Address.Clone()}
}

func (s *StateControllerAddressUnlockCondition) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize) +
		s.Address.VByteCost(costStruct, nil)
}

func (s *StateControllerAddressUnlockCondition) Equal(other UnlockCondition) bool {
	otherUnlockCond, is := other.(*StateControllerAddressUnlockCondition)
	if !is {
		return false
	}

	return s.Address.Equal(otherUnlockCond.Address)
}

func (s *StateControllerAddressUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionStateControllerAddress
}

func (s *StateControllerAddressUnlockCondition) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(UnlockConditionStateControllerAddress), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize state controller address unlock condition: %w", err)
		}).
		ReadObject(&s.Address, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, stateCtrlUnlockCondAddrGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize address for state controller address unlock condition: %w", err)
		}).
		Done()
}

func (s *StateControllerAddressUnlockCondition) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(UnlockConditionStateControllerAddress), func(err error) error {
			return fmt.Errorf("unable to serialize state controller address unlock condition type ID: %w", err)
		}).
		WriteObject(s.Address, deSeriMode, deSeriCtx, stateCtrlUnlockCondAddrGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize state controller address unlock condition address: %w", err)
		}).
		Serialize()
}

func (s *StateControllerAddressUnlockCondition) MarshalJSON() ([]byte, error) {
	jUnlockCond := &jsonStateControllerAddressUnlockCondition{}
	jUnlockCond.Type = int(UnlockConditionStateControllerAddress)
	var err error
	jUnlockCond.Address, err = addressToJSONRawMsg(s.Address)
	if err != nil {
		return nil, err
	}
	return json.Marshal(jUnlockCond)
}

func (s *StateControllerAddressUnlockCondition) UnmarshalJSON(bytes []byte) error {
	jUnlockCond := &jsonStateControllerAddressUnlockCondition{}
	if err := json.Unmarshal(bytes, jUnlockCond); err != nil {
		return err
	}
	seri, err := jUnlockCond.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*StateControllerAddressUnlockCondition)
	return nil
}

// jsonStateControllerAddressUnlockCondition defines the json representation of a StateControllerAddressUnlockCondition.
type jsonStateControllerAddressUnlockCondition struct {
	Type    int              `json:"type"`
	Address *json.RawMessage `json:"address"`
}

func (j *jsonStateControllerAddressUnlockCondition) ToSerializable() (serializer.Serializable, error) {
	unlockCond := &StateControllerAddressUnlockCondition{}

	var err error
	unlockCond.Address, err = addressFromJSONRawMsg(j.Address)
	if err != nil {
		return nil, err
	}
	return unlockCond, nil
}