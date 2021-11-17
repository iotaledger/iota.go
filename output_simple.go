package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// SimpleOutputAddressOffset defines the offset at which the address portion within a SimpleOutput begins.
	SimpleOutputAddressOffset = serializer.SmallTypeDenotationByteSize
)

var (
	simpleOutputAddrGuard = &serializer.SerializableGuard{
		ReadGuard:  addrReadGuard(allAddressTypeSet),
		WriteGuard: addrWriteGuard(allAddressTypeSet),
	}
)

// SimpleOutput is an output type which can be unlocked via a signature. It deposits onto one single address.
type SimpleOutput struct {
	// The actual address.
	Address Address
	// The amount of IOTA tokens held by the output.
	Amount uint64
}

func (s *SimpleOutput) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return costStruct.VBFactorKey.Multiply(OutputIDLength) +
		costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		s.Address.VByteCost(costStruct, nil)
}

func (s *SimpleOutput) Type() OutputType {
	return OutputSimple
}

func (s *SimpleOutput) Ident() (Address, error) {
	return s.Address, nil
}

func (s *SimpleOutput) Deposit() (uint64, error) {
	return s.Amount, nil
}

func (s *SimpleOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(OutputSimple), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize simple output: %w", err)
		}).
		ReadObject(&s.Address, deSeriMode, serializer.TypeDenotationByte, simpleOutputAddrGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize address for simple output: %w", err)
		}).
		ReadNum(&s.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for simple output: %w", err)
		}).
		Done()
}

func (s *SimpleOutput) Serialize(deSeriMode serializer.DeSerializationMode) (data []byte, err error) {
	return serializer.NewSerializer().
		WriteNum(OutputSimple, func(err error) error {
			return fmt.Errorf("unable to serialize simple output type ID: %w", err)
		}).
		WriteObject(s.Address, deSeriMode, simpleOutputAddrGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize simple output address: %w", err)
		}).
		WriteNum(s.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize simple output amount: %w", err)
		}).Serialize()
}

func (s *SimpleOutput) MarshalJSON() ([]byte, error) {
	jSimpleOutput := &jsonSimpleOutput{}

	addrJsonBytes, err := s.Address.MarshalJSON()
	if err != nil {
		return nil, err
	}
	jsonRawMsgAddr := json.RawMessage(addrJsonBytes)

	jSimpleOutput.Type = int(OutputSimple)
	jSimpleOutput.Address = &jsonRawMsgAddr
	jSimpleOutput.Amount = int(s.Amount)
	return json.Marshal(jSimpleOutput)
}

func (s *SimpleOutput) UnmarshalJSON(bytes []byte) error {
	jSimpleOutput := &jsonSimpleOutput{}
	if err := json.Unmarshal(bytes, jSimpleOutput); err != nil {
		return err
	}
	seri, err := jSimpleOutput.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*SimpleOutput)
	return nil
}

// jsonSimpleOutput defines the json representation of a SimpleOutput.
type jsonSimpleOutput struct {
	Type    int              `json:"type"`
	Address *json.RawMessage `json:"address"`
	Amount  int              `json:"amount"`
}

func (j *jsonSimpleOutput) ToSerializable() (serializer.Serializable, error) {
	dep := &SimpleOutput{Amount: uint64(j.Amount)}

	addr, err := addressFromJSONRawMsg(j.Address)
	if err != nil {
		return nil, err
	}
	dep.Address = addr

	return dep, nil
}
