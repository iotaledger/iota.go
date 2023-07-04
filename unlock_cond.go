package iotago

import (
	"fmt"
	"sort"

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

	// Type returns the type of the UnlockCondition.
	Type() UnlockConditionType

	// Equal tells whether this UnlockCondition is equal to other.
	Equal(other UnlockCondition) bool

	// Clone clones the UnlockCondition.
	Clone() UnlockCondition
}

// UnlockConditions is a slice of UnlockCondition(s).
type UnlockConditions[T UnlockCondition] []T

func (f UnlockConditions[T]) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	var sumCost VBytes
	for _, unlockCond := range f {
		sumCost += unlockCond.VBytes(rentStruct, nil)
	}

	// length prefix + sum cost of conditions
	return rentStruct.VBFactorData.Multiply(serializer.OneByte) + sumCost
}

func (f UnlockConditions[T]) WorkScore(workScoreStructure *WorkScoreStructure) WorkScore {
	var sumCost WorkScore
	for _, unlockCond := range f {
		sumCost += unlockCond.WorkScore(workScoreStructure)
	}

	return sumCost
}

// Clone clones the UnlockConditions.
func (f UnlockConditions[T]) Clone() UnlockConditions[T] {
	cpy := make(UnlockConditions[T], len(f))
	for i, v := range f {
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

// HasManalockCondition tells whether the set has both a timelock and account address unlock.
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

// HasTimelockUntil tells us whether the set has a timelock that that is still locked at slotIndex.
func (f UnlockConditionSet) HasTimelockUntil(slotIndex SlotIndex) bool {
	timelock := f.Timelock()
	if timelock == nil {
		return false
	}
	if timelock.SlotIndex <= slotIndex {
		return false
	}
	return true
	// TODO: check for off by one error
}

// tells whether the given ident can unlock an output containing this set of UnlockCondition(s)
// when taking into consideration the constraints enforced by them:
//   - If the timelocks are not expired, then nobody can unlock.
//   - If the expiration blocks are expired, then only the return identity can unlock.
//
// returns booleans indicating whether the given ident can unlock and whether the return identity can unlock.
func (f UnlockConditionSet) unlockableBy(ident Address, txCreationTime SlotIndex) (givenIdentCanUnlock bool, returnIdentCanUnlock bool) {
	if err := f.TimelocksExpired(txCreationTime); err != nil {
		return false, false
	}

	// if the return ident can unlock, then ident must be the return ident
	var returnIdent Address
	if returnIdentCanUnlock, returnIdent = f.ReturnIdentCanUnlock(txCreationTime); returnIdentCanUnlock {
		if !ident.Equal(returnIdent) {
			return false, true
		}
		return true, true
	}

	return true, false
}

// ReturnIdentCanUnlock tells whether a sender defined in an expiration unlock condition within this set is the actual
// identity which could unlock an Output containing this UnlockConditionSet given the transaction creation time.
func (f UnlockConditionSet) ReturnIdentCanUnlock(txCreationTime SlotIndex) (bool, Address) {
	expUnlockCond := f.Expiration()

	if expUnlockCond == nil {
		return false, nil
	}

	if expUnlockCond.SlotIndex <= txCreationTime {
		return true, expUnlockCond.ReturnAddress
	}

	return false, nil
}

// TimelocksExpired tells whether UnlockCondition(s) in this set which impose a timelock are expired
// in relation to the given ExternalUnlockParameters.
func (f UnlockConditionSet) TimelocksExpired(txCreationTime SlotIndex) error {
	timelock := f.Timelock()

	if timelock == nil {
		return nil
	}

	if txCreationTime < timelock.SlotIndex {
		return ierrors.Wrapf(ErrTimelockNotExpired, "slotIndex cond %d vs. tx creation slot %d", timelock.SlotIndex, txCreationTime)
	}

	return nil
}

// StorageDepositReturn returns the StorageDepositReturnUnlockCondition in the set or nil.
func (f UnlockConditionSet) StorageDepositReturn() *StorageDepositReturnUnlockCondition {
	b, has := f[UnlockConditionStorageDepositReturn]
	if !has {
		return nil
	}
	return b.(*StorageDepositReturnUnlockCondition)
}

// Address returns the AddressUnlockCondition in the set or nil.
func (f UnlockConditionSet) Address() *AddressUnlockCondition {
	b, has := f[UnlockConditionAddress]
	if !has {
		return nil
	}
	return b.(*AddressUnlockCondition)
}

// ImmutableAccount returns the ImmutableAccountUnlockCondition in the set or nil.
func (f UnlockConditionSet) ImmutableAccount() *ImmutableAccountUnlockCondition {
	b, has := f[UnlockConditionImmutableAccount]
	if !has {
		return nil
	}
	return b.(*ImmutableAccountUnlockCondition)
}

// GovernorAddress returns the GovernorAddressUnlockCondition in the set or nil.
func (f UnlockConditionSet) GovernorAddress() *GovernorAddressUnlockCondition {
	b, has := f[UnlockConditionGovernorAddress]
	if !has {
		return nil
	}
	return b.(*GovernorAddressUnlockCondition)
}

// StateControllerAddress returns the StateControllerAddressUnlockCondition in the set or nil.
func (f UnlockConditionSet) StateControllerAddress() *StateControllerAddressUnlockCondition {
	b, has := f[UnlockConditionStateControllerAddress]
	if !has {
		return nil
	}
	return b.(*StateControllerAddressUnlockCondition)
}

// Timelock returns the TimelockUnlockCondition in the set or nil.
func (f UnlockConditionSet) Timelock() *TimelockUnlockCondition {
	b, has := f[UnlockConditionTimelock]
	if !has {
		return nil
	}
	return b.(*TimelockUnlockCondition)
}

// Expiration returns the ExpirationUnlockCondition in the set or nil.
func (f UnlockConditionSet) Expiration() *ExpirationUnlockCondition {
	b, has := f[UnlockConditionExpiration]
	if !has {
		return nil
	}
	return b.(*ExpirationUnlockCondition)
}
