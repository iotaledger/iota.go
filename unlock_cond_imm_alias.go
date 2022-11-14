package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

var (
	immAliasUnlockCondAddrGuard = &serializer.SerializableGuard{
		ReadGuard:  addrReadGuard(AddressTypeSet{AddressAlias: struct{}{}}),
		WriteGuard: addrWriteGuard(AddressTypeSet{AddressAlias: struct{}{}}),
	}
)

// ImmutableAliasUnlockCondition is an UnlockCondition defining an alias which has to be unlocked.
// Unlike the AddressUnlockCondition, this unlock condition is immutable for an output which contains it,
// meaning it also only applies to ChainConstrainedOutput(s).
type ImmutableAliasUnlockCondition struct {
	Address *AliasAddress
}

func (s *ImmutableAliasUnlockCondition) Clone() UnlockCondition {
	return &ImmutableAliasUnlockCondition{Address: s.Address.Clone().(*AliasAddress)}
}

func (s *ImmutableAliasUnlockCondition) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize) +
		s.Address.VBytes(rentStruct, nil)
}

func (s *ImmutableAliasUnlockCondition) Equal(other UnlockCondition) bool {
	otherUnlockCond, is := other.(*ImmutableAliasUnlockCondition)
	if !is {
		return false
	}

	return s.Address.Equal(otherUnlockCond.Address)
}

func (s *ImmutableAliasUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionImmutableAlias
}

func (s *ImmutableAliasUnlockCondition) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(UnlockConditionImmutableAlias), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize immutable alias unlock condition: %w", err)
		}).
		ReadObject(&s.Address, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, immAliasUnlockCondAddrGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize address for immutable alias unlock condition: %w", err)
		}).
		Done()
}

func (s *ImmutableAliasUnlockCondition) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(UnlockConditionImmutableAlias), func(err error) error {
			return fmt.Errorf("unable to serialize immutable alias unlock condition type ID: %w", err)
		}).
		WriteObject(s.Address, deSeriMode, deSeriCtx, immAliasUnlockCondAddrGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize immutable alias unlock condition address: %w", err)
		}).
		Serialize()
}

func (s *ImmutableAliasUnlockCondition) Size() int {
	return util.NumByteLen(byte(UnlockConditionImmutableAlias)) + s.Address.Size()
}

func (s *ImmutableAliasUnlockCondition) MarshalJSON() ([]byte, error) {
	jUnlockCond := &jsonImmutableAliasUnlockCondition{}
	jUnlockCond.Type = int(UnlockConditionImmutableAlias)
	var err error
	jUnlockCond.Address, err = AddressToJSONRawMsg(s.Address)
	if err != nil {
		return nil, err
	}
	return json.Marshal(jUnlockCond)
}

func (s *ImmutableAliasUnlockCondition) UnmarshalJSON(bytes []byte) error {
	jUnlockCond := &jsonImmutableAliasUnlockCondition{}
	if err := json.Unmarshal(bytes, jUnlockCond); err != nil {
		return err
	}
	seri, err := jUnlockCond.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*ImmutableAliasUnlockCondition)
	return nil
}

// jsonImmutableAliasUnlockCondition defines the json representation of an ImmutableAliasUnlockCondition.
type jsonImmutableAliasUnlockCondition struct {
	Type    int              `json:"type"`
	Address *json.RawMessage `json:"address"`
}

func (j *jsonImmutableAliasUnlockCondition) ToSerializable() (serializer.Serializable, error) {
	unlockCond := &ImmutableAliasUnlockCondition{}

	addr, err := AddressFromJSONRawMsg(j.Address)
	if err != nil {
		return nil, err
	}
	unlockCond.Address = addr.(*AliasAddress)
	return unlockCond, nil
}
