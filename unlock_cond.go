package iotago

import (
	"fmt"
	"sort"

	"github.com/iotaledger/hive.go/constraints"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrNonUniqueUnlockConditions gets returned when multiple UnlockCondition(s) with the same UnlockConditionType exist within sets.
	ErrNonUniqueUnlockConditions = ierrors.New("non unique unlock conditions within outputs")
	// ErrTimelockNotExpired gets returned when timelocks in a UnlockConditionSet are not expired.
	ErrTimelockNotExpired = ierrors.New("timelock not expired")
	// ErrExpirationConditionZero gets returned when an ExpirationUnlockCondition has set the slot index to zero.
	ErrExpirationConditionZero = ierrors.New("expiration condition is zero")
	// ErrTimelockConditionZero gets returned when a TimelockUnlockCondition has set the slot index to zero.
	ErrTimelockConditionZero = ierrors.New("timelock condition is zero")
	// ErrTimelockConditionCommitmentInputRequired gets returned when a TX containing a TimelockUnlockCondition
	// does not have a commitment input.
	ErrTimelockConditionCommitmentInputRequired = ierrors.New("transaction's containing a timelock condition require a commitment input")
	// ErrExpirationConditionCommitmentInputRequired gets returned when a TX containing an ExpirationUnlockCondition
	// does not have a commitment input.
	ErrExpirationConditionCommitmentInputRequired = ierrors.New("transaction's containing an expiration condition require a commitment input")
	// ErrExpirationConditionUnlockFailed gets returned when a ExpirationUnlockCondition could not be unlocked.
	ErrExpirationConditionUnlockFailed = ierrors.New("expiration condition unlock failed")
)

// UnlockConditionType defines the type of UnlockCondition.
type UnlockConditionType byte

const (
	// UnlockConditionAddress denotes an AddressUnlockCondition.
	UnlockConditionAddress UnlockConditionType = iota
	// UnlockConditionStorageDepositReturn denotes a StorageDepositReturnUnlockCondition.
	UnlockConditionStorageDepositReturn
	// UnlockConditionTimelock denotes a TimelockUnlockCondition.
	UnlockConditionTimelock
	// UnlockConditionExpiration denotes an ExpirationUnlockCondition.
	UnlockConditionExpiration
	// UnlockConditionStateControllerAddress denotes a StateControllerAddressUnlockCondition.
	UnlockConditionStateControllerAddress
	// UnlockConditionGovernorAddress denotes a GovernorAddressUnlockCondition.
	UnlockConditionGovernorAddress
	// UnlockConditionImmutableAccount denotes an ImmutableAccountUnlockCondition.
	UnlockConditionImmutableAccount
)

func (unlockCondType UnlockConditionType) String() string {
	if int(unlockCondType) >= len(unlockCondNames) {
		return fmt.Sprintf("unknown unlock condition type: %d", unlockCondType)
	}

	return unlockCondNames[unlockCondType]
}

var (
	unlockCondNames = [UnlockConditionImmutableAccount + 1]string{
		"AddressUnlockCondition",
		"StorageDepositReturnUnlockCondition",
		"TimelockUnlockCondition",
		"ExpirationUnlockCondition",
		"StateControllerAddressUnlockCondition",
		"GovernorAddressUnlockCondition",
		"ImmutableAccountUnlockCondition",
	}
)

// UnlockCondition is an abstract building block defining the unlock conditions of an Output.
type UnlockCondition interface {
	Sizer
	NonEphemeralObject
	ProcessableObject
	constraints.Cloneable[UnlockCondition]

	// Type returns the type of the UnlockCondition.
	Type() UnlockConditionType

	// Equal tells whether this UnlockCondition is equal to other.
	Equal(other UnlockCondition) bool
}

// UnlockConditions is a slice of UnlockCondition(s).
type UnlockConditions[T UnlockCondition] []T

func (f UnlockConditions[T]) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	var sumCost VBytes
	for _, unlockCond := range f {
		sumCost += unlockCond.VBytes(rentStruct, nil)
	}

	// length prefix + sum cost of conditions
	return rentStruct.VBFactorData().Multiply(serializer.OneByte) + sumCost
}

func (f UnlockConditions[T]) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	var workScoreUnlockConds WorkScore
	for _, unlockCond := range f {
		workScoreUnlockCond, err := unlockCond.WorkScore(workScoreStructure)
		if err != nil {
			return 0, err
		}

		workScoreUnlockConds, err = workScoreUnlockConds.Add(workScoreUnlockCond)
		if err != nil {
			return 0, err
		}
	}

	return workScoreUnlockConds, nil
}

