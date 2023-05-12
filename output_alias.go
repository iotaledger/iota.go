package iotago

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/util"
)

const (
	// AccountIDLength is the byte length of an AccountID.
	AccountIDLength = blake2b.Size256
)

var (
	// ErrNonUniqueAccountOutputs gets returned when multiple AccountOutputs(s) with the same AccountID exist within sets.
	ErrNonUniqueAccountOutputs = errors.New("non unique accounts within outputs")
	// ErrInvalidAccountStateTransition gets returned when an account is doing an invalid state transition.
	ErrInvalidAccountStateTransition = errors.New("invalid account state transition")
	// ErrInvalidAccountGovernanceTransition gets returned when an account is doing an invalid governance transition.
	ErrInvalidAccountGovernanceTransition = errors.New("invalid account governance transition")
	// ErrAccountMissing gets returned when an account is missing.
	ErrAccountMissing = errors.New("account is missing")
	emptyAccountID    = [AccountIDLength]byte{}
)

// AccountID is the identifier for an account.
// It is computed as the Blake2b-256 hash of the OutputID of the output which created the account.
type AccountID [AccountIDLength]byte

func (id AccountID) Addressable() bool {
	return true
}

func (id AccountID) ToHex() string {
	return EncodeHex(id[:])
}

func (id AccountID) Key() interface{} {
	return id.String()
}

func (id AccountID) FromOutputID(in OutputID) ChainID {
	return AccountIDFromOutputID(in)
}

func (id AccountID) Empty() bool {
	return id == emptyAccountID
}

func (id AccountID) String() string {
	return EncodeHex(id[:])
}

func (id AccountID) Matches(other ChainID) bool {
	otherAccountID, isAccountID := other.(AccountID)
	if !isAccountID {
		return false
	}
	return id == otherAccountID
}

func (id AccountID) ToAddress() ChainAddress {
	var addr AccountAddress
	copy(addr[:], id[:])
	return &addr
}

// AccountIDFromOutputID returns the AccountID computed from a given OutputID.
func AccountIDFromOutputID(outputID OutputID) AccountID {
	return blake2b.Sum256(outputID[:])
}

// AccountOutputs is a slice of AccountOutput(s).
type AccountOutputs []*AccountOutput

// Every checks whether every element passes f.
// Returns either -1 if all elements passed f or the index of the first element which didn't.
func (outputs AccountOutputs) Every(f func(output *AccountOutput) bool) int {
	for i, output := range outputs {
		if !f(output) {
			return i
		}
	}
	return -1
}

// AccountOutputsSet is a set of AccountOutput(s).
type AccountOutputsSet map[AccountID]*AccountOutput

// Includes checks whether all accounts included in other exist in this set.
func (set AccountOutputsSet) Includes(other AccountOutputsSet) error {
	for accountID := range other {
		if _, has := set[accountID]; !has {
			return fmt.Errorf("%w: %s missing in source", ErrAccountMissing, accountID.ToHex())
		}
	}
	return nil
}

// EveryTuple runs f for every key which exists in both this set and other.
func (set AccountOutputsSet) EveryTuple(other AccountOutputsSet, f func(in *AccountOutput, out *AccountOutput) error) error {
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
// Returns an error if an account isn't unique across both sets.
func (set AccountOutputsSet) Merge(other AccountOutputsSet) (AccountOutputsSet, error) {
	newSet := make(AccountOutputsSet)
	for k, v := range set {
		newSet[k] = v
	}
	for k, v := range other {
		if _, has := newSet[k]; has {
			return nil, fmt.Errorf("%w: account %s exists in both sets", ErrNonUniqueAccountOutputs, k.ToHex())
		}
		newSet[k] = v
	}
	return newSet, nil
}

type (
	accountOutputUnlockCondition  interface{ UnlockCondition }
	accountOutputFeature          interface{ Feature }
	accountOutputImmFeature       interface{ Feature }
	AccountOutputUnlockConditions = UnlockConditions[accountOutputUnlockCondition]
	AccountOutputFeatures         = Features[accountOutputFeature]
	AccountOutputImmFeatures      = Features[accountOutputImmFeature]
)

// AccountOutput is an output type which represents an account.
type AccountOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount uint64 `serix:"0,mapKey=amount"`
	// The native tokens held by the output.
	NativeTokens NativeTokens `serix:"1,mapKey=nativeTokens,omitempty"`
	// The identifier for this account.
	AccountID AccountID `serix:"2,mapKey=accountId"`
	// The index of the state.
	StateIndex uint32 `serix:"3,mapKey=stateIndex"`
	// The state of the account which can only be mutated by the state controller.
	StateMetadata []byte `serix:"4,lengthPrefixType=uint16,mapKey=stateMetadata,omitempty,maxLen=8192"`
	// The counter that denotes the number of foundries created by this account.
	FoundryCounter uint32 `serix:"5,mapKey=foundryCounter"`
	// The unlock conditions on this output.
	Conditions AccountOutputUnlockConditions `serix:"6,mapKey=unlockConditions,omitempty"`
	// The features on the output.
	Features AccountOutputFeatures `serix:"7,mapKey=features,omitempty"`
	// The immutable feature on the output.
	ImmutableFeatures AccountOutputImmFeatures `serix:"8,mapKey=immutableFeatures,omitempty"`
}

