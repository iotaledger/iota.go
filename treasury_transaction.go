package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// TreasuryTransactionByteSize defines the serialized size of a TreasuryTransaction.
	TreasuryTransactionByteSize = serializer.TypeDenotationByteSize + TreasuryInputSerializedBytesSize + TreasuryOutputBytesSize
)

// TreasuryTransaction represents a transaction which moves funds from the treasury.
type TreasuryTransaction struct {
	// The input of this transaction.
	Input *TreasuryInput
	// The output of this transaction.
	Output *TreasuryOutput
}

func (t *TreasuryTransaction) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := serializer.CheckMinByteLength(TreasuryTransactionByteSize, len(data)); err != nil {
					return fmt.Errorf("invalid treasury transaction bytes: %w", err)
				}
				if err := serializer.CheckType(data, uint32(PayloadTreasuryTransaction)); err != nil {
					return fmt.Errorf("unable to deserialize treasury transaction: %w", err)
				}
			}
			return nil
		}).
		Skip(serializer.TypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip treasury transaction payload ID during deserialization: %w", err)
		}).
		ReadObject(&t.Input, deSeriMode, serializer.TypeDenotationByte, treasuryTxInputGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize treasury transaction input: %w", err)
		}).
		ReadObject(&t.Output, deSeriMode, serializer.TypeDenotationByte, treasuryTxOutputGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize treasury transaction output: %w", err)
		}).
		Done()
}

func treasuryTxOutputGuard(ty uint32) (serializer.Serializable, error) {
	if ty != uint32(OutputTreasury) {
		return nil, fmt.Errorf("receipts can only contain treasury output as outputs but got type ID %d: %w", ty, ErrUnsupportedObjectType)
	}
	return OutputSelector(ty)
}

func treasuryTxInputGuard(ty uint32) (serializer.Serializable, error) {
	if ty != uint32(InputTreasury) {
		return nil, fmt.Errorf("receipts can only contain treasury input as inputs but got type ID %d: %w", ty, ErrUnsupportedObjectType)
	}
	return InputSelector(ty)
}

func (t *TreasuryTransaction) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(PayloadTreasuryTransaction, func(err error) error {
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
	jTreasuryTransaction.Type = int(PayloadTreasuryTransaction)

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

func (j *jsonTreasuryTransaction) ToSerializable() (serializer.Serializable, error) {
	dep := &TreasuryTransaction{}

	jsonInput, err := DeserializeObjectFromJSON(j.Input, func(ty int) (JSONSerializable, error) {
		return &jsonTreasuryInput{}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't decode input from JSON: %w", err)
	}

	input, err := jsonInput.ToSerializable()
	if err != nil {
		return nil, err
	}
	dep.Input = input.(*TreasuryInput)

	jsonOutput, err := DeserializeObjectFromJSON(j.Output, func(ty int) (JSONSerializable, error) {
		return &jsonTreasuryOutput{}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't decode treasury output from JSON: %w", err)
	}

	output, err := jsonOutput.ToSerializable()
	if err != nil {
		return nil, err
	}
	dep.Output = output.(*TreasuryOutput)

	return dep, nil
}
