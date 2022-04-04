package iotago

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

const (
	// AliasIDLength is the byte length of an AliasID.
	AliasIDLength = 20
)

var (
	// ErrNonUniqueAliasOutputs gets returned when multiple AliasOutputs(s) with the same AliasID exist within sets.
	ErrNonUniqueAliasOutputs = errors.New("non unique aliases within outputs")
	// ErrInvalidAliasStateTransition gets returned when an alias is doing an invalid state transition.
	ErrInvalidAliasStateTransition = errors.New("invalid alias state transition")
	// ErrInvalidAliasGovernanceTransition gets returned when an alias is doing an invalid governance transition.
	ErrInvalidAliasGovernanceTransition = errors.New("invalid alias governance transition")
	// ErrAliasMissing gets returned when an alias is missing
	ErrAliasMissing = errors.New("alias is missing")
	emptyAliasID    = [AliasIDLength]byte{}

	aliasOutputUnlockCondsArrayRules = &serializer.ArrayRules{
		Min: 2, Max: 2,
		MustOccur: serializer.TypePrefixes{
			uint32(UnlockConditionStateControllerAddress): struct{}{},
			uint32(UnlockConditionGovernorAddress):        struct{}{},
		},
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(UnlockConditionStateControllerAddress):
				case uint32(UnlockConditionGovernorAddress):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize alias output, unsupported unlock condition type %s", ErrUnsupportedUnlockConditionType, UnlockConditionType(ty))
				}
				return UnlockConditionSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *StateControllerAddressUnlockCondition:
				case *GovernorAddressUnlockCondition:
				default:
					return fmt.Errorf("%w: in alias output", ErrUnsupportedUnlockConditionType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	aliasOutputFeatBlockArrayRules = &serializer.ArrayRules{
		Min: 0,
		Max: 3,
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(FeatureBlockSender):
				case uint32(FeatureBlockMetadata):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize alias output, unsupported feature block type %s", ErrUnsupportedFeatureBlockType, FeatureBlockType(ty))
				}
				return FeatureBlockSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *SenderFeatureBlock:
				case *MetadataFeatureBlock:
				default:
					return fmt.Errorf("%w: in alias output", ErrUnsupportedFeatureBlockType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}

	aliasOutputImmFeatBlockArrayRules = &serializer.ArrayRules{
		Min: 0,
		Max: 2,
		Guards: serializer.SerializableGuard{
			ReadGuard: func(ty uint32) (serializer.Serializable, error) {
				switch ty {
				case uint32(FeatureBlockIssuer):
				case uint32(FeatureBlockMetadata):
				default:
					return nil, fmt.Errorf("%w: unable to deserialize alias output, unsupported immutable feature block type %s", ErrUnsupportedFeatureBlockType, FeatureBlockType(ty))
				}
				return FeatureBlockSelector(ty)
			},
			WriteGuard: func(seri serializer.Serializable) error {
				switch seri.(type) {
				case *IssuerFeatureBlock:
				case *MetadataFeatureBlock:
				default:
					return fmt.Errorf("%w: in alias output", ErrUnsupportedFeatureBlockType)
				}
				return nil
			},
		},
		ValidationMode: serializer.ArrayValidationModeNoDuplicates |
			serializer.ArrayValidationModeLexicalOrdering |
			serializer.ArrayValidationModeAtMostOneOfEachTypeByte,
	}
)

// AliasOutputFeatureBlocksArrayRules returns array rules defining the constraints on FeatureBlocks within an AliasOutput.
func AliasOutputFeatureBlocksArrayRules() serializer.ArrayRules {
	return *aliasOutputFeatBlockArrayRules
}

// AliasOutputImmutableFeatureBlocksArrayRules returns array rules defining the constraints on immutable FeatureBlocks within an AliasOutput.
func AliasOutputImmutableFeatureBlocksArrayRules() serializer.ArrayRules {
	return *aliasOutputImmFeatBlockArrayRules
}

// AliasID is the identifier for an alias account.
// It is computed as the Blake2b-160 hash of the OutputID of the output which created the account.
type AliasID [AliasIDLength]byte

func (id AliasID) Addressable() bool {
	return true
}

func (id AliasID) Key() interface{} {
	return id.String()
}

func (id AliasID) FromOutputID(in OutputID) ChainID {
	return AliasIDFromOutputID(in)
}

func (id AliasID) Empty() bool {
	return id == emptyAliasID
}

func (id AliasID) String() string {
	return EncodeHex(id[:])
}

func (id AliasID) Matches(other ChainID) bool {
	otherAliasID, isAliasID := other.(AliasID)
	if !isAliasID {
		return false
	}
	return id == otherAliasID
}

func (id AliasID) ToAddress() ChainConstrainedAddress {
	var addr AliasAddress
	copy(addr[:], id[:])
	return &addr
}

// AliasIDFromOutputID returns the AliasID computed from a given OutputID.
func AliasIDFromOutputID(outputID OutputID) AliasID {
	// TODO: maybe use pkg with Sum160 exposed
	blake2b160, _ := blake2b.New(20, nil)
	var aliasID AliasID
	if _, err := blake2b160.Write(outputID[:]); err != nil {
		panic(err)
	}
	copy(aliasID[:], blake2b160.Sum(nil))
	return aliasID
}

// AliasOutputs is a slice of AliasOutput(s).
type AliasOutputs []*AliasOutput

// Every checks whether every element passes f.
// Returns either -1 if all elements passed f or the index of the first element which didn't
func (outputs AliasOutputs) Every(f func(output *AliasOutput) bool) int {
	for i, output := range outputs {
		if !f(output) {
			return i
		}
	}
	return -1
}

// AliasOutputsSet is a set of AliasOutput(s).
type AliasOutputsSet map[AliasID]*AliasOutput

// Includes checks whether all aliases included in other exist in this set.
func (set AliasOutputsSet) Includes(other AliasOutputsSet) error {
	for aliasID := range other {
		if _, has := set[aliasID]; !has {
			return fmt.Errorf("%w: %s missing in source", ErrAliasMissing, aliasID)
		}
	}
	return nil
}

// EveryTuple runs f for every key which exists in both this set and other.
func (set AliasOutputsSet) EveryTuple(other AliasOutputsSet, f func(in *AliasOutput, out *AliasOutput) error) error {
	for k, v := range set {
		v2, has := other[k]
		if !has {
			continue
		}
		if err := f(v, v2); err != nil {
			return err
		}
	}
	return nil
}

// Merge merges other with this set in a new set.
// Returns an error if an alias isn't unique across both sets.
func (set AliasOutputsSet) Merge(other AliasOutputsSet) (AliasOutputsSet, error) {
	newSet := make(AliasOutputsSet)
	for k, v := range set {
		newSet[k] = v
	}
	for k, v := range other {
		if _, has := newSet[k]; has {
			return nil, fmt.Errorf("%w: alias %s exists in both sets", ErrNonUniqueAliasOutputs, k)
		}
		newSet[k] = v
	}
	return newSet, nil
}

// AliasOutput is an output type which represents an alias account.
type AliasOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64
	// The native tokens held by the output.
	NativeTokens NativeTokens
	// The identifier for this alias account.
	AliasID AliasID
	// The index of the state.
	StateIndex uint32
	// The state of the alias account which can only be mutated by the state controller.
	StateMetadata []byte
	// The counter that denotes the number of foundries created by this alias account.
	FoundryCounter uint32
	// The unlock conditions on this output.
	Conditions UnlockConditions
	// The feature blocks on the output.
	Blocks FeatureBlocks
	// The immutable feature blocks on the output.
	ImmutableBlocks FeatureBlocks
}

