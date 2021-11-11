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
	NativeTokens NativeTokens
	// The deposit address.
	Address Address
	// The feature blocks which modulate the constraints on the output.
	Blocks FeatureBlocks
}

func (e *ExtendedOutput) NativeTokenSet() NativeTokens {
	return e.NativeTokens
}

func (e *ExtendedOutput) FeatureBlocks() FeatureBlocks {
	return e.Blocks
}

func (e *ExtendedOutput) Deposit() (uint64, error) {
	return e.Amount, nil
}

func (e *ExtendedOutput) Ident() (Address, error) {
	return e.Address, nil
}

func (e *ExtendedOutput) Type() OutputType {
	return OutputExtended
}

func (e *ExtendedOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(OutputExtended), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize extended output: %w", err)
		}).
		ReadNum(&e.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for extended output: %w", err)
		}).
		ReadSliceOfObjects(&e.NativeTokens, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationNone, func(ty uint32) (serializer.Serializable, error) {
			return &NativeToken{}, nil
		}, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize native tokens for extended output: %w", err)
		}).
		ReadObject(&e.Address, deSeriMode, serializer.TypeDenotationByte, AddressSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize address for extended output: %w", err)
		}).
		ReadSliceOfObjects(&e.Blocks, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationByte, extendedOutputFeatureBlocksGuard, featBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize feature blocks for NFT output: %w", err)
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

func extendedOutputFeatureBlocksGuard(ty uint32) (serializer.Serializable, error) {
	if !featureBlocksSupportedByExtendedOutput(ty) {
		return nil, fmt.Errorf("%w: unable to deserialize extended output, unsupported feature block type %s", ErrUnsupportedFeatureBlockType, FeatureBlockTypeToString(FeatureBlockType(ty)))
	}
	return FeatureBlockSelector(ty)
}

func featureBlocksSupportedByExtendedOutput(ty uint32) bool {
	switch ty {
	case uint32(FeatureBlockSender):
	case uint32(FeatureBlockReturn):
	case uint32(FeatureBlockTimelockMilestoneIndex):
	case uint32(FeatureBlockTimelockUnix):
	case uint32(FeatureBlockExpirationMilestoneIndex):
	case uint32(FeatureBlockExpirationUnix):
	case uint32(FeatureBlockMetadata):
	default:
		return false
	}
	return true
}

func (e *ExtendedOutput) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := outputAmountValidator(-1, e); err != nil {
					return fmt.Errorf("%w: unable to serialize extended output", err)
				}

				if err := isValidAddrType(e.Address); err != nil {
					return fmt.Errorf("invalid address set in extended output: %w", err)
				}

				if err := featureBlockSupported(e.FeatureBlocks(), featureBlocksSupportedByExtendedOutput); err != nil {
					return fmt.Errorf("invalid feature blocks set in extended output: %w", err)
				}
			}
			return nil
		}).
		Do(func() {
			if deSeriMode.HasMode(serializer.DeSeriModePerformLexicalOrdering) {
				seris := e.NativeTokens.ToSerializables()
				sort.Sort(serializer.SortedSerializables(seris))
				e.NativeTokens.FromSerializables(seris)
			}
		}).
		WriteNum(OutputExtended, func(err error) error {
			return fmt.Errorf("unable to serialize extended output type ID: %w", err)
		}).
		WriteNum(e.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize extended output amount: %w", err)
		}).
		WriteSliceOfObjects(&e.NativeTokens, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, nativeTokensArrayRules.ToWrittenObjectConsumer(deSeriMode), func(err error) error {
			return fmt.Errorf("unable to serialize extended output native tokens: %w", err)
		}).
		WriteObject(e.Address, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize extended output address: %w", err)
		}).
		WriteSliceOfObjects(&e.Blocks, deSeriMode, serializer.SeriLengthPrefixTypeAsUint16, featBlockArrayRules.ToWrittenObjectConsumer(deSeriMode), func(err error) error {
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

	jExtendedOutput.NativeTokens, err = serializablesToJSONRawMsgs(e.NativeTokens.ToSerializables())
	if err != nil {
		return nil, err
	}

	jExtendedOutput.Blocks, err = serializablesToJSONRawMsgs(e.Blocks.ToSerializables())
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

	e.NativeTokens, err = nativeTokensFromJSONRawMsg(j.NativeTokens)
	if err != nil {
		return nil, err
	}

	e.Blocks, err = featureBlocksFromJSONRawMsg(j.Blocks)
	if err != nil {
		return nil, err
	}

	return e, nil
}
