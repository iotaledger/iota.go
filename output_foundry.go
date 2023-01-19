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
				case uint32(FeatureMetadata):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize foundry output, unsupported feature type %s", ErrUnsupportedFeatureType, FeatureType(ty))
				}
				return FeatureSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *MetadataFeature:
				default:
					return fmt.Errorf("%w: in foundry output", ErrUnsupportedFeatureType)
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
				case uint32(FeatureMetadata):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize foundry output, unsupported immutable feature type %s", ErrUnsupportedFeatureType, FeatureType(ty))
				}
				return FeatureSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *MetadataFeature:
				default:
					return fmt.Errorf("%w: in foundry output", ErrUnsupportedFeatureType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}
)

// FoundryOutputFeaturesArrayRules returns array rules defining the constraints on Features within an FoundryOutput.
func FoundryOutputFeaturesArrayRules() serializer.ArrayRules {
	return *foundryOutputFeatBlockArrayRules
}

// FoundryOutputImmutableFeaturesArrayRules returns array rules defining the constraints on immutable Features within an FoundryOutput.
func FoundryOutputImmutableFeaturesArrayRules() serializer.ArrayRules {
	return *foundryOutputImmFeatBlockArrayRules
}

// FoundryID defines the identifier for a foundry consisting out of the address, serial number and TokenScheme.
type FoundryID [FoundryIDLength]byte

func (fID FoundryID) ToHex() string {
	return EncodeHex(fID[:])
}

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
	// The token scheme this foundry uses.
	TokenScheme TokenScheme
	// The unlock conditions on this output.
	Conditions UnlockConditions
	// The feature on the output.
	Features Features
	// The immutable feature on the output.
	ImmutableFeatures Features
}

func (f *FoundryOutput) Clone() Output {
	return &FoundryOutput{
		Amount:            f.Amount,
		NativeTokens:      f.NativeTokens.Clone(),
		SerialNumber:      f.SerialNumber,
		TokenScheme:       f.TokenScheme.Clone(),
		Conditions:        f.Conditions.Clone(),
		Features:          f.Features.Clone(),
		ImmutableFeatures: f.ImmutableFeatures.Clone(),
	}
}

func (f *FoundryOutput) Ident() Address {
	return f.Conditions.MustSet().ImmutableAlias().Address
}

func (f *FoundryOutput) UnlockableBy(ident Address, extParas *ExternalUnlockParameters) bool {
	ok, _ := outputUnlockable(f, nil, ident, extParas)
	return ok
}

func (f *FoundryOutput) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return outputOffsetVByteCost(rentStruct) +
		// prefix + amount
		rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		f.NativeTokens.VBytes(rentStruct, nil) +
		// serial number
		rentStruct.VBFactorData.Multiply(serializer.UInt32ByteSize) +
		f.TokenScheme.VBytes(rentStruct, nil) +
		f.Conditions.VBytes(rentStruct, nil) +
		f.Features.VBytes(rentStruct, nil) +
		f.ImmutableFeatures.VBytes(rentStruct, nil)
}

func (f *FoundryOutput) ByteSizeKey() uint64 {
	return outputOffsetByteSizeKey() +
		f.NativeTokens.ByteSizeKey() +
		f.TokenScheme.ByteSizeKey() +
		f.Conditions.ByteSizeKey() +
		f.Features.ByteSizeKey() +
		f.ImmutableFeatures.ByteSizeKey()
}

func (f *FoundryOutput) ByteSizeData() uint64 {
	return outputOffsetByteSizeData() +
		// prefix + amount
		serializer.SmallTypeDenotationByteSize + serializer.UInt64ByteSize +
		f.NativeTokens.ByteSizeData() +
		// serial number
		serializer.UInt32ByteSize +
		f.TokenScheme.ByteSizeData() +
		f.Conditions.ByteSizeData() +
		f.Features.ByteSizeData() +
		f.ImmutableFeatures.ByteSizeData()
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
		return fmt.Errorf("missing input transitioning alias output %s for new foundry output %s", aliasID.ToHex(), thisFoundryID)
	}

	outAlias, ok := semValCtx.WorkingSet.OutChains[aliasID]
	if !ok {
		return fmt.Errorf("missing output transitioning alias output %s for new foundry output %s", aliasID.ToHex(), thisFoundryID)
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
		return fmt.Errorf("new foundry output %s's serial number is not between the foundry counter interval of [%d,%d)", thisFoundryID.ToHex(), startSerial, endIncSerial)
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
			return fmt.Errorf("new foundry output %s at index %d has bigger equal serial number than this foundry %s", otherFoundryID.ToHex(), outputIndex, thisFoundryID.ToHex())
		}
	}
	return nil
}

