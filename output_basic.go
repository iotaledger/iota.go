package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

var (
	basicOutputUnlockCondsArrayRules = &serializer.ArrayRules{
		Min: 1, Max: 4,
		MustOccur: serializer.TypePrefixes{
			uint32(UnlockConditionAddress): struct{}{},
		},
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(UnlockConditionAddress):
				case uint32(UnlockConditionStorageDepositReturn):
				case uint32(UnlockConditionTimelock):
				case uint32(UnlockConditionExpiration):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize basic output, unsupported unlock condition type %s", ErrUnsupportedUnlockConditionType, UnlockConditionType(ty))
				}
				return UnlockConditionSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *AddressUnlockCondition:
				case *StorageDepositReturnUnlockCondition:
				case *TimelockUnlockCondition:
				case *ExpirationUnlockCondition:
				default:
					return fmt.Errorf("%w: in basic output", ErrUnsupportedUnlockConditionType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	basicOutputFeatBlockArrayRules = &serializer.ArrayRules{
		Min: 0,
		Max: 8,
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(FeatureBlockSender):
				case uint32(FeatureBlockMetadata):
				case uint32(FeatureBlockTag):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize basic output, unsupported feature block type %s", ErrUnsupportedFeatureBlockType, FeatureBlockType(ty))
				}
				return FeatureBlockSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *SenderFeatureBlock:
				case *MetadataFeatureBlock:
				case *TagFeatureBlock:
				default:
					return fmt.Errorf("%w: in basic output", ErrUnsupportedFeatureBlockType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}
)

// BasicOutputUnlockConditionsArrayRules returns array rules defining the constraints on UnlockConditions within an BasicOutput.
func BasicOutputUnlockConditionsArrayRules() serializer.ArrayRules {
	return *basicOutputUnlockCondsArrayRules
}

// BasicOutputFeatureBlocksArrayRules returns array rules defining the constraints on FeatureBlocks within an BasicOutput.
func BasicOutputFeatureBlocksArrayRules() serializer.ArrayRules {
	return *basicOutputFeatBlockArrayRules
}

// BasicOutputs is a slice of BasicOutput(s).
type BasicOutputs []*BasicOutput

// BasicOutput is an output type which can hold native tokens and feature blocks.
type BasicOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64
	// The native tokens held by the output.
	NativeTokens NativeTokens
	// The unlock conditions on this output.
	Conditions UnlockConditions
	// The feature blocks which extending the output metadata.
	Blocks FeatureBlocks
}

func (e *BasicOutput) Clone() Output {
	return &BasicOutput{
		Amount:       e.Amount,
		NativeTokens: e.NativeTokens.Clone(),
		Conditions:   e.Conditions.Clone(),
		Blocks:       e.Blocks.Clone(),
	}
}

func (e *BasicOutput) UnlockableBy(ident Address, extParas *ExternalUnlockParameters) bool {
	ok, _ := outputUnlockable(e, nil, ident, extParas)
	return ok
}

func (e *BasicOutput) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return outputOffsetVByteCost(costStruct) +
		// prefix + amount
		costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		e.NativeTokens.VByteCost(costStruct, nil) +
		e.Conditions.VByteCost(costStruct, nil) +
		e.Blocks.VByteCost(costStruct, nil)
}

func (e *BasicOutput) NativeTokenSet() NativeTokens {
	return e.NativeTokens
}

func (e *BasicOutput) FeatureBlocks() FeatureBlocks {
	return e.Blocks
}

func (e *BasicOutput) UnlockConditions() UnlockConditions {
	return e.Conditions
}

func (e *BasicOutput) Deposit() uint64 {
	return e.Amount
}

func (e *BasicOutput) Ident() Address {
	return e.Conditions.MustSet().Address().Address
}

func (e *BasicOutput) Type() OutputType {
	return OutputBasic
}

func (e *BasicOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(OutputBasic), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize basic output: %w", err)
		}).
		ReadNum(&e.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for basic output: %w", err)
		}).
		ReadSliceOfObjects(&e.NativeTokens, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationNone, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize native tokens for basic output: %w", err)
		}).
		ReadSliceOfObjects(&e.Conditions, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, basicOutputUnlockCondsArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize unlock conditions for basic output: %w", err)
		}).
		ReadSliceOfObjects(&e.Blocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, basicOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize feature blocks for basic output: %w", err)
		}).
		Done()
}

func (e *BasicOutput) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(OutputBasic), func(err error) error {
			return fmt.Errorf("unable to serialize basic output type ID: %w", err)
		}).
		WriteNum(e.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize basic output amount: %w", err)
		}).
		WriteSliceOfObjects(&e.NativeTokens, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize basic output native tokens: %w", err)
		}).
		WriteSliceOfObjects(&e.Conditions, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, basicOutputUnlockCondsArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize basic output unlock conditions: %w", err)
		}).
		WriteSliceOfObjects(&e.Blocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, basicOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize basic output feature blocks: %w", err)
		}).
		Serialize()
}

func (e *BasicOutput) Size() int {
	return util.NumByteLen(byte(OutputBasic)) +
		util.NumByteLen(e.Amount) +
		e.NativeTokens.Size() +
		e.Conditions.Size() +
		e.Blocks.Size()
}

func (e *BasicOutput) MarshalJSON() ([]byte, error) {
	var err error
	jExtendedOutput := &jsonExtendedOutput{
		Type:   int(OutputBasic),
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

func (e *BasicOutput) UnmarshalJSON(bytes []byte) error {
	jExtendedOutput := &jsonExtendedOutput{}
	if err := json.Unmarshal(bytes, jExtendedOutput); err != nil {
		return err
	}
	seri, err := jExtendedOutput.ToSerializable()
	if err != nil {
		return err
	}
	*e = *seri.(*BasicOutput)
	return nil
}

// jsonExtendedOutput defines the json representation of a BasicOutput.
type jsonExtendedOutput struct {
	Type         int                `json:"type"`
	Amount       int                `json:"amount"`
	NativeTokens []*json.RawMessage `json:"nativeTokens"`
	Conditions   []*json.RawMessage `json:"unlockConditions"`
	Blocks       []*json.RawMessage `json:"featureBlocks"`
}

func (j *jsonExtendedOutput) ToSerializable() (serializer.Serializable, error) {
	var err error
	e := &BasicOutput{
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
