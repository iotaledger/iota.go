package iotago

import (
	"bytes"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// AnchorIDLength is the byte length of an AnchorID.
	AnchorIDLength = IdentifierLength
)

var (
	// ErrNonUniqueAnchorOutputs gets returned when multiple AnchorOutputs(s) with the same AnchorID exist within sets.
	ErrNonUniqueAnchorOutputs = ierrors.New("non unique Anchors within outputs")
	// ErrInvalidAnchorStateTransition gets returned when an Anchor is doing an invalid state transition.
	ErrInvalidAnchorStateTransition = ierrors.New("invalid Anchor state transition")
	// ErrInvalidAnchorGovernanceTransition gets returned when an Anchor is doing an invalid governance transition.
	ErrInvalidAnchorGovernanceTransition = ierrors.New("invalid Anchor governance transition")
	// ErrAnchorMissing gets returned when an Anchor is missing.
	ErrAnchorMissing = ierrors.New("Anchor is missing")

	emptyAnchorID = [AnchorIDLength]byte{}
)

func EmptyAnchorID() AnchorID {
	return emptyAnchorID
}

// AnchorID is the identifier for an Anchor.
// It is computed as the Blake2b-256 hash of the OutputID of the output which created the Anchor.
type AnchorID = Identifier

// AnchorIDs are IDs of Anchors.
type AnchorIDs []AnchorID

// AnchorIDFromOutputID returns the AnchorID computed from a given OutputID.
func AnchorIDFromOutputID(outputID OutputID) AnchorID {
	return blake2b.Sum256(outputID[:])
}

// AnchorOutputs is a slice of AnchorOutput(s).
type AnchorOutputs []*AnchorOutput

// Every checks whether every element passes f.
// Returns either -1 if all elements passed f or the index of the first element which didn't.
func (outputs AnchorOutputs) Every(f func(output *AnchorOutput) bool) int {
	for i, output := range outputs {
		if !f(output) {
			return i
		}
	}

	return -1
}

// AnchorOutputsSet is a set of AnchorOutput(s).
type AnchorOutputsSet map[AnchorID]*AnchorOutput

// Includes checks whether all Anchors included in other exist in this set.
func (set AnchorOutputsSet) Includes(other AnchorOutputsSet) error {
	for AnchorID := range other {
		if _, has := set[AnchorID]; !has {
			return ierrors.Wrapf(ErrAnchorMissing, "%s missing in source", AnchorID.ToHex())
		}
	}

	return nil
}

// EveryTuple runs f for every key which exists in both this set and other.
func (set AnchorOutputsSet) EveryTuple(other AnchorOutputsSet, f func(in *AnchorOutput, out *AnchorOutput) error) error {
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
// Returns an error if an Anchor isn't unique across both sets.
func (set AnchorOutputsSet) Merge(other AnchorOutputsSet) (AnchorOutputsSet, error) {
	newSet := make(AnchorOutputsSet)
	for k, v := range set {
		newSet[k] = v
	}
	for k, v := range other {
		if _, has := newSet[k]; has {
			return nil, ierrors.Wrapf(ErrNonUniqueAnchorOutputs, "Anchor %s exists in both sets", k.ToHex())
		}
		newSet[k] = v
	}

	return newSet, nil
}

type (
	anchorOutputUnlockCondition  interface{ UnlockCondition }
	anchorOutputFeature          interface{ Feature }
	anchorOutputImmFeature       interface{ Feature }
	AnchorOutputUnlockConditions = UnlockConditions[anchorOutputUnlockCondition]
	AnchorOutputFeatures         = Features[anchorOutputFeature]
	AnchorOutputImmFeatures      = Features[anchorOutputImmFeature]
)

// AnchorOutput is an output type which represents an Anchor.
type AnchorOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount BaseToken `serix:"0,mapKey=amount"`
	// The stored mana held by the output.
	Mana Mana `serix:"1,mapKey=mana"`
	// The identifier for this Anchor.
	AnchorID AnchorID `serix:"3,mapKey=AnchorId"`
	// The index of the state.
	StateIndex uint32 `serix:"4,mapKey=stateIndex"`
	// The state of the Anchor which can only be mutated by the state controller.
	StateMetadata []byte `serix:"5,lengthPrefixType=uint16,mapKey=stateMetadata,omitempty,maxLen=8192"`
	// The unlock conditions on this output.
	Conditions AnchorOutputUnlockConditions `serix:"7,mapKey=unlockConditions,omitempty"`
	// The features on the output.
	Features AnchorOutputFeatures `serix:"8,mapKey=features,omitempty"`
	// The immutable feature on the output.
	ImmutableFeatures AnchorOutputImmFeatures `serix:"9,mapKey=immutableFeatures,omitempty"`
}

func (a *AnchorOutput) GovernorAddress() Address {
	return a.Conditions.MustSet().GovernorAddress().Address
}

func (a *AnchorOutput) StateController() Address {
	return a.Conditions.MustSet().StateControllerAddress().Address
}