func (a *AliasOutput) GovernorAddress() Address {
	return a.Conditions.MustSet().GovernorAddress().Address
}

func (a *AliasOutput) StateController() Address {
	return a.Conditions.MustSet().StateControllerAddress().Address
}

func (a *AliasOutput) Clone() Output {
	return &AliasOutput{
		Amount:          a.Amount,
		AliasID:         a.AliasID,
		NativeTokens:    a.NativeTokens.Clone(),
		StateIndex:      a.StateIndex,
		StateMetadata:   append([]byte(nil), a.StateMetadata...),
		FoundryCounter:  a.FoundryCounter,
		Conditions:      a.Conditions.Clone(),
		Blocks:          a.Blocks.Clone(),
		ImmutableBlocks: a.ImmutableBlocks.Clone(),
	}
}

func (a *AliasOutput) UnlockableBy(ident Address, next TransDepIdentOutput, extParas *ExternalUnlockParameters) (bool, error) {
	return outputUnlockable(a, next, ident, extParas)
}

func (a *AliasOutput) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return outputOffsetVByteCost(costStruct) +
		// prefix + amount
		costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		a.NativeTokens.VByteCost(costStruct, nil) +
		costStruct.VBFactorData.Multiply(AliasIDLength) +
		// state index, state meta length, state meta, foundry counter
		costStruct.VBFactorData.Multiply(uint64(serializer.UInt32ByteSize+serializer.UInt16ByteSize+len(a.StateMetadata)+serializer.UInt32ByteSize)) +
		a.Conditions.VByteCost(costStruct, nil) +
		a.Blocks.VByteCost(costStruct, nil) +
		a.ImmutableBlocks.VByteCost(costStruct, nil)
}

