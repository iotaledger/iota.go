package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrImplicitAccountDestructionDisallowed gets returned if an implicit account is destroyed, which is not allowed.
	ErrImplicitAccountDestructionDisallowed = ierrors.New("cannot destroy implicit account; must be transitioned to account")
	// ErrMultipleImplicitAccountCreationAddresses gets return when there is more than one
	// Implicit Account Creation Address on the input side of a transaction.
	ErrMultipleImplicitAccountCreationAddresses = ierrors.New("multiple implicit account creation addresses on the input side")
	// ErrAccountInvalidFoundryCounter gets returned when the foundry counter in an account decreased
	// or did not increase by the number of new foundries.
	ErrAccountInvalidFoundryCounter = ierrors.New("foundry counter in account decreased or did not increase by the number of new foundries")
)

type (
	AccountOutputUnlockCondition  interface{ UnlockCondition }
	AccountOutputFeature          interface{ Feature }
	AccountOutputImmFeature       interface{ Feature }
	AccountOutputUnlockConditions = UnlockConditions[AccountOutputUnlockCondition]
	AccountOutputFeatures         = Features[AccountOutputFeature]
	AccountOutputImmFeatures      = Features[AccountOutputImmFeature]
)

// AccountOutput is an output type which represents an account.
type AccountOutput struct {
	// The amount of IOTA tokens held by the output.
	Amount BaseToken `serix:""`
	// The stored mana held by the output.
	Mana Mana `serix:""`
	// The identifier for this account.
	AccountID AccountID `serix:""`
	// The counter that denotes the number of foundries created by this account.
	FoundryCounter uint32 `serix:""`
	// The unlock conditions on this output.
	UnlockConditions AccountOutputUnlockConditions `serix:",omitempty"`
	// The features on the output.
	Features AccountOutputFeatures `serix:",omitempty"`
	// The immutable feature on the output.
	ImmutableFeatures AccountOutputImmFeatures `serix:",omitempty"`
}

func (a *AccountOutput) Clone() Output {
	return &AccountOutput{
		Amount:            a.Amount,
		Mana:              a.Mana,
		AccountID:         a.AccountID,
		FoundryCounter:    a.FoundryCounter,
		UnlockConditions:  a.UnlockConditions.Clone(),
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

	if !a.UnlockConditions.Equal(otherOutput.UnlockConditions) {
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

func (a *AccountOutput) UnlockableBy(addr Address, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) bool {
	ok, _ := outputUnlockableBy(a, nil, addr, pastBoundedSlotIndex, futureBoundedSlotIndex)
	return ok
}

func (a *AccountOutput) StorageScore(storageScoreStruct *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return storageScoreStruct.OffsetOutput +
		storageScoreStruct.FactorData().Multiply(StorageScore(a.Size())) +
		a.UnlockConditions.StorageScore(storageScoreStruct, nil) +
		a.Features.StorageScore(storageScoreStruct, nil) +
		a.ImmutableFeatures.StorageScore(storageScoreStruct, nil)
}

func (a *AccountOutput) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	workScoreConditions, err := a.UnlockConditions.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreFeatures, err := a.Features.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreImmutableFeatures, err := a.ImmutableFeatures.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	return workScoreParameters.Output.Add(workScoreConditions, workScoreFeatures, workScoreImmutableFeatures)
}

func (a *AccountOutput) Owner() Address {
	return a.UnlockConditions.MustSet().Address().Address
}

func (a *AccountOutput) ChainID() ChainID {
	return a.AccountID
}

func (a *AccountOutput) FeatureSet() FeatureSet {
	return a.Features.MustSet()
}

func (a *AccountOutput) UnlockConditionSet() UnlockConditionSet {
	return a.UnlockConditions.MustSet()
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
		a.UnlockConditions.Size() +
		a.Features.Size() +
		a.ImmutableFeatures.Size()
}