// Clone clones the UnlockConditions.
func (f UnlockConditions[T]) Clone() UnlockConditions[T] {
	cpy := make(UnlockConditions[T], len(f))
	for i, v := range f {
		//nolint:forcetypeassert // we can safely assume that this is of type T
		cpy[i] = v.Clone().(T)
	}

	return cpy
}

func (f UnlockConditions[T]) Size() int {
	sum := serializer.OneByte // 1 byte length prefix
	for _, uc := range f {
		sum += uc.Size()
	}

	return sum
}

// Set converts the slice into an UnlockConditionSet.
// Returns an error if an UnlockConditionType occurs multiple times.
func (f UnlockConditions[T]) Set() (UnlockConditionSet, error) {
	set := make(UnlockConditionSet)
	for _, block := range f {
		if _, has := set[block.Type()]; has {
			return nil, ErrNonUniqueUnlockConditions
		}
		set[block.Type()] = block
	}

	return set, nil
}

// MustSet works like Set but panics if an error occurs.
// This function is therefore only safe to be called when it is given,
// that an UnlockConditions slice does not contain the same UnlockConditionType multiple times.
func (f UnlockConditions[T]) MustSet() UnlockConditionSet {
	set, err := f.Set()
	if err != nil {
		panic(err)
	}

	return set
}

// Upsert adds the given unlock condition or updates the previous one if existing.
func (f *UnlockConditions[T]) Upsert(unlockCondition T) {
	for i, ele := range *f {
		if ele.Type() == unlockCondition.Type() {
			(*f)[i] = unlockCondition

			return
		}
	}
	*f = append(*f, unlockCondition)
}

// Sort sorts the UnlockConditions in place by type.
func (f UnlockConditions[T]) Sort() {
	sort.Slice(f, func(i, j int) bool { return f[i].Type() < f[j].Type() })
}

// UnlockConditionSet is a set of UnlockCondition(s).
type UnlockConditionSet map[UnlockConditionType]UnlockCondition

// HasStorageDepositReturnCondition tells whether this set has a StorageDepositReturnUnlockCondition.
func (f UnlockConditionSet) HasStorageDepositReturnCondition() bool {
	return f.StorageDepositReturn() != nil
}

// HasExpirationCondition tells whether this set has an ExpirationUnlockCondition.
func (f UnlockConditionSet) HasExpirationCondition() bool {
	return f.Expiration() != nil
}

// HasTimelockCondition tells whether this set has a TimelockUnlockCondition.
func (f UnlockConditionSet) HasTimelockCondition() bool {
	return f.Timelock() != nil
}

// HasManalockCondition tells whether the set has both an account address unlock
// and a timelock that is still locked at slotIndex.
func (f UnlockConditionSet) HasManalockCondition(accountID AccountID, slotIndex SlotIndex) bool {
	if !f.HasTimelockUntil(slotIndex) {
		return false
	}
	unlockAddress := f.Address()
	if unlockAddress == nil {
		return false
	}
	if unlockAddress.Address.Type() != AddressAccount {
		return false
	}
	if !unlockAddress.Address.Equal(accountID.ToAddress()) {
		return false
	}

	return true
}

// HasTimelockUntil tells us whether the set has a timelock that is still locked at slotIndex.
func (f UnlockConditionSet) HasTimelockUntil(slotIndex SlotIndex) bool {
	// TODO: Test this.
	timelock := f.Timelock()
	return timelock != nil && slotIndex < timelock.SlotIndex
}

// tells whether the given ident can unlock an output containing this set of UnlockCondition(s)
// returns booleans indicating whether the given ident can unlock and whether the return identity can unlock.
func (f UnlockConditionSet) unlockableBy(ident Address, owner Address, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) bool {
	if err := f.TimelocksExpired(futureBoundedSlotIndex); err != nil {
		return false
	}

	// if the return ident can unlock, then ident must be the return ident
	if returnIdentCanUnlock, returnIdent := f.ReturnIdentCanUnlock(futureBoundedSlotIndex); returnIdentCanUnlock {
		return ident.Equal(returnIdent)
	}

	// if the past bounded index is less than the expiration slot index, then owner can unlock
	if f.OwnerIdentCanUnlock(pastBoundedSlotIndex) {
		return ident.Equal(owner)
	}

	return false
}