//	- For output AliasOutput(s) with non-zeroed AliasID, there must be a corresponding input AliasOutput where either
//	  its AliasID is zeroed and StateIndex and FoundryCounter are zero or an input AliasOutput with the same AliasID.
//	- On alias state transitions:
//		- The StateIndex must be incremented by 1
//		- Only Amount, NativeTokens, StateIndex, StateMetadata and FoundryCounter can be mutated
//	- On alias governance transition:
//		- Only StateController (must be mutated), GovernanceController and the MetadataBlock can be mutated
func (a *AliasOutput) ValidateStateTransition(transType ChainTransitionType, next ChainConstrainedOutput, semValCtx *SemanticValidationContext) error {
	var err error
	switch transType {
	case ChainTransitionTypeGenesis:
		err = a.genesisValid(semValCtx)
	case ChainTransitionTypeStateChange:
		err = a.stateChangeValid(semValCtx, next)
	case ChainTransitionTypeDestroy:
		return nil
	default:
		panic("unknown chain transition type in AliasOutput")
	}
	if err != nil {
		return &ChainTransitionError{Inner: err, Msg: fmt.Sprintf("alias %s", a.AliasID)}
	}
	return nil
}

func (a *AliasOutput) genesisValid(semValCtx *SemanticValidationContext) error {
	if !a.AliasID.Empty() {
		return fmt.Errorf("AliasOutput's ID is not zeroed even though it is new")
	}
	return IsIssuerOnOutputUnlocked(a, semValCtx.WorkingSet.UnlockedIdents)
}

func (a *AliasOutput) stateChangeValid(semValCtx *SemanticValidationContext, next ChainConstrainedOutput) error {
	nextState, is := next.(*AliasOutput)
	if !is {
		return fmt.Errorf("can only state transition to another alias output")
	}
	if !a.ImmutableBlocks.Equal(nextState.ImmutableBlocks) {
		return fmt.Errorf("old state %s, next state %s", a.ImmutableBlocks, nextState.ImmutableBlocks)
	}
	if a.StateIndex == nextState.StateIndex {
		return a.GovernanceSTVF(nextState, semValCtx)
	}
	return a.StateSTVF(nextState, semValCtx)
}

func (a *AliasOutput) Ident(nextState TransDepIdentOutput) (Address, error) {
	// if there isn't a next state, then only the governance address can destroy the alias
	if nextState == nil {
		return a.GovernorAddress(), nil
	}
	otherAliasOutput, isAliasOutput := nextState.(*AliasOutput)
	if !isAliasOutput {
		return nil, fmt.Errorf("%w: expected AliasOutput but got %s for ident computation", ErrTransDepIdentOutputNextInvalid, nextState.Type())
	}
	switch {
	case a.StateIndex == otherAliasOutput.StateIndex:
		return a.GovernorAddress(), nil
	case a.StateIndex+1 == otherAliasOutput.StateIndex:
		return a.StateController(), nil
	default:
		return nil, fmt.Errorf("%w: can not compute right ident for alias output as state index delta is invalid", ErrTransDepIdentOutputNextInvalid)
	}
}

