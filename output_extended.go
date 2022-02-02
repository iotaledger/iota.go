package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	extOutputUnlockCondsArrayRules = &serializer.ArrayRules{
		Min: 1, Max: 4,
		MustOccur: serializer.TypePrefixes{
			uint32(UnlockConditionAddress): struct{}{},
		},
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(UnlockConditionAddress):
				case uint32(UnlockConditionDustDepositReturn):
				case uint32(UnlockConditionTimelock):
				case uint32(UnlockConditionExpiration):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize extended output, unsupported unlock condition type %s", ErrUnsupportedUnlockConditionType, UnlockConditionType(ty))
				}
				return UnlockConditionSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *AddressUnlockCondition:
				case *DustDepositReturnUnlockCondition:
				case *TimelockUnlockCondition:
				case *ExpirationUnlockCondition:
				default:
					return fmt.Errorf("%w: in extended output", ErrUnsupportedUnlockConditionType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	extOutputFeatBlockArrayRules = &serializer.ArrayRules{
		Min: 0,
		Max: 8,
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(FeatureBlockSender):
				case uint32(FeatureBlockMetadata):
				case uint32(FeatureBlockTag):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize extended output, unsupported feature block type %s", ErrUnsupportedFeatureBlockType, FeatureBlockType(ty))
				}
				return FeatureBlockSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *SenderFeatureBlock:
				case *MetadataFeatureBlock:
				case *TagFeatureBlock:
				default:
					return fmt.Errorf("%w: in extended output", ErrUnsupportedFeatureBlockType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}
)

// ExtendedOutputUnlockConditionsArrayRules returns array rules defining the constraints on UnlockConditions within an ExtendedOutput.
func ExtendedOutputUnlockConditionsArrayRules() serializer.ArrayRules {
	return *extOutputUnlockCondsArrayRules
}

// ExtendedOutputFeatureBlocksArrayRules returns array rules defining the constraints on FeatureBlocks within an ExtendedOutput.
func ExtendedOutputFeatureBlocksArrayRules() serializer.ArrayRules {
	return *extOutputFeatBlockArrayRules
}

// ExtendedOutputs is a slice of ExtendedOutput(s).
type ExtendedOutputs []*ExtendedOutput

// ExtendedOutput is an output type which can hold native tokens and feature blocks.
type ExtendedOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64
	// The native tokens held by the output.
	NativeTokens NativeTokens
	// The unlock conditions on this output.
	Conditions UnlockConditions
	// The feature blocks which extending the output metadata.
	Blocks FeatureBlocks
}

func (e *ExtendedOutput) Clone() Output {
	return &ExtendedOutput{
		Amount:       e.Amount,
		NativeTokens: e.NativeTokens.Clone(),
		Conditions:   e.Conditions.Clone(),
		Blocks:       e.Blocks.Clone(),
	}
}

func (e *ExtendedOutput) UnlockableBy(ident Address, extParas *ExternalUnlockParameters) bool {
	ok, _ := outputUnlockable(e, nil, ident, extParas)
	return ok
}

func (e *ExtendedOutput) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return outputOffsetVByteCost(costStruct) +
		// prefix + amount
		costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		e.NativeTokens.VByteCost(costStruct, nil) +
		e.Conditions.VByteCost(costStruct, nil) +
		e.Blocks.VByteCost(costStruct, nil)
}

func (e *ExtendedOutput) NativeTokenSet() NativeTokens {
	return e.NativeTokens
}

func (e *ExtendedOutput) FeatureBlocks() FeatureBlocks {
	return e.Blocks
}

func (e *ExtendedOutput) UnlockConditions() UnlockConditions {
	return e.Conditions
}

func (e *ExtendedOutput) Deposit() uint64 {
	return e.Amount
}

func (e *ExtendedOutput) Ident() Address {
	return e.Conditions.MustSet().Address().Address
}

func (e *ExtendedOutput) Type() OutputType {
	return OutputExtended
}

func (e *ExtendedOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(OutputExtended), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize extended output: %w", err)
		}).
		ReadNum(&e.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for extended output: %w", err)
		}).
		ReadSliceOfObjects(&e.NativeTokens, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationNone, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize native tokens for extended output: %w", err)
		}).
		ReadSliceOfObjects(&e.Conditions, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, extOutputUnlockCondsArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize unlock conditions for extended output: %w", err)
		}).
		ReadSliceOfObjects(&e.Blocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, extOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize feature blocks for extended output: %w", err)
		}).
		Done()
}

func (e *ExtendedOutput) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(OutputExtended, func(err error) error {
			return fmt.Errorf("unable to serialize extended output type ID: %w", err)
		}).
		WriteNum(e.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize extended output amount: %w", err)
		}).
		WriteSliceOfObjects(&e.NativeTokens, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize extended output native tokens: %w", err)
		}).
		WriteSliceOfObjects(&e.Conditions, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, extOutputUnlockCondsArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize extended output unlock conditions: %w", err)
		}).
		WriteSliceOfObjects(&e.Blocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, extOutputFeatBlockArrayRules, func(err error) error {
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

	jExtendedOutput.NativeTokens, err = serializablesToJSONRawMsgs(e.NativeTokens.ToSerializables())
	if err != nil {
		return nil, err
	}

	jExtendedOutput.Conditions, err = serializablesToJSONRawMsgs(e.Conditions.ToSerializables())
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
	Conditions   []*json.RawMessage `json:"unlockConditions"`
	Blocks       []*json.RawMessage `json:"featureBlocks"`
}

func (j *jsonExtendedOutput) ToSerializable() (serializer.Serializable, error) {
	var err error
	e := &ExtendedOutput{
		Amount: uint64(j.Amount),
	}

	e.NativeTokens, err = nativeTokensFromJSONRawMsg(j.NativeTokens)
	if err != nil {
		return nil, err
	}

	e.Conditions, err = unlockConditionsFromJSONRawMsg(j.Conditions)
	if err != nil {
		return nil, err
	}

	e.Blocks, err = featureBlocksFromJSONRawMsg(j.Blocks)
	if err != nil {
		return nil, err
	}

	return e, nil
}
