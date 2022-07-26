package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	treasuryTxInputGuard = serializer.SerializableGuard{
		ReadGuard: func(ty uint32) (serializer.Serializable, error) {
			if InputType(ty) != InputTreasury {
				return nil, fmt.Errorf("%w: treasury tx only supports treasury input as input", ErrTypeIsNotSupportedInput)
			}
			return InputSelector(ty)
		},
		WriteGuard: func(seri serializer.Serializable) error {
			if seri == nil {
				return fmt.Errorf("%w: because nil", ErrTypeIsNotSupportedInput)
			}
			if _, is := seri.(*TreasuryInput); !is {
				return fmt.Errorf("%w: treasury tx only supports treasury input as input", ErrTypeIsNotSupportedInput)
			}
			return nil
		},
	}

	treasuryTxOutputGuard = serializer.SerializableGuard{
		ReadGuard: func(ty uint32) (serializer.Serializable, error) {
			if OutputType(ty) != OutputTreasury {
				return nil, fmt.Errorf("%w: treasury tx only supports treasury output as output", ErrTypeIsNotSupportedInput)
			}
			return OutputSelector(ty)
		},
		WriteGuard: func(seri serializer.Serializable) error {
			if seri == nil {
				return fmt.Errorf("%w: because nil", ErrTypeIsNotSupportedInput)
			}
			if _, is := seri.(*TreasuryOutput); !is {
				return fmt.Errorf("%w: treasury tx only supports treasury output as output", ErrTypeIsNotSupportedInput)
			}
			return nil
		},
	}
)

// TreasuryTransaction represents a transaction which moves funds from the treasury.
type TreasuryTransaction struct {
	// The input of this transaction.
	Input *TreasuryInput
	// The output of this transaction.
	Output *TreasuryOutput
}

func (t *TreasuryTransaction) Size() int {
	return serializer.UInt32ByteSize + t.Input.Size() + t.Output.Size()
}

func (t *TreasuryTransaction) Clone() *TreasuryTransaction {
	return &TreasuryTransaction{
		Input:  t.Input.Clone(),
		Output: t.Output.Clone().(*TreasuryOutput),
	}
}

func (t *TreasuryTransaction) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(PayloadTreasuryTransaction), serializer.TypeDenotationUint32, func(err error) error {
			return fmt.Errorf("unable to deserialize treasury transaction: %w", err)
		}).
		ReadObject(&t.Input, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, treasuryTxInputGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize treasury transaction input: %w", err)
		}).
		ReadObject(&t.Output, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, treasuryTxOutputGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize treasury transaction output: %w", err)
		}).
		Done()
}

func (t *TreasuryTransaction) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(PayloadTreasuryTransaction, func(err error) error {
			return fmt.Errorf("unable to serialize treasury transaction type ID: %w", err)
		}).
		WriteObject(t.Input, deSeriMode, deSeriCtx, treasuryTxInputGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize treasury transaction input: %w", err)
		}).
		WriteObject(t.Output, deSeriMode, deSeriCtx, treasuryTxOutputGuard.WriteGuard, func(err error) error {
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
