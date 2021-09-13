package iotago

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/hive.go/serializer"
)

const (
	// SigLockedSingleOutputEd25519AddrBytesSize defines the size of a SigLockedSingleOutput containing an Ed25519Address as its deposit address.
	SigLockedSingleOutputEd25519AddrBytesSize = serializer.SmallTypeDenotationByteSize + Ed25519AddressSerializedBytesSize + serializer.UInt64ByteSize

	// SigLockedSingleOutputBytesMinSize defines the minimum size a SigLockedSingleOutput.
	SigLockedSingleOutputBytesMinSize = SigLockedSingleOutputEd25519AddrBytesSize
	// SigLockedSingleOutputAddressOffset defines the offset at which the address portion within a SigLockedSingleOutput begins.
	SigLockedSingleOutputAddressOffset = serializer.SmallTypeDenotationByteSize
)

// SigLockedSingleOutput is an output type which can be unlocked via a signature. It deposits onto one single address.
type SigLockedSingleOutput struct {
	// The actual address.
	Address serializer.Serializable `json:"address"`
	// The amount to deposit.
	Amount uint64 `json:"amount"`
}

func (s *SigLockedSingleOutput) Type() OutputType {
	return OutputSigLockedSingleOutput
}

func (s *SigLockedSingleOutput) Target() (serializer.Serializable, error) {
	return s.Address, nil
}

func (s *SigLockedSingleOutput) Deposit() (uint64, error) {
	return s.Amount, nil
}

func (s *SigLockedSingleOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := serializer.CheckMinByteLength(SigLockedSingleOutputBytesMinSize, len(data)); err != nil {
					return fmt.Errorf("invalid signature locked single output bytes: %w", err)
				}
				if err := serializer.CheckTypeByte(data, OutputSigLockedSingleOutput); err != nil {
					return fmt.Errorf("unable to deserialize signature locked single output: %w", err)
				}
			}
			return nil
		}).
		Skip(serializer.SmallTypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip signature locked single output type during deserialization: %w", err)
		}).
		ReadObject(func(seri serializer.Serializable) { s.Address = seri }, deSeriMode, serializer.TypeDenotationByte, AddressSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize address for signature locked single output: %w", err)
		}).
		ReadNum(&s.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for signature locked single output: %w", err)
		}).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, s); err != nil {
					return fmt.Errorf("%w: unable to deserialize signature locked single output", err)
				}
			}
			return nil
		}).
		Done()
}

func (s *SigLockedSingleOutput) Serialize(deSeriMode serializer.DeSerializationMode) (data []byte, err error) {
	return serializer.NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, s); err != nil {
					return fmt.Errorf("%w: unable to serialize signature locked single output", err)
				}

				switch s.Address.(type) {
				case *Ed25519Address:
				default:
					return fmt.Errorf("%w: signature locked single output defines unknown address", ErrUnknownAddrType)
				}
			}
			return nil
		}).
		WriteNum(OutputSigLockedSingleOutput, func(err error) error {
			return fmt.Errorf("unable to serialize signature locked single output type ID: %w", err)
		}).
		WriteObject(s.Address, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize signature locked single output address: %w", err)
		}).
		WriteNum(s.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize signature locked single output amount: %w", err)
		}).Serialize()
}

func (s *SigLockedSingleOutput) MarshalJSON() ([]byte, error) {
	jSigLockedSingleOutput := &jsonSigLockedSingleOutput{}

	addrJsonBytes, err := s.Address.MarshalJSON()
	if err != nil {
		return nil, err
	}
	jsonRawMsgAddr := json.RawMessage(addrJsonBytes)

	jSigLockedSingleOutput.Type = int(OutputSigLockedSingleOutput)
	jSigLockedSingleOutput.Address = &jsonRawMsgAddr
	jSigLockedSingleOutput.Amount = int(s.Amount)
	return json.Marshal(jSigLockedSingleOutput)
}

func (s *SigLockedSingleOutput) UnmarshalJSON(bytes []byte) error {
	jSigLockedSingleOutput := &jsonSigLockedSingleOutput{}
	if err := json.Unmarshal(bytes, jSigLockedSingleOutput); err != nil {
		return err
	}
	seri, err := jSigLockedSingleOutput.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*SigLockedSingleOutput)
	return nil
}

// jsonSigLockedSingleOutput defines the json representation of a SigLockedSingleOutput.
type jsonSigLockedSingleOutput struct {
	Type    int              `json:"type"`
	Address *json.RawMessage `json:"address"`
	Amount  int              `json:"amount"`
}

func (j *jsonSigLockedSingleOutput) ToSerializable() (serializer.Serializable, error) {
	dep := &SigLockedSingleOutput{Amount: uint64(j.Amount)}

	jsonAddr, err := DeserializeObjectFromJSON(j.Address, jsonAddressSelector)
	if err != nil {
		return nil, fmt.Errorf("can't decode address type from JSON: %w", err)
	}

	dep.Address, err = jsonAddr.ToSerializable()
	if err != nil {
		return nil, err
	}
	return dep, nil
}