func (a *AnchorOutput) Clone() Output {
	return &AnchorOutput{
		Amount:            a.Amount,
		Mana:              a.Mana,
		AnchorID:          a.AnchorID,
		StateIndex:        a.StateIndex,
		StateMetadata:     append([]byte(nil), a.StateMetadata...),
		Conditions:        a.Conditions.Clone(),
		Features:          a.Features.Clone(),
		ImmutableFeatures: a.ImmutableFeatures.Clone(),
	}
}

func (a *AnchorOutput) Equal(other Output) bool {
	otherOutput, isSameType := other.(*AnchorOutput)
	if !isSameType {
		return false
	}

	if a.Amount != otherOutput.Amount {
		return false
	}

	if a.Mana != otherOutput.Mana {
		return false
	}

	if a.AnchorID != otherOutput.AnchorID {
		return false
	}

	if a.StateIndex != otherOutput.StateIndex {
		return false
	}

	if !bytes.Equal(a.StateMetadata, otherOutput.StateMetadata) {
		return false
	}

	if !a.Conditions.Equal(otherOutput.Conditions) {
		return false
	}

	if !a.Features.Equal(otherOutput.Features) {
		return false
	}

	if !a.ImmutableFeatures.Equal(otherOutput.ImmutableFeatures) {
		return false
	}

	return true
}

func (a *AnchorOutput) UnlockableBy(ident Address, next TransDepIdentOutput, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) (bool, error) {
	return outputUnlockableBy(a, next, ident, pastBoundedSlotIndex, futureBoundedSlotIndex)
}

func (a *AnchorOutput) StorageScore(rentStruct *RentStructure, _ StorageScoreFunc) StorageScore {
	return storageScoreOffsetOutput(rentStruct) +
		rentStruct.StorageScoreFactorData().Multiply(StorageScore(a.Size())) +
		a.Conditions.StorageScore(rentStruct, nil) +
		a.Features.StorageScore(rentStruct, nil) +
		a.ImmutableFeatures.StorageScore(rentStruct, nil)
}

func (a *AnchorOutput) syntacticallyValidate() error {
	// Address should never be nil.
	stateControllerAddress := a.Conditions.MustSet().StateControllerAddress().Address
	governorAddress := a.Conditions.MustSet().GovernorAddress().Address

	if (stateControllerAddress.Type() == AddressImplicitAccountCreation) || (governorAddress.Type() == AddressImplicitAccountCreation) {
		return ErrImplicitAccountCreationAddressInInvalidOutput
	}

	return nil
}

func (a *AnchorOutput) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	workScoreConditions, err := a.Conditions.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	workScoreFeatures, err := a.Features.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	workScoreImmutableFeatures, err := a.ImmutableFeatures.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	return workScoreConditions.Add(workScoreFeatures, workScoreImmutableFeatures)
}

func (a *AnchorOutput) Ident(nextState TransDepIdentOutput) (Address, error) {
	// if there isn't a next state, then only the governance address can destroy the Anchor
	if nextState == nil {
		return a.GovernorAddress(), nil
	}
	otherAnchorOutput, isAnchorOutput := nextState.(*AnchorOutput)
	if !isAnchorOutput {
		return nil, ierrors.Wrapf(ErrTransDepIdentOutputNextInvalid, "expected AnchorOutput but got %s for ident computation", nextState.Type())
	}
	switch {
	case a.StateIndex == otherAnchorOutput.StateIndex:
		return a.GovernorAddress(), nil
	case a.StateIndex+1 == otherAnchorOutput.StateIndex:
		return a.StateController(), nil
	default:
		return nil, ierrors.Wrap(ErrTransDepIdentOutputNextInvalid, "can not compute right ident for Anchor output as state index delta is invalid")
	}
}

func (a *AnchorOutput) ChainID() ChainID {
	return a.AnchorID
}

func (a *AnchorOutput) AnchorEmpty() bool {
	return a.AnchorID == emptyAnchorID
}

func (a *AnchorOutput) FeatureSet() FeatureSet {
	return a.Features.MustSet()
}

func (a *AnchorOutput) UnlockConditionSet() UnlockConditionSet {
	return a.Conditions.MustSet()
}

func (a *AnchorOutput) ImmutableFeatureSet() FeatureSet {
	return a.ImmutableFeatures.MustSet()
}

func (a *AnchorOutput) BaseTokenAmount() BaseToken {
	return a.Amount
}

func (a *AnchorOutput) StoredMana() Mana {
	return a.Mana
}

func (a *AnchorOutput) Target() (Address, error) {
	addr := new(AnchorAddress)
	copy(addr[:], a.AnchorID[:])

	return addr, nil
}

func (a *AnchorOutput) Type() OutputType {
	return OutputAnchor
}

func (a *AnchorOutput) Size() int {
	// OutputType
	return serializer.OneByte +
		BaseTokenSize +
		ManaSize +
		AnchorIDLength +
		// StateIndex
		serializer.UInt32ByteSize +
		serializer.UInt16ByteSize +
		len(a.StateMetadata) +
		a.Conditions.Size() +
		a.Features.Size() +
		a.ImmutableFeatures.Size()
}
