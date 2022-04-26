package iotago

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

const (
	// FoundryIDLength is the byte length of a FoundryID consisting out of the alias address, serial number and token scheme.
	FoundryIDLength = AliasAddressSerializedBytesSize + serializer.UInt32ByteSize + serializer.OneByte
)

var (
	// ErrNonUniqueFoundryOutputs gets returned when multiple FoundryOutput(s) with the same FoundryID exist within an OutputsByType.
	ErrNonUniqueFoundryOutputs = errors.New("non unique foundries within outputs")

	emptyFoundryID = [FoundryIDLength]byte{}

	foundryOutputUnlockCondsArrayRules = &serializer.ArrayRules{
		Min: 1, Max: 1,
		MustOccur: serializer.TypePrefixes{
			uint32(UnlockConditionImmutableAlias): struct{}{},
		},
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(UnlockConditionImmutableAlias):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize foundry output, unsupported unlock condition type %s", ErrUnsupportedUnlockConditionType, UnlockConditionType(ty))
				}
				return UnlockConditionSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *ImmutableAliasUnlockCondition:
				default:
					return fmt.Errorf("%w: in foundry output", ErrUnsupportedUnlockConditionType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	foundryOutputFeatBlockArrayRules = &serializer.ArrayRules{
		Min: 0,
		Max: 1,
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(FeatureBlockMetadata):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize foundry output, unsupported feature block type %s", ErrUnsupportedFeatureBlockType, FeatureBlockType(ty))
				}
				return FeatureBlockSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *MetadataFeatureBlock:
				default:
					return fmt.Errorf("%w: in foundry output", ErrUnsupportedFeatureBlockType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	foundryOutputImmFeatBlockArrayRules = &serializer.ArrayRules{
		Min: 0,
		Max: 1,
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(FeatureBlockMetadata):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize foundry output, unsupported immutable feature block type %s", ErrUnsupportedFeatureBlockType, FeatureBlockType(ty))
				}
				return FeatureBlockSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *MetadataFeatureBlock:
				default:
					return fmt.Errorf("%w: in foundry output", ErrUnsupportedFeatureBlockType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}
)

// FoundryOutputFeatureBlocksArrayRules returns array rules defining the constraints on FeatureBlocks within an FoundryOutput.
func FoundryOutputFeatureBlocksArrayRules() serializer.ArrayRules {
	return *foundryOutputFeatBlockArrayRules
}

// FoundryOutputImmutableFeatureBlocksArrayRules returns array rules defining the constraints on immutable FeatureBlocks within an FoundryOutput.
func FoundryOutputImmutableFeatureBlocksArrayRules() serializer.ArrayRules {
	return *foundryOutputImmFeatBlockArrayRules
}

// TokenTag is a tag holding some additional data which might be interpreted by higher layers.
type TokenTag = [TokenTagLength]byte

// FoundryID defines the identifier for a foundry consisting out of the address, serial number and TokenScheme.
type FoundryID [FoundryIDLength]byte

func (fID FoundryID) Addressable() bool {
	return false
}

// FoundrySerialNumber returns the serial number of the foundry.
func (fID FoundryID) FoundrySerialNumber() uint32 {
	return binary.LittleEndian.Uint32(fID[AliasAddressSerializedBytesSize : AliasAddressSerializedBytesSize+serializer.UInt32ByteSize])
}

func (fID FoundryID) Matches(other ChainID) bool {
	otherFID, is := other.(FoundryID)
	if !is {
		return false
	}
	return fID == otherFID
}

func (fID FoundryID) ToAddress() ChainConstrainedAddress {
	panic("foundry ID is not addressable")
}

func (fID FoundryID) Empty() bool {
	return fID == emptyFoundryID
}

func (fID FoundryID) Key() interface{} {
	return fID.String()
}

func (fID FoundryID) String() string {
	return EncodeHex(fID[:])
}

// FoundryOutputs is a slice of FoundryOutput(s).
type FoundryOutputs []*FoundryOutput

// FoundryOutputsSet is a set of FoundryOutput(s).
type FoundryOutputsSet map[FoundryID]*FoundryOutput

// FoundryOutput is an output type which controls the supply of user defined native tokens.
type FoundryOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64
	// The native tokens held by the output.
	NativeTokens NativeTokens
	// The serial number of the foundry.
	SerialNumber uint32
	// The tag which is always the last 12 bytes of the tokens generated by this foundry.
	TokenTag TokenTag
	// The token scheme this foundry uses.
	TokenScheme TokenScheme
	// The unlock conditions on this output.
	Conditions UnlockConditions
	// The feature blocks on the output.
	Blocks FeatureBlocks
	// The immutable feature blocks on the output.
	ImmutableBlocks FeatureBlocks
}

func (f *FoundryOutput) Clone() Output {
	return &FoundryOutput{
		Amount:          f.Amount,
		NativeTokens:    f.NativeTokens.Clone(),
		SerialNumber:    f.SerialNumber,
		TokenTag:        f.TokenTag,
		TokenScheme:     f.TokenScheme.Clone(),
		Conditions:      f.Conditions.Clone(),
		Blocks:          f.Blocks.Clone(),
		ImmutableBlocks: f.ImmutableBlocks.Clone(),
	}
}

func (f *FoundryOutput) Ident() Address {
	return f.Conditions.MustSet().ImmutableAlias().Address
}

func (f *FoundryOutput) UnlockableBy(ident Address, extParas *ExternalUnlockParameters) bool {
	ok, _ := outputUnlockable(f, nil, ident, extParas)
	return ok
}

func (f *FoundryOutput) VBytes(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return outputOffsetVByteCost(costStruct) +
		// prefix + amount
		costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		f.NativeTokens.VBytes(costStruct, nil) +
		// serial number, token tag
		costStruct.VBFactorData.Multiply(serializer.UInt32ByteSize+TokenTagLength) +
		f.TokenScheme.VByte(costStruct, nil) +
		f.Conditions.VByte(costStruct, nil) +
		f.Blocks.VByte(costStruct, nil) +
		f.ImmutableBlocks.VByte(costStruct, nil)
}

func (f *FoundryOutput) Chain() ChainID {
	foundryID, err := f.ID()
	if err != nil {
		panic(err)
	}
	return foundryID
}

func (f *FoundryOutput) ValidateStateTransition(transType ChainTransitionType, next ChainConstrainedOutput, semValCtx *SemanticValidationContext) error {
	inSums := semValCtx.WorkingSet.InNativeTokens
	outSums := semValCtx.WorkingSet.OutNativeTokens

	var err error
	switch transType {
	case ChainTransitionTypeGenesis:
		err = f.genesisValid(semValCtx, f.MustID(), outSums)
	case ChainTransitionTypeStateChange:
		err = f.stateChangeValid(next, inSums, outSums)
	case ChainTransitionTypeDestroy:
		err = f.destructionValid(inSums, outSums)
	default:
		panic("unknown chain transition type in FoundryOutput")
	}
	if err != nil {
		return &ChainTransitionError{Inner: err, Msg: fmt.Sprintf("foundry %s, token %s", f.MustID(), f.MustNativeTokenID())}
	}
	return nil
}

func (f *FoundryOutput) genesisValid(semValCtx *SemanticValidationContext, thisFoundryID FoundryID, outSums NativeTokenSum) error {

	nativeTokenID := f.MustNativeTokenID()
	if err := f.TokenScheme.StateTransition(ChainTransitionTypeGenesis, nil, nil, outSums.ValueOrBigInt0(nativeTokenID)); err != nil {
		return err
	}

	// grab foundry counter from transitioning AliasOutput
	aliasID := f.Ident().(*AliasAddress).AliasID()
	inAlias, ok := semValCtx.WorkingSet.InChains[aliasID]
	if !ok {
		return fmt.Errorf("missing input transitioning alias output %s for new foundry output %s", aliasID, thisFoundryID)
	}

	outAlias, ok := semValCtx.WorkingSet.OutChains[aliasID]
	if !ok {
		return fmt.Errorf("missing output transitioning alias output %s for new foundry output %s", aliasID, thisFoundryID)
	}

	if err := f.validSerialNumber(semValCtx, inAlias.(*AliasOutput), outAlias.(*AliasOutput), thisFoundryID); err != nil {
		return err
	}

	return nil
}

func (f *FoundryOutput) validSerialNumber(semValCtx *SemanticValidationContext, inAlias *AliasOutput, outAlias *AliasOutput, thisFoundryID FoundryID) error {
	// this new foundry's serial number must be between the given foundry counter interval
	startSerial := inAlias.FoundryCounter
	endIncSerial := outAlias.FoundryCounter
	if startSerial >= f.SerialNumber || f.SerialNumber > endIncSerial {
		return fmt.Errorf("new foundry output %s's serial number is not between the foundry counter interval of [%d,%d)", thisFoundryID, startSerial, endIncSerial)
	}

	// OPTIMIZE: this loop happens on every STVF of every new foundry output
	// check order of serial number
	for outputIndex, output := range semValCtx.WorkingSet.Tx.Essence.Outputs {
		otherFoundryOutput, is := output.(*FoundryOutput)
		if !is {
			continue
		}

		if !otherFoundryOutput.Ident().Equal(f.Ident()) {
			continue
		}

		otherFoundryID, err := otherFoundryOutput.ID()
		if err != nil {
			return err
		}

		if _, isNotNew := semValCtx.WorkingSet.InChains[otherFoundryID]; isNotNew {
			continue
		}

		// only check up to own foundry whether it is ordered
		if otherFoundryID == thisFoundryID {
			break
		}

		if otherFoundryOutput.SerialNumber >= f.SerialNumber {
			return fmt.Errorf("new foundry output %s at index %d has bigger equal serial number than this foundry %s", otherFoundryID, outputIndex, thisFoundryID)
		}
	}
	return nil
}

func (f *FoundryOutput) stateChangeValid(next ChainConstrainedOutput, inSums NativeTokenSum, outSums NativeTokenSum) error {
	nextState, is := next.(*FoundryOutput)
	if !is {
		return fmt.Errorf("foundry output can only state transition to another foundry output")
	}

	if !f.ImmutableBlocks.Equal(nextState.ImmutableBlocks) {
		return fmt.Errorf("old state %s, next state %s", f.ImmutableBlocks, nextState.ImmutableBlocks)
	}

	// the check for the serial number and token scheme not being mutated is implicit
	// as a change would cause the foundry ID to be different, which would result in
	// no matching foundry to be found to validate the state transition against
	switch {
	case f.MustID() != nextState.MustID():
		// impossible invariant as the STVF should be called via the matching next foundry output
		panic(fmt.Sprintf("foundry IDs mismatch in state transition validation function: have %v got %v", f.MustID(), nextState.MustID()))
	case f.TokenTag != nextState.TokenTag:
		return fmt.Errorf("token tag mismatch wanted %s but got %s", f.TokenTag, nextState.TokenTag)
	}

	nativeTokenID := f.MustNativeTokenID()
	if err := f.TokenScheme.StateTransition(ChainTransitionTypeStateChange, nextState.TokenScheme, inSums.ValueOrBigInt0(nativeTokenID), outSums.ValueOrBigInt0(nativeTokenID)); err != nil {
		return err
	}

	return nil
}

func (f *FoundryOutput) destructionValid(inSums NativeTokenSum, outSums NativeTokenSum) error {
	nativeTokenID := f.MustNativeTokenID()
	if err := f.TokenScheme.StateTransition(ChainTransitionTypeDestroy, nil, inSums.ValueOrBigInt0(nativeTokenID), outSums.ValueOrBigInt0(nativeTokenID)); err != nil {
		return err
	}
	return nil
}

// ID returns the FoundryID of this FoundryOutput.
func (f *FoundryOutput) ID() (FoundryID, error) {
	var foundryID FoundryID
	addrBytes, err := f.Ident().Serialize(serializer.DeSeriModeNoValidation, nil)
	if err != nil {
		return foundryID, err
	}
	copy(foundryID[:], addrBytes)
	binary.LittleEndian.PutUint32(foundryID[len(addrBytes):], f.SerialNumber)
	foundryID[len(foundryID)-1] = byte(f.TokenScheme.Type())
	return foundryID, nil
}

// MustID works like ID but panics if an error occurs.
func (f *FoundryOutput) MustID() FoundryID {
	id, err := f.ID()
	if err != nil {
		panic(err)
	}
	return id
}

// MustNativeTokenID works like NativeTokenID but panics if there is an error.
func (f *FoundryOutput) MustNativeTokenID() NativeTokenID {
	nativeTokenID, err := f.NativeTokenID()
	if err != nil {
		panic(err)
	}
	return nativeTokenID
}

// NativeTokenID returns the NativeTokenID this FoundryOutput operates on.
func (f *FoundryOutput) NativeTokenID() (NativeTokenID, error) {
	var nativeTokenID NativeTokenID
	foundryID, err := f.ID()
	if err != nil {
		return nativeTokenID, err
	}
	copy(nativeTokenID[:], foundryID[:])
	copy(nativeTokenID[len(foundryID):], f.TokenTag[:])
	return nativeTokenID, nil
}

func (f *FoundryOutput) NativeTokenSet() NativeTokens {
	return f.NativeTokens
}

func (f *FoundryOutput) UnlockConditions() UnlockConditions {
	return f.Conditions
}

func (f *FoundryOutput) FeatureBlocks() FeatureBlocks {
	return f.Blocks
}

func (f *FoundryOutput) ImmutableFeatureBlocks() FeatureBlocks {
	return f.ImmutableBlocks
}

func (f *FoundryOutput) Deposit() uint64 {
	return f.Amount
}

func (f *FoundryOutput) Type() OutputType {
	return OutputFoundry
}

func (f *FoundryOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(OutputFoundry), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize foundry output: %w", err)
		}).
		ReadNum(&f.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for foundry output: %w", err)
		}).
		ReadSliceOfObjects(&f.NativeTokens, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationNone, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize native tokens for foundry output: %w", err)
		}).
		ReadNum(&f.SerialNumber, func(err error) error {
			return fmt.Errorf("unable to deserialize serial number for foundry output: %w", err)
		}).
		ReadArrayOf12Bytes(&f.TokenTag, func(err error) error {
			return fmt.Errorf("unable to deserialize token tag for foundry output: %w", err)
		}).
		ReadObject(&f.TokenScheme, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, wrappedTokenSchemeSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize token scheme for foundry output: %w", err)
		}).
		ReadSliceOfObjects(&f.Conditions, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, foundryOutputUnlockCondsArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize unlock conditions for foundry output: %w", err)
		}).
		ReadSliceOfObjects(&f.Blocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, foundryOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize feature blocks for foundry output: %w", err)
		}).
		ReadSliceOfObjects(&f.ImmutableBlocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, foundryOutputImmFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize immutable feature blocks for foundry output: %w", err)
		}).
		Done()
}