// OwnerIdentCanUnlock tells whether the target address defined in an expiration unlock condition within this set is
// allowed to unlock an Output containing this UnlockConditionSet given the past bounded slot index of the tx defined as
// the slot index of the commitment input plus the max committable age.
func (f UnlockConditionSet) OwnerIdentCanUnlock(pastBoundedSlotIndex SlotIndex) bool {
	expUnlockCond := f.Expiration()

	// if there is not expiration unlock, then the owner can unlock.
	if expUnlockCond == nil {
		return true
	}

	if pastBoundedSlotIndex < expUnlockCond.SlotIndex {
		return true
	}

	return false
}

// ReturnIdentCanUnlock tells whether a sender defined in an expiration unlock condition within this set is the actual
// identity which could unlock an Output containing this UnlockConditionSet given the future bounded slot index of the tx
// defined as the slot index of the commitment input plus the min committable age.
func (f UnlockConditionSet) ReturnIdentCanUnlock(futureBoundedSlotIndex SlotIndex) (bool, Address) {
	expUnlockCond := f.Expiration()

	if expUnlockCond == nil {
		return false, nil
	}

	if futureBoundedSlotIndex >= expUnlockCond.SlotIndex {
		return true, expUnlockCond.ReturnAddress
	}

	return false, nil
}

// TimelocksExpired tells whether UnlockCondition(s) in this set which impose a timelock are expired
// in relation to the given future bounded slot index. The provided slot index is the slot index of the commitment
// input which is being spent by the transaction plus the min committable age.
func (f UnlockConditionSet) TimelocksExpired(futureBoundedIndex SlotIndex) error {
	timelock := f.Timelock()

	if timelock == nil {
		return nil
	}

	if futureBoundedIndex < timelock.SlotIndex {
		return ierrors.Wrapf(ErrTimelockNotExpired, "slotIndex cond is %d, while tx creation slot could be up to %d", timelock.SlotIndex, futureBoundedIndex)
	}

	return nil
}

// StorageDepositReturn returns the StorageDepositReturnUnlockCondition in the set or nil.
func (f UnlockConditionSet) StorageDepositReturn() *StorageDepositReturnUnlockCondition {
	b, has := f[UnlockConditionStorageDepositReturn]
	if !has {
		return nil
	}

	//nolint:forcetypeassert // we can safely assume that this is a StorageDepositReturnUnlockCondition
	return b.(*StorageDepositReturnUnlockCondition)
}

// Address returns the AddressUnlockCondition in the set or nil.
func (f UnlockConditionSet) Address() *AddressUnlockCondition {
	b, has := f[UnlockConditionAddress]
	if !has {
		return nil
	}

	//nolint:forcetypeassert // we can safely assume that this is an AddressUnlockCondition
	return b.(*AddressUnlockCondition)
}

// ImmutableAccount returns the ImmutableAccountUnlockCondition in the set or nil.
func (f UnlockConditionSet) ImmutableAccount() *ImmutableAccountUnlockCondition {
	b, has := f[UnlockConditionImmutableAccount]
	if !has {
		return nil
	}

	//nolint:forcetypeassert // we can safely assume that this is an ImmutableAccountUnlockCondition
	return b.(*ImmutableAccountUnlockCondition)
}

// GovernorAddress returns the GovernorAddressUnlockCondition in the set or nil.
func (f UnlockConditionSet) GovernorAddress() *GovernorAddressUnlockCondition {
	b, has := f[UnlockConditionGovernorAddress]
	if !has {
		return nil
	}

	//nolint:forcetypeassert // we can safely assume that this is an GovernorAddressUnlockCondition
	return b.(*GovernorAddressUnlockCondition)
}

// StateControllerAddress returns the StateControllerAddressUnlockCondition in the set or nil.
func (f UnlockConditionSet) StateControllerAddress() *StateControllerAddressUnlockCondition {
	b, has := f[UnlockConditionStateControllerAddress]
	if !has {
		return nil
	}

	//nolint:forcetypeassert // we can safely assume that this is an StateControllerAddressUnlockCondition
	return b.(*StateControllerAddressUnlockCondition)
}

// Timelock returns the TimelockUnlockCondition in the set or nil.
func (f UnlockConditionSet) Timelock() *TimelockUnlockCondition {
	b, has := f[UnlockConditionTimelock]
	if !has {
		return nil
	}

	//nolint:forcetypeassert // we can safely assume that this is an TimelockUnlockCondition
	return b.(*TimelockUnlockCondition)
}

// Expiration returns the ExpirationUnlockCondition in the set or nil.
func (f UnlockConditionSet) Expiration() *ExpirationUnlockCondition {
	b, has := f[UnlockConditionExpiration]
	if !has {
		return nil
	}

	//nolint:forcetypeassert // we can safely assume that this is an ExpirationUnlockCondition
	return b.(*ExpirationUnlockCondition)
}
