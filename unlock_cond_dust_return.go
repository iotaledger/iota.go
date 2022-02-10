package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

var (
	dustDepReturnUnlockCondAddrGuard = &serializer.SerializableGuard{
		ReadGuard:  addrReadGuard(allAddressTypeSet),
		WriteGuard: addrWriteGuard(allAddressTypeSet),
	}
)

// DustDepositReturnUnlockCondition is an unlock condition which defines
// the amount of tokens which must be sent back to the return identity, when the output in which it occurs in,
// is consumed.
// If a transaction consumes multiple outputs which have a DustDepositReturnUnlockCondition, then on the output side at least
// the sum of all occurring DustDepositReturnUnlockCondition(s) on the input side must be deposited to the designated return identity.
type DustDepositReturnUnlockCondition struct {
	ReturnAddress Address
	Amount        uint64
}

func (s *DustDepositReturnUnlockCondition) Clone() UnlockCondition {
	return &DustDepositReturnUnlockCondition{
		ReturnAddress: s.ReturnAddress.Clone(),
		Amount:        s.Amount,
	}
}

func (s *DustDepositReturnUnlockCondition) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		s.ReturnAddress.VByteCost(costStruct, nil)
}

func (s *DustDepositReturnUnlockCondition) Equal(other UnlockCondition) bool {
	otherBlock, is := other.(*DustDepositReturnUnlockCondition)
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

func (s *DustDepositReturnUnlockCondition) Type() UnlockConditionType {
	return UnlockConditionDustDepositReturn
}

func (s *DustDepositReturnUnlockCondition) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(UnlockConditionDustDepositReturn), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize dust deposit return unlock condition: %w", err)
		}).
		ReadObject(&s.ReturnAddress, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, dustDepReturnUnlockCondAddrGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize return address for dust deposit return unlock condition: %w", err)
		}).
		ReadNum(&s.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for dust deposit return unlock condition: %w", err)
		}).
		Done()
}

func (s *DustDepositReturnUnlockCondition) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(UnlockConditionDustDepositReturn), func(err error) error {
			return fmt.Errorf("unable to serialize dust deposit return unlock condition type ID: %w", err)
		}).
		WriteObject(s.ReturnAddress, deSeriMode, deSeriCtx, dustDepReturnUnlockCondAddrGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize dust deposit return unlock condition return address: %w", err)
		}).
		WriteNum(s.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize dust deposit return unlock condition amount: %w", err)
		}).
		Serialize()
}

func (s *DustDepositReturnUnlockCondition) Size() int {
	return util.NumByteLen(byte(UnlockConditionDustDepositReturn)) + s.ReturnAddress.Size() + 8
}

func (s *DustDepositReturnUnlockCondition) MarshalJSON() ([]byte, error) {
	jUnlockCond := &jsonDustDepositReturnUnlockCondition{Amount: int(s.Amount)}
	jUnlockCond.Type = int(UnlockConditionDustDepositReturn)
	var err error
	jUnlockCond.ReturnAddress, err = addressToJSONRawMsg(s.ReturnAddress)
	if err != nil {
		return nil, err
	}
	return json.Marshal(jUnlockCond)
}

func (s *DustDepositReturnUnlockCondition) UnmarshalJSON(bytes []byte) error {
	jUnlockCond := &jsonDustDepositReturnUnlockCondition{}
	if err := json.Unmarshal(bytes, jUnlockCond); err != nil {
		return err
	}
	seri, err := jUnlockCond.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*DustDepositReturnUnlockCondition)
	return nil
}

// jsonDustDepositReturnUnlockCondition defines the json representation of a DustDepositReturnUnlockCondition.
type jsonDustDepositReturnUnlockCondition struct {
	Type          int              `json:"type"`
	ReturnAddress *json.RawMessage `json:"returnAddress"`
	Amount        int              `json:"amount"`
}

func (j *jsonDustDepositReturnUnlockCondition) ToSerializable() (serializer.Serializable, error) {
	unlockCond := &DustDepositReturnUnlockCondition{Amount: uint64(j.Amount)}

	var err error
	unlockCond.ReturnAddress, err = addressFromJSONRawMsg(j.ReturnAddress)
	if err != nil {
		return nil, err
	}

	return unlockCond, nil
}
