package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

var (
	storageDepReturnUnlockCondAddrGuard = &serializer.SerializableGuard{
		ReadGuard:  AddressReadGuard(allAddressTypeSet),
		WriteGuard: AddressWriteGuard(allAddressTypeSet),
	}
)

// StorageDepositReturnUnlockCondition is an unlock condition which defines
// the amount of tokens which must be sent back to the return identity, when the output in which it occurs in, is consumed.
// If a transaction consumes multiple outputs which have a StorageDepositReturnUnlockCondition, then on the output side at least
// the sum of all occurring StorageDepositReturnUnlockCondition(s) on the input side must be deposited to the designated return identity.
type StorageDepositReturnUnlockCondition struct {
	ReturnAddress Address
	Amount        uint64
}

func (s *StorageDepositReturnUnlockCondition) Clone() UnlockCondition {
	return &StorageDepositReturnUnlockCondition{
		ReturnAddress: s.ReturnAddress.Clone(),
		Amount:        s.Amount,
	}
}

func (s *StorageDepositReturnUnlockCondition) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		s.ReturnAddress.VBytes(rentStruct, nil)
}

func (s *StorageDepositReturnUnlockCondition) Equal(other UnlockCondition) bool {
	otherBlock, is := other.(*StorageDepositReturnUnlockCondition)
	if !is {
		return false
	}

	switch {
	case !s.ReturnAddress.Equal(otherBlock.ReturnAddress):
		return false
	case s.Amount != otherBlock.Amount:
		return false
	}

	return true
}

func (s *StorageDepositReturnUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionStorageDepositReturn
}

func (s *StorageDepositReturnUnlockCondition) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(UnlockConditionStorageDepositReturn), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize storage deposit return unlock condition: %w", err)
		}).
		ReadObject(&s.ReturnAddress, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, storageDepReturnUnlockCondAddrGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize return address for storage deposit return unlock condition: %w", err)
		}).
		ReadNum(&s.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for storage deposit return unlock condition: %w", err)
		}).
		Done()
}

func (s *StorageDepositReturnUnlockCondition) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(UnlockConditionStorageDepositReturn), func(err error) error {
			return fmt.Errorf("unable to serialize storage deposit return unlock condition type ID: %w", err)
		}).
		WriteObject(s.ReturnAddress, deSeriMode, deSeriCtx, storageDepReturnUnlockCondAddrGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize storage deposit return unlock condition return address: %w", err)
		}).
		WriteNum(s.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize storage deposit return unlock condition amount: %w", err)
		}).
		Serialize()
}

func (s *StorageDepositReturnUnlockCondition) Size() int {
	return util.NumByteLen(byte(UnlockConditionStorageDepositReturn)) + s.ReturnAddress.Size() + serializer.UInt64ByteSize
}

func (s *StorageDepositReturnUnlockCondition) MarshalJSON() ([]byte, error) {
	jUnlockCond := &jsonStorageDepositReturnUnlockCondition{Amount: EncodeUint64(s.Amount)}
	jUnlockCond.Type = int(UnlockConditionStorageDepositReturn)
	var err error
	jUnlockCond.ReturnAddress, err = AddressToJSONRawMsg(s.ReturnAddress)
	if err != nil {
		return nil, err
	}
	return json.Marshal(jUnlockCond)
}

func (s *StorageDepositReturnUnlockCondition) UnmarshalJSON(bytes []byte) error {
	jUnlockCond := &jsonStorageDepositReturnUnlockCondition{}
	if err := json.Unmarshal(bytes, jUnlockCond); err != nil {
		return err
	}
	seri, err := jUnlockCond.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*StorageDepositReturnUnlockCondition)
	return nil
}

// jsonStorageDepositReturnUnlockCondition defines the json representation of a StorageDepositReturnUnlockCondition.
type jsonStorageDepositReturnUnlockCondition struct {
	Type          int              `json:"type"`
	ReturnAddress *json.RawMessage `json:"returnAddress"`
	Amount        string           `json:"amount"`
}

func (j *jsonStorageDepositReturnUnlockCondition) ToSerializable() (serializer.Serializable, error) {
	var err error
	unlockCond := &StorageDepositReturnUnlockCondition{}

	unlockCond.Amount, err = DecodeUint64(j.Amount)
	if err != nil {
		return nil, err
	}

	unlockCond.ReturnAddress, err = AddressFromJSONRawMsg(j.ReturnAddress)
	if err != nil {
		return nil, err
	}

	return unlockCond, nil
}
