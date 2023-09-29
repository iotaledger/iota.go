package iotago

import (
	"bytes"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// AccountIDLength is the byte length of an AccountID.
	AccountIDLength = IdentifierLength
)

var (
	// ErrNonUniqueAccountOutputs gets returned when multiple AccountOutputs(s) with the same AccountID exist within sets.
	ErrNonUniqueAccountOutputs = ierrors.New("non unique accounts within outputs")
	// ErrInvalidAccountStateTransition gets returned when an account is doing an invalid state transition.
	ErrInvalidAccountStateTransition = ierrors.New("invalid account state transition")
	// ErrInvalidAccountGovernanceTransition gets returned when an account is doing an invalid governance transition.
	ErrInvalidAccountGovernanceTransition = ierrors.New("invalid account governance transition")
	// ErrInvalidBlockIssuerTransition gets returned when an account tries to transition block issuer expiry too soon.
	ErrInvalidBlockIssuerTransition = ierrors.New("invalid block issuer transition")
	// ErrAccountLocked gets returned when an account has negative block issuance credits.
	ErrAccountLocked = ierrors.New("account is locked due to negative block issuance credits")
	// ErrBlockIssuanceCreditInputRequired gets returned when a transaction containing an account with a block issuer feature
	// does not have a Block Issuance Credit Input.
	ErrBlockIssuanceCreditInputRequired = ierrors.New("account with block issuer feature requires a block issuance credit input")
	// ErrInvalidStakingTransition gets returned when an account tries to do an invalid transition with a Staking Feature.
	ErrInvalidStakingTransition = ierrors.New("invalid staking transition")
	// ErrAccountMissing gets returned when an account is missing.
	ErrAccountMissing = ierrors.New("account is missing")
	// ErrImplicitAccountDestruction gets returned when an implicit account is being destroyed.
	ErrImplicitAccountDestruction = ierrors.New("implicit account destruction is not allowed")
	// ErrInvalidImplicitAccountConversion gets returned when there is an invalid conversion from implicit account to account output.
	ErrInvalidImplicitAccountConversion = ierrors.New("invalid conversion from implicit account to account output")
	emptyAccountID                      = [AccountIDLength]byte{}
)

func EmptyAccountID() AccountID {
	return emptyAccountID
}

// AccountID is the identifier for an account.
// It is computed as the Blake2b-256 hash of the OutputID of the output which created the account.
type AccountID = Identifier

// AccountIDs are IDs of accounts.
type AccountIDs []AccountID

func (id AccountID) Addressable() bool {
	return true
}

func (id AccountID) Key() interface{} {
	return id.String()
}

func (id AccountID) FromOutputID(in OutputID) ChainID {
	return AccountIDFromOutputID(in)
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
			return ierrors.Wrapf(ErrAccountMissing, "%s missing in source", accountID.ToHex())
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
			return nil, ierrors.Wrapf(ErrNonUniqueAccountOutputs, "account %s exists in both sets", k.ToHex())
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
	Amount BaseToken `serix:"0,mapKey=amount"`
	// The stored mana held by the output.
	Mana Mana `serix:"1,mapKey=mana"`
	// The native tokens held by the output.
	NativeTokens NativeTokens `serix:"2,mapKey=nativeTokens,omitempty"`
	// The identifier for this account.
	AccountID AccountID `serix:"3,mapKey=accountId"`
	// The index of the state.
	StateIndex uint32 `serix:"4,mapKey=stateIndex"`
	// The state of the account which can only be mutated by the state controller.
	StateMetadata []byte `serix:"5,lengthPrefixType=uint16,mapKey=stateMetadata,omitempty,maxLen=8192"`
	// The counter that denotes the number of foundries created by this account.
	FoundryCounter uint32 `serix:"6,mapKey=foundryCounter"`
	// The unlock conditions on this output.
	Conditions AccountOutputUnlockConditions `serix:"7,mapKey=unlockConditions,omitempty"`
	// The features on the output.
	Features AccountOutputFeatures `serix:"8,mapKey=features,omitempty"`
	// The immutable feature on the output.
	ImmutableFeatures AccountOutputImmFeatures `serix:"9,mapKey=immutableFeatures,omitempty"`
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
		Mana:              a.Mana,
		NativeTokens:      a.NativeTokens.Clone(),
		AccountID:         a.AccountID,
		StateIndex:        a.StateIndex,
		StateMetadata:     append([]byte(nil), a.StateMetadata...),
		FoundryCounter:    a.FoundryCounter,
		Conditions:        a.Conditions.Clone(),
		Features:          a.Features.Clone(),
		ImmutableFeatures: a.ImmutableFeatures.Clone(),
	}
}