func (f *FoundryOutput) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(OutputFoundry), func(err error) error {
			return fmt.Errorf("unable to serialize foundry output type ID: %w", err)
		}).
		WriteNum(f.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output amount: %w", err)
		}).
		WriteSliceOfObjects(&f.NativeTokens, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output native tokens: %w", err)
		}).
		WriteNum(f.SerialNumber, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output serial number: %w", err)
		}).
		WriteBytes(f.TokenTag[:], func(err error) error {
			return fmt.Errorf("unable to serialize foundry output token tag: %w", err)
		}).
		WriteObject(f.TokenScheme, deSeriMode, deSeriCtx, tokenSchemeWriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output token scheme: %w", err)
		}).
		WriteSliceOfObjects(&f.Conditions, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, foundryOutputUnlockCondsArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output unlock conditions: %w", err)
		}).
		WriteSliceOfObjects(&f.Blocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, foundryOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output feature blocks: %w", err)
		}).
		WriteSliceOfObjects(&f.ImmutableBlocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, foundryOutputImmFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output immutable feature blocks: %w", err)
		}).
		Serialize()
}

func (f *FoundryOutput) Size() int {
	return util.NumByteLen(byte(OutputFoundry)) +
		util.NumByteLen(f.Amount) +
		f.NativeTokens.Size() +
		util.NumByteLen(f.SerialNumber) +
		TokenTagLength +
		f.TokenScheme.Size() +
		f.Conditions.Size() +
		f.Blocks.Size() +
		f.ImmutableBlocks.Size()
}

