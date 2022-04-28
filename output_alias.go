package iotago

import (
	"bytes"
	"errors"
	"fmt"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

const (
	// AliasIDLength is the byte length of an AliasID.
	AliasIDLength = blake2b.Size256
)

var (
	// ErrNonUniqueAliasOutputs gets returned when multiple AliasOutputs(s) with the same AliasID exist within sets.
	ErrNonUniqueAliasOutputs = errors.New("non unique aliases within outputs")
	// ErrInvalidAliasStateTransition gets returned when an alias is doing an invalid state transition.
	ErrInvalidAliasStateTransition = errors.New("invalid alias state transition")
	// ErrInvalidAliasGovernanceTransition gets returned when an alias is doing an invalid governance transition.
	ErrInvalidAliasGovernanceTransition = errors.New("invalid alias governance transition")
	// ErrAliasMissing gets returned when an alias is missing.
	ErrAliasMissing = errors.New("alias is missing")
	emptyAliasID    = [AliasIDLength]byte{}
)

// AliasID is the identifier for an alias account.
// It is computed as the Blake2b-256 hash of the OutputID of the output which created the account.
type AliasID [AliasIDLength]byte

func (id AliasID) Addressable() bool {
	return true
}

func (id AliasID) ToHex() string {
	return EncodeHex(id[:])
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
	return blake2b.Sum256(outputID[:])
}

// AliasOutputs is a slice of AliasOutput(s).
type AliasOutputs []*AliasOutput

// Every checks whether every element passes f.
// Returns either -1 if all elements passed f or the index of the first element which didn't.
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

type (
	AliasUnlockCondition interface{ UnlockCondition }
	AliasFeature         interface{ Feature }
	AliasImmFeature      interface{ Feature }
)

// AliasOutput is an output type which represents an alias account.
type AliasOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64 `serix:"0,mapKey=amount"`
	// The native tokens held by the output.
	NativeTokens NativeTokens `serix:"1,mapKey=nativeTokens,omitempty"`
	// The identifier for this alias account.
	AliasID AliasID `serix:"2,mapKey=aliasId"`
	// The index of the state.
	StateIndex uint32 `serix:"3,mapKey=stateIndex"`
	// The state of the alias account which can only be mutated by the state controller.
	StateMetadata []byte `serix:"4,lengthPrefixType=uint16,mapKey=stateMetadata,omitempty,maxLen=8192"`
	// The counter that denotes the number of foundries created by this alias account.
	FoundryCounter uint32 `serix:"5,mapKey=foundryCounter"`
	// The unlock conditions on this output.
	Conditions UnlockConditions[AliasUnlockCondition] `serix:"6,mapKey=unlockConditions,omitempty"`
	// The features on the output.
	Features Features[AliasFeature] `serix:"7,mapKey=features,omitempty"`
	// The immutable feature on the output.
	ImmutableFeatures Features[AliasImmFeature] `serix:"8,mapKey=immutableFeatures,omitempty"`
}

func (a *AliasOutput) GovernorAddress() Address {
	return a.Conditions.MustSet().GovernorAddress().Address
}

func (a *AliasOutput) StateController() Address {
	return a.Conditions.MustSet().StateControllerAddress().Address
}

func (a *AliasOutput) Clone() Output {
	return &AliasOutput{
		Amount:            a.Amount,
		AliasID:           a.AliasID,
		NativeTokens:      a.NativeTokens.Clone(),
		StateIndex:        a.StateIndex,
		StateMetadata:     append([]byte(nil), a.StateMetadata...),
		FoundryCounter:    a.FoundryCounter,
		Conditions:        a.Conditions.Clone(),
		Features:          a.Features.Clone(),
		ImmutableFeatures: a.ImmutableFeatures.Clone(),
	}
}

func (a *AliasOutput) UnlockableBy(ident Address, next TransDepIdentOutput, extParas *ExternalUnlockParameters) (bool, error) {
	return outputUnlockable(a, next, ident, extParas)
}

func (a *AliasOutput) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return outputOffsetVByteCost(rentStruct) +
		// prefix + amount
		rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		a.NativeTokens.VBytes(rentStruct, nil) +
		rentStruct.VBFactorData.Multiply(AliasIDLength) +
		// state index, state meta length, state meta, foundry counter
		rentStruct.VBFactorData.Multiply(uint64(serializer.UInt32ByteSize+serializer.UInt16ByteSize+len(a.StateMetadata)+serializer.UInt32ByteSize)) +
		a.Conditions.VBytes(rentStruct, nil) +
		a.Features.VBytes(rentStruct, nil) +
		a.ImmutableFeatures.VBytes(rentStruct, nil)
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
	if !a.ImmutableFeatures.Equal(nextState.ImmutableFeatures) {
		return fmt.Errorf("old state %s, next state %s", a.ImmutableFeatures, nextState.ImmutableFeatures)
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
// Under a governance transition, only the StateController, GovernanceController and MetadataFeature can change.
func (a *AliasOutput) GovernanceSTVF(nextAliasOutput *AliasOutput, _ *SemanticValidationContext) error {
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
// Under a state transition, only Amount, NativeTokens, StateIndex, StateMetadata, SenderFeature and FoundryCounter can change.
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

	if err := FeatureUnchanged(FeatureMetadata, a.Features.MustSet(), nextAliasOutput.Features.MustSet()); err != nil {
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

func (a *AliasOutput) NativeTokenList() NativeTokens {
	return a.NativeTokens
}

func (a *AliasOutput) FeatureSet() FeatureSet {
	return a.Features.MustSet()
}

func (a *AliasOutput) UnlockConditionSet() UnlockConditionSet {
	return a.Conditions.MustSet()
}

func (a *AliasOutput) ImmutableFeatureSet() FeatureSet {
	return a.ImmutableFeatures.MustSet()
}

func (a *AliasOutput) Deposit() uint64 {
	return a.Amount
}

func (a *AliasOutput) Target() (Address, error) {
	addr := new(AliasAddress)
	copy(addr[:], a.AliasID[:])
	return addr, nil
}

func (a *AliasOutput) Type() OutputType {
	return OutputAlias
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
		a.Features.Size() +
		a.ImmutableFeatures.Size()
}