func (a *AliasOutput) Chain() ChainID {
	return a.AliasID
}

// GovernanceSTVF checks whether the governance transition with other is valid.
// Under a governance transition, only the StateController, GovernanceController and MetadataFeatureBlock can change.
func (a *AliasOutput) GovernanceSTVF(nextAliasOutput *AliasOutput, semValCtx *SemanticValidationContext) error {
	switch {
	case a.Amount != nextAliasOutput.Amount:
		return fmt.Errorf("%w: amount changed, in %d / out %d ", ErrInvalidAliasGovernanceTransition, a.Amount, nextAliasOutput.Amount)
	case !a.NativeTokens.Equal(nextAliasOutput.NativeTokens):
		return fmt.Errorf("%w: native tokens changed, in %v / out %v", ErrInvalidAliasGovernanceTransition, a.NativeTokens, nextAliasOutput.NativeTokens)
	case a.StateIndex != nextAliasOutput.StateIndex:
		return fmt.Errorf("%w: state index changed, in %d / out %d", ErrInvalidAliasGovernanceTransition, a.StateIndex, nextAliasOutput.StateIndex)
	case !bytes.Equal(a.StateMetadata, nextAliasOutput.StateMetadata):
		return fmt.Errorf("%w: state metadata changed, in %v / out %v", ErrInvalidAliasGovernanceTransition, a.StateMetadata, nextAliasOutput.StateMetadata)
	case a.FoundryCounter != nextAliasOutput.FoundryCounter:
		return fmt.Errorf("%w: foundry counter changed, in %d / out %d", ErrInvalidAliasGovernanceTransition, a.FoundryCounter, nextAliasOutput.FoundryCounter)
	}
	return nil
}

// StateSTVF checks whether the state transition with other is valid.
// Under a state transition, only Amount, NativeTokens, StateIndex, StateMetadata, SenderFeatureBlock and FoundryCounter can change.
func (a *AliasOutput) StateSTVF(nextAliasOutput *AliasOutput, semValCtx *SemanticValidationContext) error {
	switch {
	case !a.StateController().Equal(nextAliasOutput.StateController()):
		return fmt.Errorf("%w: state controller changed, in %v / out %v", ErrInvalidAliasStateTransition, a.StateController(), nextAliasOutput.StateController())
	case !a.GovernorAddress().Equal(nextAliasOutput.GovernorAddress()):
		return fmt.Errorf("%w: governance controller changed, in %v / out %v", ErrInvalidAliasStateTransition, a.StateController(), nextAliasOutput.StateController())
	case a.FoundryCounter > nextAliasOutput.FoundryCounter:
		return fmt.Errorf("%w: foundry counter of next state is less than previous, in %d / out %d", ErrInvalidAliasStateTransition, a.FoundryCounter, nextAliasOutput.FoundryCounter)
	case a.StateIndex+1 != nextAliasOutput.StateIndex:
		return fmt.Errorf("%w: state index %d on the input side but %d on the output side", ErrInvalidAliasStateTransition, a.StateIndex, nextAliasOutput.StateIndex)
	}

	if err := FeatureBlockUnchanged(FeatureBlockMetadata, a.Blocks.MustSet(), nextAliasOutput.Blocks.MustSet()); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidAliasStateTransition, err)
	}

	// check that for a foundry counter change, X amount of foundries were actually created
	if a.FoundryCounter == nextAliasOutput.FoundryCounter {
		return nil
	}

	var seenNewFoundriesOfAlias uint32
	for _, output := range semValCtx.WorkingSet.Tx.Essence.Outputs {
		foundryOutput, is := output.(*FoundryOutput)
		if !is {
			continue
		}

		if _, notNew := semValCtx.WorkingSet.InChains[foundryOutput.MustID()]; notNew {
			continue
		}

		foundryAliasID := foundryOutput.Ident().(*AliasAddress).Chain()
		if !foundryAliasID.Matches(nextAliasOutput.AliasID) {
			continue
		}
		seenNewFoundriesOfAlias++
	}

	expectedNewFoundriesCount := nextAliasOutput.FoundryCounter - a.FoundryCounter
	if expectedNewFoundriesCount != seenNewFoundriesOfAlias {
		return fmt.Errorf("%w: %d new foundries were created but the alias output's foundry counter changed by %d", ErrInvalidAliasStateTransition, seenNewFoundriesOfAlias, expectedNewFoundriesCount)
	}

	return nil
}