func (f *FoundryOutput) MarshalJSON() ([]byte, error) {
	var err error
	jFoundryOutput := &jsonFoundryOutput{
		Type:         int(OutputFoundry),
		Amount:       EncodeUint64(f.Amount),
		SerialNumber: int(f.SerialNumber),
	}

	jFoundryOutput.NativeTokens, err = serializablesToJSONRawMsgs(f.NativeTokens.ToSerializables())
	if err != nil {
		return nil, err
	}

	jFoundryOutput.TokenTag = EncodeHex(f.TokenTag[:])

	jTokenSchemeBytes, err := f.TokenScheme.MarshalJSON()
	if err != nil {
		return nil, err
	}
	jsonRawMsgTokenScheme := json.RawMessage(jTokenSchemeBytes)
	jFoundryOutput.TokenScheme = &jsonRawMsgTokenScheme

	jFoundryOutput.Conditions, err = serializablesToJSONRawMsgs(f.Conditions.ToSerializables())
	if err != nil {
		return nil, err
	}

	jFoundryOutput.Blocks, err = serializablesToJSONRawMsgs(f.Blocks.ToSerializables())
	if err != nil {
		return nil, err
	}

	jFoundryOutput.ImmutableBlocks, err = serializablesToJSONRawMsgs(f.ImmutableBlocks.ToSerializables())
	if err != nil {
		return nil, err
	}

	return json.Marshal(jFoundryOutput)
}

