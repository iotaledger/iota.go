package iotago

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/iotaledger/hive.go/serializer"
)

// ExtendedOutput is an output type which can hold native tokens and feature blocks.
type ExtendedOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64
	// The native tokens held by the output.
	NativeTokens serializer.Serializables
	// The deposit address.
	Address serializer.Serializable
	// The feature blocks which modulate the constraints on the output.
	Blocks serializer.Serializables
}

func (e *ExtendedOutput) NativeTokenSet() serializer.Serializables {
	return e.NativeTokens
}

func (e *ExtendedOutput) Deposit() (uint64, error) {
	return e.Amount, nil
}

func (e *ExtendedOutput) Target() (serializer.Serializable, error) {
	return e.Address, nil
}

func (e *ExtendedOutput) Type() OutputType {
	return OutputExtended
}

func (e *ExtendedOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := serializer.CheckTypeByte(data, OutputExtended); err != nil {
					return fmt.Errorf("unable to deserialize extended output: %w", err)
				}
			}
			return nil
		}).
		Skip(serializer.SmallTypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip extended output type during deserialization: %w", err)
		}).
		ReadNum(&e.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for extended output: %w", err)
		}).
		ReadSliceOfObjects(func(seri serializer.Serializables) { e.NativeTokens = seri }, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationNone, func(ty uint32) (serializer.Serializable, error) {
			return &NativeToken{}, nil
		}, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize native tokens for extended output: %w", err)
		}).
		ReadObject(func(seri serializer.Serializable) { e.Address = seri }, deSeriMode, serializer.TypeDenotationByte, AddressSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize address for extended output: %w", err)
		}).
		ReadSliceOfObjects(func(seri serializer.Serializables) { e.Blocks = seri }, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationByte, FeatureBlockSelector, featBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize feature blocks for extended output: %w", err)
		}).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, e); err != nil {
					return fmt.Errorf("%w: unable to deserialize extended output", err)
				}
			}
			return nil
		}).
		Done()
}

func (e *ExtendedOutput) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	var nativeTokensWrittenConsumer serializer.WrittenObjectConsumer
	return serializer.NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, e); err != nil {
					return fmt.Errorf("%w: unable to serialize extended output", err)
				}

				if err := isValidAddrType(e.Address); err != nil {
					return fmt.Errorf("invalid address set in extended output: %w", err)
				}
				nativeTokensLexicalNoDupsValidator := nativeTokensArrayRules.LexicalOrderWithoutDupsValidator()
				nativeTokensWrittenConsumer = func(index int, written []byte) error {
					if err := nativeTokensLexicalNoDupsValidator(index, written); err != nil {
						return fmt.Errorf("%w: unable to serialize native tokens of extended output since inputs are not lexically sorted or contain duplicates", err)
					}
					return nil
				}
			}
			return nil
		}).
		Do(func() {
			if deSeriMode.HasMode(serializer.DeSeriModePerformLexicalOrdering) {
				sort.Sort(serializer.SortedSerializables(e.NativeTokens))
			}
		}).
		WriteNum(OutputExtended, func(err error) error {
			return fmt.Errorf("unable to serialize extended output type ID: %w", err)
		}).
		WriteNum(e.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize extended output amount: %w", err)
		}).
		WriteSliceOfObjects(e.NativeTokens, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, nativeTokensWrittenConsumer, func(err error) error {
			return fmt.Errorf("unable to serialize extended output native tokens: %w", err)
		}).
		WriteObject(e.Address, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize extended output address: %w", err)
		}).
		WriteSliceOfObjects(e.Blocks, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, nil, func(err error) error {
			return fmt.Errorf("unable to serialize extended output feature blocks: %w", err)
		}).
		Serialize()
}

func (e *ExtendedOutput) MarshalJSON() ([]byte, error) {
	var err error
	jExtendedOutput := &jsonExtendedOutput{
		Type:   int(OutputExtended),
		Amount: int(e.Amount),
	}

	jExtendedOutput.Address, err = addressToJSONRawMsg(e.Address)
	if err != nil {
		return nil, err
	}

	jExtendedOutput.NativeTokens, err = serializablesToJSONRawMsgs(e.NativeTokens)
	if err != nil {
		return nil, err
	}

	jExtendedOutput.Blocks, err = serializablesToJSONRawMsgs(e.Blocks)
	if err != nil {
		return nil, err
	}

	return json.Marshal(jExtendedOutput)
}

func (e *ExtendedOutput) UnmarshalJSON(bytes []byte) error {
	jExtendedOutput := &jsonExtendedOutput{}
	if err := json.Unmarshal(bytes, jExtendedOutput); err != nil {
		return err
	}
	seri, err := jExtendedOutput.ToSerializable()
	if err != nil {
		return err
	}
	*e = *seri.(*ExtendedOutput)
	return nil
}

// jsonExtendedOutput defines the json representation of a ExtendedOutput.
type jsonExtendedOutput struct {
	Type         int                `json:"type"`
	Amount       int                `json:"amount"`
	NativeTokens []*json.RawMessage `json:"nativeTokens"`
	Address      *json.RawMessage   `json:"address"`
	Blocks       []*json.RawMessage `json:"blocks"`
}

func (j *jsonExtendedOutput) ToSerializable() (serializer.Serializable, error) {
	var err error
	e := &ExtendedOutput{
		Amount: uint64(j.Amount),
	}

	e.Address, err = addressFromJSONRawMsg(j.Address)
	if err != nil {
		return nil, err
	}

	e.NativeTokens, err = jsonRawMsgsToSerializables(j.NativeTokens, func(ty int) (JSONSerializable, error) {
		return &jsonNativeToken{}, nil
	})
	if err != nil {
		return nil, err
	}

	e.Blocks, err = jsonRawMsgsToSerializables(j.Blocks, jsonFeatureBlockSelector)
	if err != nil {
		return nil, err
	}

	return e, nil
}