func (a *AccountOutput) GovernorAddress() Address {
	return a.Conditions.MustSet().GovernorAddress().Address
}

func (a *AccountOutput) StateController() Address {
	return a.Conditions.MustSet().StateControllerAddress().Address
}

func (a *AccountOutput) Clone() Output {
	return &AccountOutput{
		Amount:            a.Amount,
		AccountID:         a.AccountID,
		NativeTokens:      a.NativeTokens.Clone(),
		StateIndex:        a.StateIndex,
		StateMetadata:     append([]byte(nil), a.StateMetadata...),
		FoundryCounter:    a.FoundryCounter,
		Conditions:        a.Conditions.Clone(),
		Features:          a.Features.Clone(),
		ImmutableFeatures: a.ImmutableFeatures.Clone(),
	}
}

func (a *AccountOutput) UnlockableBy(ident Address, next TransDepIdentOutput, extParams *ExternalUnlockParameters) (bool, error) {
	return outputUnlockable(a, next, ident, extParams)
}

func (a *AccountOutput) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return outputOffsetVByteCost(rentStruct) +
		// prefix + amount
		rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize) +
		a.NativeTokens.VBytes(rentStruct, nil) +
		rentStruct.VBFactorData.Multiply(AccountIDLength) +
		// state index, state meta length, state meta, foundry counter
		rentStruct.VBFactorData.Multiply(VBytes(serializer.UInt32ByteSize+serializer.UInt16ByteSize+len(a.StateMetadata)+serializer.UInt32ByteSize)) +
		a.Conditions.VBytes(rentStruct, nil) +
		a.Features.VBytes(rentStruct, nil) +
		a.ImmutableFeatures.VBytes(rentStruct, nil)
}

func (a *AccountOutput) Ident(nextState TransDepIdentOutput) (Address, error) {
	// if there isn't a next state, then only the governance address can destroy the account
	if nextState == nil {
		return a.GovernorAddress(), nil
	}
	otherAccountOutput, isAccountOutput := nextState.(*AccountOutput)
	if !isAccountOutput {
		return nil, fmt.Errorf("%w: expected AccountOutput but got %s for ident computation", ErrTransDepIdentOutputNextInvalid, nextState.Type())
	}
	switch {
	case a.StateIndex == otherAccountOutput.StateIndex:
		return a.GovernorAddress(), nil
	case a.StateIndex+1 == otherAccountOutput.StateIndex:
		return a.StateController(), nil
	default:
		return nil, fmt.Errorf("%w: can not compute right ident for account output as state index delta is invalid", ErrTransDepIdentOutputNextInvalid)
	}
}

func (a *AccountOutput) Chain() ChainID {
	return a.AccountID
}

func (a *AccountOutput) AccountEmpty() bool {
	return a.AccountID == emptyAccountID
}

func (a *AccountOutput) NativeTokenList() NativeTokens {
	return a.NativeTokens
}

func (a *AccountOutput) FeatureSet() FeatureSet {
	return a.Features.MustSet()
}

func (a *AccountOutput) UnlockConditionSet() UnlockConditionSet {
	return a.Conditions.MustSet()
}

func (a *AccountOutput) ImmutableFeatureSet() FeatureSet {
	return a.ImmutableFeatures.MustSet()
}

func (a *AccountOutput) Deposit() uint64 {
	return a.Amount
}

func (a *AccountOutput) Target() (Address, error) {
	addr := new(AccountAddress)
	copy(addr[:], a.AccountID[:])
	return addr, nil
}

func (a *AccountOutput) Type() OutputType {
	return OutputAccount
}

func (a *AccountOutput) Size() int {
	return util.NumByteLen(byte(OutputAccount)) +
		util.NumByteLen(a.Amount) +
		a.NativeTokens.Size() +
		AccountIDLength +
		util.NumByteLen(a.StateIndex) +
		serializer.UInt16ByteSize +
		len(a.StateMetadata) +
		util.NumByteLen(a.FoundryCounter) +
		a.Conditions.Size() +
		a.Features.Size() +
		a.ImmutableFeatures.Size()
}