func (f *FoundryOutput) UnmarshalJSON(bytes []byte) error {
	jFoundryOutput := &jsonFoundryOutput{}
	if err := json.Unmarshal(bytes, jFoundryOutput); err != nil {
		return err
	}
	seri, err := jFoundryOutput.ToSerializable()
	if err != nil {
		return err
	}
	*f = *seri.(*FoundryOutput)
	return nil
}

// jsonFoundryOutput defines the json representation of a FoundryOutput.
type jsonFoundryOutput struct {
	Type            int                `json:"type"`
	Amount          string             `json:"amount"`
	NativeTokens    []*json.RawMessage `json:"nativeTokens"`
	SerialNumber    int                `json:"serialNumber"`
	TokenTag        string             `json:"tokenTag"`
	TokenScheme     *json.RawMessage   `json:"tokenScheme"`
	Conditions      []*json.RawMessage `json:"unlockConditions"`
	Blocks          []*json.RawMessage `json:"featureBlocks"`
	ImmutableBlocks []*json.RawMessage `json:"immutableFeatureBlocks"`
}

func (j *jsonFoundryOutput) ToSerializable() (serializer.Serializable, error) {
	var err error
	e := &FoundryOutput{
		SerialNumber: uint32(j.SerialNumber),
	}

	e.Amount, err = DecodeUint64(j.Amount)
	if err != nil {
		return nil, err
	}

	e.NativeTokens, err = nativeTokensFromJSONRawMsg(j.NativeTokens)
	if err != nil {
		return nil, err
	}

	tokenTagBytes, err := DecodeHex(j.TokenTag)
	if err != nil {
		return nil, err
	}
	copy(e.TokenTag[:], tokenTagBytes)

	e.TokenScheme, err = tokenSchemeFromJSONRawMsg(j.TokenScheme)
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

	e.ImmutableBlocks, err = featureBlocksFromJSONRawMsg(j.ImmutableBlocks)
	if err != nil {
		return nil, err
	}

	return e, nil
}
