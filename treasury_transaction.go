package iotago

import (
	"encoding/json"
	"fmt"
)

const (
	// Defines the TreasuryTransaction payload's ID.
	TreasuryTransactionPayloadTypeID uint32 = 4
	// Defines the serialized size of a TreasuryTransaction.
	TreasuryTransactionByteSize = TypeDenotationByteSize + TreasuryInputSerializedBytesSize + TreasuryOutputBytesSize
)

// TreasuryTransaction represents a transaction which moves funds from the treasury.
type TreasuryTransaction struct {
	// The input of this transaction.
	Input Serializable `json:"input"`
	// The output of this transaction.
	Output Serializable `json:"output"`
}

func (t *TreasuryTransaction) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	return NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := checkMinByteLength(TreasuryTransactionByteSize, len(data)); err != nil {
					return fmt.Errorf("invalid treasury transaction bytes: %w", err)
				}
				if err := checkType(data, TreasuryTransactionPayloadTypeID); err != nil {
					return fmt.Errorf("unable to deserialize treasury transaction: %w", err)
				}
			}
			return nil
		}).
		Skip(TypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip treasury transaction payload ID during deserialization: %w", err)
		}).
		ReadObject(func(seri Serializable) { t.Input = seri }, deSeriMode, TypeDenotationByte, func(ty uint32) (Serializable, error) {
			if ty != uint32(InputTreasury) {
				return nil, fmt.Errorf("receipts can only contain treasury input as inputs but got type ID %d: %w", ty, ErrUnsupportedObjectType)
			}
			return InputSelector(ty)
		}, func(err error) error {
			return fmt.Errorf("unable to deserialize treasury transaction input: %w", err)
		}).
		ReadObject(func(seri Serializable) { t.Output = seri }, deSeriMode, TypeDenotationByte, func(ty uint32) (Serializable, error) {
			if ty != uint32(OutputTreasuryOutput) {
				return nil, fmt.Errorf("receipts can only contain treasury output as outputs but got type ID %d: %w", ty, ErrUnsupportedObjectType)
			}
			return OutputSelector(ty)
		}, func(err error) error {
			return fmt.Errorf("unable to deserialize treasury transaction output: %w", err)
		}).
		Done()
}

func (t *TreasuryTransaction) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if _, isUTXOInput := t.Input.(*TreasuryInput); !isUTXOInput {
			return nil, fmt.Errorf("%w: treasury transaction must contain a UTXO input but got %T instead", ErrInvalidBytes, t.Input)
		}
		if _, isTreasuryOutput := t.Output.(*TreasuryOutput); !isTreasuryOutput {
			return nil, fmt.Errorf("%w: treasury transaction must contain a treasury output but got %T instead", ErrInvalidBytes, t.Output)
		}
	}
	return NewSerializer().
		WriteNum(TreasuryTransactionPayloadTypeID, func(err error) error {
			return fmt.Errorf("unable to serialize treasury transaction type ID: %w", err)
		}).
		WriteObject(t.Input, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize treasury transaction input: %w", err)
		}).
		WriteObject(t.Output, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize treasury transaction output: %w", err)
		}).
		Serialize()
}

func (t *TreasuryTransaction) MarshalJSON() ([]byte, error) {
	jTreasuryTransaction := &jsonTreasuryTransaction{}
	jTreasuryTransaction.Type = int(TreasuryTransactionPayloadTypeID)

	jsonInput, err := t.Input.MarshalJSON()
	if err != nil {
		return nil, err
	}
	rawJsonInput := json.RawMessage(jsonInput)
	jTreasuryTransaction.Input = &rawJsonInput

	jsonOutput, err := t.Output.MarshalJSON()
	if err != nil {
		return nil, err
	}
	rawJsonOutput := json.RawMessage(jsonOutput)
	jTreasuryTransaction.Output = &rawJsonOutput

	return json.Marshal(jTreasuryTransaction)
}

func (t *TreasuryTransaction) UnmarshalJSON(bytes []byte) error {
	jTreasuryTransaction := &jsonTreasuryTransaction{}
	if err := json.Unmarshal(bytes, jTreasuryTransaction); err != nil {
		return err
	}
	seri, err := jTreasuryTransaction.ToSerializable()
	if err != nil {
		return err
	}
	*t = *seri.(*TreasuryTransaction)
	return nil
}

// jsonTreasuryTransaction defines the json representation of a TreasuryTransaction.
type jsonTreasuryTransaction struct {
	Type   int              `json:"type"`
	Input  *json.RawMessage `json:"input"`
	Output *json.RawMessage `json:"output"`
}

func (j *jsonTreasuryTransaction) ToSerializable() (Serializable, error) {
	dep := &TreasuryTransaction{}

	jsonInput, err := DeserializeObjectFromJSON(j.Input, func(ty int) (JSONSerializable, error) {
		return &jsonTreasuryInput{}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't decode input from JSON: %w", err)
	}

	dep.Input, err = jsonInput.ToSerializable()
	if err != nil {
		return nil, err
	}

	jsonOutput, err := DeserializeObjectFromJSON(j.Output, func(ty int) (JSONSerializable, error) {
		return &jsonTreasuryOutput{}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't decode treasury output from JSON: %w", err)
	}

	dep.Output, err = jsonOutput.ToSerializable()
	if err != nil {
		return nil, err
	}

	return dep, nil
}