func (a *AccountOutput) Equal(other Output) bool {
	otherOutput, isSameType := other.(*AccountOutput)
	if !isSameType {
		return false
	}

	if a.Amount != otherOutput.Amount {
		return false
	}

	if a.Mana != otherOutput.Mana {
		return false
	}

	if !a.NativeTokens.Equal(otherOutput.NativeTokens) {
		return false
	}

	if a.AccountID != otherOutput.AccountID {
		return false
	}

	if a.StateIndex != otherOutput.StateIndex {
		return false
	}

	if !bytes.Equal(a.StateMetadata, otherOutput.StateMetadata) {
		return false
	}

	if a.FoundryCounter != otherOutput.FoundryCounter {
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

func (a *AccountOutput) UnlockableBy(ident Address, next TransDepIdentOutput, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) (bool, error) {
	return outputUnlockableBy(a, next, ident, pastBoundedSlotIndex, futureBoundedSlotIndex)
}

func (a *AccountOutput) StorageScore(rentStruct *RentStructure, _ StorageScoreFunc) StorageScore {
	return storageScoreOffsetOutput(rentStruct) +
		rentStruct.StorageScoreFactorData().Multiply(StorageScore(a.Size())) +
		a.NativeTokens.StorageScore(rentStruct, nil) +
		a.Conditions.StorageScore(rentStruct, nil) +
		a.Features.StorageScore(rentStruct, nil) +
		a.ImmutableFeatures.StorageScore(rentStruct, nil)
}

func (a *AccountOutput) syntacticallyValidate() error {
	// Address should never be nil.
	stateControllerAddress := a.Conditions.MustSet().StateControllerAddress().Address
	governorAddress := a.Conditions.MustSet().GovernorAddress().Address

	if (stateControllerAddress.Type() == AddressImplicitAccountCreation) || (governorAddress.Type() == AddressImplicitAccountCreation) {
		return ErrImplicitAccountCreationAddressInInvalidOutput
	}

	return nil
}

func (a *AccountOutput) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	workScoreNativeTokens, err := a.NativeTokens.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

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

	return workScoreNativeTokens.Add(workScoreConditions, workScoreFeatures, workScoreImmutableFeatures)
}

func (a *AccountOutput) Ident(nextState TransDepIdentOutput) (Address, error) {
	// if there isn't a next state, then only the governance address can destroy the account
	if nextState == nil {
		return a.GovernorAddress(), nil
	}
	otherAccountOutput, isAccountOutput := nextState.(*AccountOutput)
	if !isAccountOutput {
		return nil, ierrors.Wrapf(ErrTransDepIdentOutputNextInvalid, "expected AccountOutput but got %s for ident computation", nextState.Type())
	}
	switch {
	case a.StateIndex == otherAccountOutput.StateIndex:
		return a.GovernorAddress(), nil
	case a.StateIndex+1 == otherAccountOutput.StateIndex:
		return a.StateController(), nil
	default:
		return nil, ierrors.Wrap(ErrTransDepIdentOutputNextInvalid, "can not compute right ident for account output as state index delta is invalid")
	}
}

func (a *AccountOutput) ChainID() ChainID {
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

func (a *AccountOutput) BaseTokenAmount() BaseToken {
	return a.Amount
}

func (a *AccountOutput) StoredMana() Mana {
	return a.Mana
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
	// OutputType
	return serializer.OneByte +
		BaseTokenSize +
		ManaSize +
		a.NativeTokens.Size() +
		AccountIDLength +
		// StateIndex
		serializer.UInt32ByteSize +
		serializer.UInt16ByteSize +
		len(a.StateMetadata) +
		// FoundryCounter
		serializer.UInt32ByteSize +
		a.Conditions.Size() +
		a.Features.Size() +
		a.ImmutableFeatures.Size()
}