func (f *FoundryOutput) stateChangeValid(next ChainConstrainedOutput, inSums NativeTokenSum, outSums NativeTokenSum) error {
	nextState, is := next.(*FoundryOutput)
	if !is {
		return errors.New("foundry output can only state transition to another foundry output")
	}

	if !f.ImmutableFeatures.Equal(nextState.ImmutableFeatures) {
		return fmt.Errorf("old state %s, next state %s", f.ImmutableFeatures, nextState.ImmutableFeatures)
	}

	// the check for the serial number and token scheme not being mutated is implicit
	// as a change would cause the foundry ID to be different, which would result in
	// no matching foundry to be found to validate the state transition against
	switch {
	case f.MustID() != nextState.MustID():
		// impossible invariant as the STVF should be called via the matching next foundry output
		panic(fmt.Sprintf("foundry IDs mismatch in state transition validation function: have %v got %v", f.MustID(), nextState.MustID()))
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
	return f.ID()
}

func (f *FoundryOutput) NativeTokenList() NativeTokens {
	return f.NativeTokens
}

func (f *FoundryOutput) FeatureSet() FeatureSet {
	return f.Features.MustSet()
}

func (f *FoundryOutput) UnlockConditionSet() UnlockConditionSet {
	return f.Conditions.MustSet()
}

func (f *FoundryOutput) ImmutableFeatureSet() FeatureSet {
	return f.ImmutableFeatures.MustSet()
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
		ReadObject(&f.TokenScheme, deSeriMode, deSeriCtx, serializer.TypeDenotationByte, wrappedTokenSchemeSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize token scheme for foundry output: %w", err)
		}).
		ReadSliceOfObjects(&f.Conditions, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, foundryOutputUnlockCondsArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize unlock conditions for foundry output: %w", err)
		}).
		ReadSliceOfObjects(&f.Features, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, foundryOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize features for foundry output: %w", err)
		}).
		ReadSliceOfObjects(&f.ImmutableFeatures, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, foundryOutputImmFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize immutable features for foundry output: %w", err)
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
		WriteObject(f.TokenScheme, deSeriMode, deSeriCtx, tokenSchemeWriteGuard, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output token scheme: %w", err)
		}).
		WriteSliceOfObjects(&f.Conditions, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, foundryOutputUnlockCondsArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output unlock conditions: %w", err)
		}).
		WriteSliceOfObjects(&f.Features, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, foundryOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output features: %w", err)
		}).
		WriteSliceOfObjects(&f.ImmutableFeatures, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, foundryOutputImmFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize foundry output immutable features: %w", err)
		}).
		Serialize()
}

func (f *FoundryOutput) Size() int {
	return util.NumByteLen(byte(OutputFoundry)) +
		util.NumByteLen(f.Amount) +
		f.NativeTokens.Size() +
		util.NumByteLen(f.SerialNumber) +
		f.TokenScheme.Size() +
		f.Conditions.Size() +
		f.Features.Size() +
		f.ImmutableFeatures.Size()
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

	jFoundryOutput.Features, err = serializablesToJSONRawMsgs(f.Features.ToSerializables())
	if err != nil {
		return nil, err
	}

	jFoundryOutput.ImmutableFeatures, err = serializablesToJSONRawMsgs(f.ImmutableFeatures.ToSerializables())
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
	Type              int                `json:"type"`
	Amount            string             `json:"amount"`
	NativeTokens      []*json.RawMessage `json:"nativeTokens,omitempty"`
	SerialNumber      int                `json:"serialNumber"`
	TokenScheme       *json.RawMessage   `json:"tokenScheme"`
	Conditions        []*json.RawMessage `json:"unlockConditions,omitempty"`
	Features          []*json.RawMessage `json:"features,omitempty"`
	ImmutableFeatures []*json.RawMessage `json:"immutableFeatures,omitempty"`
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

	e.TokenScheme, err = tokenSchemeFromJSONRawMsg(j.TokenScheme)
	if err != nil {
		return nil, err
	}

	e.Conditions, err = unlockConditionsFromJSONRawMsg(j.Conditions)
	if err != nil {
		return nil, err
	}

	e.Features, err = featuresFromJSONRawMsg(j.Features)
	if err != nil {
		return nil, err
	}

	e.ImmutableFeatures, err = featuresFromJSONRawMsg(j.ImmutableFeatures)
	if err != nil {
		return nil, err
	}

	return e, nil
}