func (a *AliasOutput) AliasEmpty() bool {
	return a.AliasID == emptyAliasID
}

func (a *AliasOutput) NativeTokenSet() NativeTokens {
	return a.NativeTokens
}

func (a *AliasOutput) FeatureBlocks() FeatureBlocks {
	return a.Blocks
}

func (a *AliasOutput) ImmutableFeatureBlocks() FeatureBlocks {
	return a.ImmutableBlocks
}

func (a *AliasOutput) UnlockConditions() UnlockConditions {
	return a.Conditions
}

func (a *AliasOutput) Deposit() uint64 {
	return a.Amount
}

func (a *AliasOutput) Target() (serializer.Serializable, error) {
	addr := new(AliasAddress)
	copy(addr[:], a.AliasID[:])
	return addr, nil
}

func (a *AliasOutput) Type() OutputType {
	return OutputAlias
}

func (a *AliasOutput) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	return serializer.NewDeserializer(data).
		CheckTypePrefix(uint32(OutputAlias), serializer.TypeDenotationByte, func(err error) error {
			return fmt.Errorf("unable to deserialize alias output: %w", err)
		}).
		ReadNum(&a.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for alias output: %w", err)
		}).
		ReadSliceOfObjects(&a.NativeTokens, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationNone, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize native tokens for alias output: %w", err)
		}).
		ReadBytesInPlace(a.AliasID[:], func(err error) error {
			return fmt.Errorf("unable to deserialize alias ID for alias output: %w", err)
		}).
		ReadNum(&a.StateIndex, func(err error) error {
			return fmt.Errorf("unable to deserialize state index for alias output: %w", err)
		}).
		ReadVariableByteSlice(&a.StateMetadata, serializer.SeriLengthPrefixTypeAsUint16, func(err error) error {
			return fmt.Errorf("unable to deserialize state metadata for alias output: %w", err)
		}, MaxMetadataLength).
		ReadNum(&a.FoundryCounter, func(err error) error {
			return fmt.Errorf("unable to deserialize foundry counter for alias output: %w", err)
		}).
		ReadSliceOfObjects(&a.Conditions, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, aliasOutputUnlockCondsArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize unlock conditions for alias output: %w", err)
		}).
		ReadSliceOfObjects(&a.Blocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, aliasOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize feature blocks for alias output: %w", err)
		}).
		ReadSliceOfObjects(&a.ImmutableBlocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, serializer.TypeDenotationByte, aliasOutputImmFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize immutable feature blocks for alias output: %w", err)
		}).
		Done()
}

func (a *AliasOutput) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteNum(byte(OutputAlias), func(err error) error {
			return fmt.Errorf("unable to serialize alias output type ID: %w", err)
		}).
		WriteNum(a.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize alias output amount: %w", err)
		}).
		WriteSliceOfObjects(&a.NativeTokens, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, nativeTokensArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize alias output native tokens: %w", err)
		}).
		WriteBytes(a.AliasID[:], func(err error) error {
			return fmt.Errorf("unable to serialize alias output alias ID: %w", err)
		}).
		WriteNum(a.StateIndex, func(err error) error {
			return fmt.Errorf("unable to serialize alias output state index: %w", err)
		}).
		WriteVariableByteSlice(a.StateMetadata, serializer.SeriLengthPrefixTypeAsUint16, func(err error) error {
			return fmt.Errorf("unable to serialize alias output state metadata: %w", err)
		}).
		WriteNum(a.FoundryCounter, func(err error) error {
			return fmt.Errorf("unable to serialize alias output foundry counter: %w", err)
		}).
		WriteSliceOfObjects(&a.Conditions, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, aliasOutputUnlockCondsArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize alias output unlock conditions: %w", err)
		}).
		WriteSliceOfObjects(&a.Blocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, aliasOutputFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize alias output feature blocks: %w", err)
		}).
		WriteSliceOfObjects(&a.ImmutableBlocks, deSeriMode, deSeriCtx, serializer.SeriLengthPrefixTypeAsByte, aliasOutputImmFeatBlockArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize alias output immutable feature blocks: %w", err)
		}).
		Serialize()
}

