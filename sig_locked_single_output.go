package iota

import (
	"encoding/json"
	"fmt"
)

const (
	// The size of a sig locked single deposit containing a WOTS address as its deposit address.
	SigLockedSingleOutputWOTSAddrBytesSize = SmallTypeDenotationByteSize + WOTSAddressSerializedBytesSize + UInt64ByteSize
	// The size of a sig locked single deposit containing an Ed25519 address as its deposit address.
	SigLockedSingleOutputEd25519AddrBytesSize = SmallTypeDenotationByteSize + Ed25519AddressSerializedBytesSize + UInt64ByteSize

	// Defines the minimum size a sig locked single deposit must be.
	SigLockedSingleOutputBytesMinSize = SigLockedSingleOutputEd25519AddrBytesSize
	// Defines the offset at which the address portion within a sig locked single deposit begins.
	SigLockedSingleOutputAddressOffset = SmallTypeDenotationByteSize
)

// SigLockedSingleOutput is an output type which can be unlocked via a signature. It deposits onto one single address.
type SigLockedSingleOutput struct {
	// The actual address.
	Address Serializable `json:"address"`
	// The amount to deposit.
	Amount uint64 `json:"amount"`
}

func (s *SigLockedSingleOutput) Type() OutputType {
	return OutputSigLockedSingleOutput
}

func (s *SigLockedSingleOutput) Target() (Serializable, error) {
	return s.Address, nil
}

func (s *SigLockedSingleOutput) Deposit() (uint64, error) {
	return s.Amount, nil
}

func (s *SigLockedSingleOutput) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	return NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := checkMinByteLength(SigLockedSingleOutputBytesMinSize, len(data)); err != nil {
					return fmt.Errorf("invalid signature locked single output bytes: %w", err)
				}
				if err := checkTypeByte(data, OutputSigLockedSingleOutput); err != nil {
					return fmt.Errorf("unable to deserialize signature locked single output: %w", err)
				}
			}
			return nil
		}).
		Skip(SmallTypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip signature locked single output type during deserialization: %w", err)
		}).
		ReadObject(func(seri Serializable) { s.Address = seri }, deSeriMode, TypeDenotationByte, AddressSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize address for signature locked single output: %w", err)
		}).
		ReadNum(&s.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for signature locked single output: %w", err)
		}).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, s); err != nil {
					return fmt.Errorf("%w: unable to deserialize signature locked single output", err)
				}
			}
			return nil
		}).
		Done()
}

func (s *SigLockedSingleOutput) Serialize(deSeriMode DeSerializationMode) (data []byte, err error) {
	return NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, s); err != nil {
					return fmt.Errorf("%w: unable to serialize signature locked single output", err)
				}

				switch s.Address.(type) {
				case *WOTSAddress:
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
	jsonDep := &jsonsiglockedsingleoutput{}

	addrJsonBytes, err := s.Address.MarshalJSON()
	if err != nil {
		return nil, err
	}
	jsonRawMsgAddr := json.RawMessage(addrJsonBytes)

	jsonDep.Address = &jsonRawMsgAddr
	jsonDep.Amount = int(s.Amount)
	jsonDep.Type = int(OutputSigLockedSingleOutput)
	return json.Marshal(jsonDep)
}

func (s *SigLockedSingleOutput) UnmarshalJSON(bytes []byte) error {
	jsonDep := &jsonsiglockedsingleoutput{}
	if err := json.Unmarshal(bytes, jsonDep); err != nil {
		return err
	}
	seri, err := jsonDep.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*SigLockedSingleOutput)
	return nil
}

// jsonsiglockedsingleoutput defines the json representation of a SigLockedSingleOutput.
type jsonsiglockedsingleoutput struct {
	Type    int              `json:"type"`
	Address *json.RawMessage `json:"address"`
	Amount  int              `json:"amount"`
}

func (j *jsonsiglockedsingleoutput) ToSerializable() (Serializable, error) {
	dep := &SigLockedSingleOutput{Amount: uint64(j.Amount)}

	jsonAddr, err := DeserializeObjectFromJSON(j.Address, jsonaddressselector)
	if err != nil {
		return nil, fmt.Errorf("can't decode address type from JSON: %w", err)
	}

	dep.Address, err = jsonAddr.ToSerializable()
	if err != nil {
		return nil, err
	}
	return dep, nil
}
