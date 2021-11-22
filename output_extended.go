package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

var (
	extendedOutputAddrGuard = serializer.SerializableGuard{
		ReadGuard:  addrReadGuard(allAddressTypeSet),
		WriteGuard: addrWriteGuard(allAddressTypeSet),
	}

	extendedOutputFeatBlockArrayRules = &serializer.ArrayRules{
		Min: 0,
		Max: 8,
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(FeatureBlockSender):
				case uint32(FeatureBlockDustDepositReturn):
				case uint32(FeatureBlockTimelockMilestoneIndex):
				case uint32(FeatureBlockTimelockUnix):
				case uint32(FeatureBlockExpirationMilestoneIndex):
				case uint32(FeatureBlockExpirationUnix):
				case uint32(FeatureBlockMetadata):
				case uint32(FeatureBlockIndexation):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize extended output, unsupported feature block type %s", ErrUnsupportedFeatureBlockType, FeatureBlockTypeToString(FeatureBlockType(ty)))
				}
				return FeatureBlockSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *SenderFeatureBlock:
				case *DustDepositReturnFeatureBlock:
				case *TimelockMilestoneIndexFeatureBlock:
				case *TimelockUnixFeatureBlock:
				case *ExpirationMilestoneIndexFeatureBlock:
				case *ExpirationUnixFeatureBlock:
				case *MetadataFeatureBlock:
				case *IndexationFeatureBlock:
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

// ExtendedOutputFeatureBlocksArrayRules returns array rules defining the constraints on FeatureBlocks within an ExtendedOutput.
func ExtendedOutputFeatureBlocksArrayRules() serializer.ArrayRules {
	return *extendedOutputFeatBlockArrayRules
}

// ExtendedOutput is an output type which can hold native tokens and feature blocks.
type ExtendedOutput struct {
	// The deposit address.
	Address Address
	// The amount of IOTA tokens held by the output.
	Amount uint64
	// The native tokens held by the output.
	NativeTokens NativeTokens
	// The feature blocks which modulate the constraints on the output.
	Blocks FeatureBlocks
}

func (e *ExtendedOutput) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return costStruct.VBFactorKey.Multiply(OutputIDLength) +
		costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		e.NativeTokens.VByteCost(costStruct, nil) +
		e.Address.VByteCost(costStruct, nil) +
		e.Blocks.VByteCost(costStruct, nil)
}

func (e *ExtendedOutput) NativeTokenSet() NativeTokens {
	return e.NativeTokens
}

func (e *ExtendedOutput) FeatureBlocks() FeatureBlocks {
	return e.Blocks
}

func (e *ExtendedOutput) Deposit() uint64 {
	return e.Amount
}

func (e *ExtendedOutput) Ident() (Address, error) {
	return e.Address, nil
}

func (e *ExtendedOutput) Type() OutputType {
	return OutputExtended
}

func (e *ExtendedOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(OutputExtended), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize extended output: %w", err)
		}).
		ReadObject(&e.Address, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, extendedOutputAddrGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("unable to deserialize address for extended output: %w", err)
		}).
		ReadNum(&e.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for extended output: %w", err)
		}).
		ReadSliceOfObjects(&e.NativeTokens, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsUint16, serializer.TypeDenotationNone, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize native tokens for extended output: %w", err)
		}).
		ReadSliceOfObjects(&e.Blocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, extendedOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize feature blocks for NFT output: %w", err)
		}).
		Done()
}

func (e *ExtendedOutput) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(OutputExtended, func(err error) error {
			return fmt.Errorf("unable to serialize extended output type ID: %w", err)
		}).
		WriteObject(e.Address, deSeriMode, deSeriCtx, extendedOutputAddrGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize extended output address: %w", err)
		}).
		WriteNum(e.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize extended output amount: %w", err)
		}).
		WriteSliceOfObjects(&e.NativeTokens, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsUint16, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize extended output native tokens: %w", err)
		}).
		WriteSliceOfObjects(&e.Blocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, extendedOutputFeatBlockArrayRules, func(err error) error {
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