func (a *AliasOutput) Size() int {
	return util.NumByteLen(byte(OutputAlias)) +
		util.NumByteLen(a.Amount) +
		a.NativeTokens.Size() +
		AliasIDLength +
		util.NumByteLen(a.StateIndex) +
		serializer.UInt16ByteSize +
		len(a.StateMetadata) +
		util.NumByteLen(a.FoundryCounter) +
		a.Conditions.Size() +
		a.Blocks.Size() +
		a.ImmutableBlocks.Size()
}

func (a *AliasOutput) MarshalJSON() ([]byte, error) {
	var err error
	jAliasOutput := &jsonAliasOutput{
		Type:           int(OutputAlias),
		Amount:         EncodeUint64(a.Amount),
		StateIndex:     int(a.StateIndex),
		FoundryCounter: int(a.FoundryCounter),
	}

	jAliasOutput.NativeTokens, err = serializablesToJSONRawMsgs(a.NativeTokens.ToSerializables())
	if err != nil {
		return nil, err
	}

	jAliasOutput.AliasID = EncodeHex(a.AliasID[:])

	jAliasOutput.StateMetadata = EncodeHex(a.StateMetadata)

	jAliasOutput.Conditions, err = serializablesToJSONRawMsgs(a.Conditions.ToSerializables())
	if err != nil {
		return nil, err
	}

	jAliasOutput.Blocks, err = serializablesToJSONRawMsgs(a.Blocks.ToSerializables())
	if err != nil {
		return nil, err
	}

	jAliasOutput.ImmutableBlocks, err = serializablesToJSONRawMsgs(a.ImmutableBlocks.ToSerializables())
	if err != nil {
		return nil, err
	}

	return json.Marshal(jAliasOutput)
}

func (a *AliasOutput) UnmarshalJSON(bytes []byte) error {
	jAliasOutput := &jsonAliasOutput{}
	if err := json.Unmarshal(bytes, jAliasOutput); err != nil {
		return err
	}
	seri, err := jAliasOutput.ToSerializable()
	if err != nil {
		return err
	}
	*a = *seri.(*AliasOutput)
	return nil
}

// jsonAliasOutput defines the json representation of an AliasOutput.
type jsonAliasOutput struct {
	Type            int                `json:"type"`
	Amount          string             `json:"amount"`
	NativeTokens    []*json.RawMessage `json:"nativeTokens"`
	AliasID         string             `json:"aliasId"`
	StateIndex      int                `json:"stateIndex"`
	StateMetadata   string             `json:"stateMetadata"`
	FoundryCounter  int                `json:"foundryCounter"`
	Conditions      []*json.RawMessage `json:"unlockConditions"`
	Blocks          []*json.RawMessage `json:"featureBlocks"`
	ImmutableBlocks []*json.RawMessage `json:"immutableFeatureBlocks"`
}

func (j *jsonAliasOutput) ToSerializable() (serializer.Serializable, error) {
	var err error
	e := &AliasOutput{
		StateIndex:     uint32(j.StateIndex),
		FoundryCounter: uint32(j.FoundryCounter),
	}

	e.Amount, err = DecodeUint64(j.Amount)
	if err != nil {
		return nil, err
	}

	e.NativeTokens, err = nativeTokensFromJSONRawMsg(j.NativeTokens)
	if err != nil {
		return nil, err
	}

	aliasIDSlice, err := DecodeHex(j.AliasID)
	if err != nil {
		return nil, err
	}
	copy(e.AliasID[:], aliasIDSlice)

	e.StateMetadata, err = DecodeHex(j.StateMetadata)
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
