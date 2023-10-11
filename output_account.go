package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrNonUniqueAccountOutputs gets returned when multiple AccountOutputs(s) with the same AccountID exist within sets.
	ErrNonUniqueAccountOutputs = ierrors.New("non unique accounts within outputs")
	// ErrInvalidAccountStateTransition gets returned when an account is doing an invalid state transition.
	ErrInvalidAccountStateTransition = ierrors.New("invalid account state transition")
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
	// ErrImplicitAccountDestructionDisallowed gets returned if an implicit account is destroyed, which is not allowed.
	ErrImplicitAccountDestructionDisallowed = ierrors.New("cannot destroy implicit account; must be transitioned to account")
	// ErrMultipleImplicitAccountCreationAddresses gets return when there is more than one
	// Implicit Account Creation Address on the input side of a transaction.
	ErrMultipleImplicitAccountCreationAddresses = ierrors.New("multiple implicit account creation addresses on the input side")
)

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
	// The identifier for this account.
	AccountID AccountID `serix:"3,mapKey=accountId"`
	// The counter that denotes the number of foundries created by this account.
	FoundryCounter uint32 `serix:"6,mapKey=foundryCounter"`
	// The unlock conditions on this output.
	Conditions AccountOutputUnlockConditions `serix:"7,mapKey=unlockConditions,omitempty"`
	// The features on the output.
	Features AccountOutputFeatures `serix:"8,mapKey=features,omitempty"`
	// The immutable feature on the output.
	ImmutableFeatures AccountOutputImmFeatures `serix:"9,mapKey=immutableFeatures,omitempty"`
}

func (a *AccountOutput) Clone() Output {
	return &AccountOutput{
		Amount:            a.Amount,
		Mana:              a.Mana,
		AccountID:         a.AccountID,
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

	if a.AccountID != otherOutput.AccountID {
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
		a.Conditions.StorageScore(rentStruct, nil) +
		a.Features.StorageScore(rentStruct, nil) +
		a.ImmutableFeatures.StorageScore(rentStruct, nil)
}

func (a *AccountOutput) syntacticallyValidate() error {
	// Address should never be nil.
	if a.Conditions.MustSet().Address().Address.Type() == AddressImplicitAccountCreation {
		return ErrImplicitAccountCreationAddressInInvalidOutput
	}

	return nil
}

func (a *AccountOutput) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
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

func (a *AccountOutput) Ident() Address {
	return a.Conditions.MustSet().Address().Address
}

func (a *AccountOutput) ChainID() ChainID {
	return a.AccountID
}

func (a *AccountOutput) AccountEmpty() bool {
	return a.AccountID == EmptyAccountID
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
		AccountIDLength +
		// FoundryCounter
		serializer.UInt32ByteSize +
		a.Conditions.Size() +
		a.Features.Size() +
		a.ImmutableFeatures.Size()
}
