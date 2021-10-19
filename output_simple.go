package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// SimpleOutputEd25519AddrBytesSize defines the size of a SimpleOutput containing an Ed25519Address as its deposit address.
	SimpleOutputEd25519AddrBytesSize = serializer.SmallTypeDenotationByteSize + Ed25519AddressSerializedBytesSize + serializer.UInt64ByteSize

	// SimpleOutputBytesMinSize defines the minimum size a SimpleOutput.
	SimpleOutputBytesMinSize = SimpleOutputEd25519AddrBytesSize
	// SimpleOutputAddressOffset defines the offset at which the address portion within a SimpleOutput begins.
	SimpleOutputAddressOffset = serializer.SmallTypeDenotationByteSize
)

// SimpleOutput is an output type which can be unlocked via a signature. It deposits onto one single address.
type SimpleOutput struct {
	// The actual address.
	Address serializer.Serializable `json:"address"`
	// The amount of IOTA tokens held by the output.
	Amount uint64 `json:"amount"`
}

func (s *SimpleOutput) Type() OutputType {
	return OutputSimple
}

func (s *SimpleOutput) Target() (serializer.Serializable, error) {
	return s.Address, nil
}

func (s *SimpleOutput) Deposit() (uint64, error) {
	return s.Amount, nil
}

func (s *SimpleOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := serializer.CheckMinByteLength(SimpleOutputBytesMinSize, len(data)); err != nil {
					return fmt.Errorf("invalid simple output bytes: %w", err)
				}
				if err := serializer.CheckTypeByte(data, OutputSimple); err != nil {
					return fmt.Errorf("unable to deserialize simple output: %w", err)
				}
			}
			return nil
		}).
		Skip(serializer.SmallTypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip simple output type during deserialization: %w", err)
		}).
		ReadObject(func(seri serializer.Serializable) { s.Address = seri }, deSeriMode, serializer.TypeDenotationByte, AddressSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize address for simple output: %w", err)
		}).
		ReadNum(&s.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for simple output: %w", err)
		}).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, s); err != nil {
					return fmt.Errorf("%w: unable to deserialize simple output", err)
				}
			}
			return nil
		}).
		Done()
}

func (s *SimpleOutput) Serialize(deSeriMode serializer.DeSerializationMode) (data []byte, err error) {
	return serializer.NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, s); err != nil {
					return fmt.Errorf("%w: unable to serialize simple output", err)
				}

				if err := isValidAddrType(s.Address); err != nil {
					return fmt.Errorf("invalid address set in simple output: %w", err)
				}
			}
			return nil
		}).
		WriteNum(OutputSimple, func(err error) error {
			return fmt.Errorf("unable to serialize simple output type ID: %w", err)
		}).
		WriteObject(s.Address, deSeriMode, func(err error) error {
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
