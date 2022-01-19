package iotago

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrNonUniqueUnlockConditions gets returned when multiple UnlockCondition(s) with the same UnlockConditionType exist within sets.
	ErrNonUniqueUnlockConditions = errors.New("non unique unlock conditions within outputs")
	// ErrTimelockNotExpired gets returned when timelocks in a UnlockConditionsSet are not expired.
	ErrTimelockNotExpired = errors.New("timelock not expired")
	// ErrExpirationConditionsZero gets returned when an ExpirationUnlockCondition has set the milestone index and timestamp to zero.
	ErrExpirationConditionsZero = errors.New("expiration conditions are both zero")
	// ErrTimelockConditionsZero gets returned when a TimelockUnlockCondition has set the milestone index and timestamp to zero.
	ErrTimelockConditionsZero = errors.New("timelock conditions are both zero")
)

// UnlockConditionType defines the type of feature blocks.
type UnlockConditionType byte

const (
	// UnlockConditionAddress denotes an AddressUnlockCondition.
	UnlockConditionAddress UnlockConditionType = iota
	// UnlockConditionDustDepositReturn denotes a DustDepositReturnUnlockCondition.
	UnlockConditionDustDepositReturn
	// UnlockConditionTimelock denotes a TimelockUnlockCondition.
	UnlockConditionTimelock
	// UnlockConditionExpiration denotes an ExpirationUnlockCondition.
	UnlockConditionExpiration
	// UnlockConditionStateControllerAddress denotes a StateControllerAddressUnlockCondition.
	UnlockConditionStateControllerAddress
	// UnlockConditionGovernorAddress denotes a GovernorAddressUnlockCondition.
	UnlockConditionGovernorAddress
)

// UnlockCondition is an abstract building block defining the unlock conditions of an Output.
type UnlockCondition interface {
	serializer.Serializable
	NonEphemeralObject

	// Type returns the type of the UnlockCondition.
	Type() UnlockConditionType

	// Equal tells whether this UnlockCondition is equal to other.
	Equal(other UnlockCondition) bool

	// Clone clones the UnlockCondition.
	Clone() UnlockCondition
}

// UnlockConditions is a slice of UnlockCondition(s).
type UnlockConditions []UnlockCondition

func (f UnlockConditions) VByteCost(costStruct *RentStructure, override VByteCostFunc) uint64 {
	// TODO: adjust
	return 0
}

// Clone clones the UnlockConditions.
func (f UnlockConditions) Clone() UnlockConditions {
	cpy := make(UnlockConditions, len(f))
	for i, v := range f {
		cpy[i] = v.Clone()
	}
	return cpy
}

func (f UnlockConditions) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(f))
	for i, x := range f {
		seris[i] = x.(serializer.Serializable)
	}
	return seris
}

func (f *UnlockConditions) FromSerializables(seris serializer.Serializables) {
	*f = make(UnlockConditions, len(seris))
	for i, seri := range seris {
		(*f)[i] = seri.(UnlockCondition)
	}
}

// Set converts the slice into an UnlockConditionsSet.
// Returns an error if an UnlockConditionType occurs multiple times.
func (f UnlockConditions) Set() (UnlockConditionsSet, error) {
	set := make(UnlockConditionsSet)
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
func (f UnlockConditions) MustSet() UnlockConditionsSet {
	set, err := f.Set()
	if err != nil {
		panic(err)
	}
	return set
}

// UnlockConditionsSet is a set of UnlockCondition(s).
type UnlockConditionsSet map[UnlockConditionType]UnlockCondition

// HasExpirationConditions tells whether this set has any conditions putting an expiration.
func (f UnlockConditionsSet) HasExpirationConditions() bool {
	return f.Expiration() != nil
}

// tells whether the given ident can unlock an output containing this set of UnlockCondition(s)
// when taking into consideration the constraints enforced by them:
//	- If the timelocks are not expired, then nobody can unlock.
//	- If the expiration blocks are expired, then only the return identity can unlock.
// returns booleans indicating whether the given ident can unlock and whether the return identity can unlock.
func (f UnlockConditionsSet) unlockableBy(ident Address, extParas *ExternalUnlockParameters) (givenIdentCanUnlock bool, returnIdentCanUnlock bool) {
	if err := f.TimelocksExpired(extParas); err != nil {
		return false, false
	}

	// if the return ident can unlock, then ident must be the return ident
	var returnIdent Address
	if returnIdentCanUnlock, returnIdent = f.returnIdentCanUnlock(extParas); returnIdentCanUnlock {
		if !ident.Equal(returnIdent) {
			return false, true
		}
		return true, true
	}

	return true, false
}

// tells whether a sender defined in an expiration unlock condition within this set is the actual
// identity which could unlock an Output containing this UnlockConditionsSet given the ExternalUnlockParameters.
func (f UnlockConditionsSet) returnIdentCanUnlock(extParas *ExternalUnlockParameters) (bool, Address) {
	expUnlockCond := f.Expiration()

	if expUnlockCond == nil {
		return false, nil
	}

	switch {
	case expUnlockCond.MilestoneIndex != 0 && expUnlockCond.UnixTime != 0:
		if expUnlockCond.MilestoneIndex <= extParas.ConfMsIndex && expUnlockCond.UnixTime <= extParas.ConfUnix {
			return true, expUnlockCond.ReturnAddress
		}

	case expUnlockCond.MilestoneIndex != 0:
		if expUnlockCond.MilestoneIndex <= extParas.ConfMsIndex {
			return true, expUnlockCond.ReturnAddress
		}

	case expUnlockCond.UnixTime != 0:
		if expUnlockCond.UnixTime <= extParas.ConfUnix {
			return true, expUnlockCond.ReturnAddress
		}
	}

	return false, nil
}

// TimelocksExpired tells whether UnlockCondition(s) in this set which impose a timelock are expired
// in relation to the given ExternalUnlockParameters.
func (f UnlockConditionsSet) TimelocksExpired(extParas *ExternalUnlockParameters) error {
	timelock := f.Timelock()

	if timelock == nil {
		return nil
	}

	switch {
	case timelock.MilestoneIndex != 0 && extParas.ConfMsIndex < timelock.MilestoneIndex:
		return fmt.Errorf("%w: (ms index) cond %d vs. ext %d", ErrTimelockNotExpired, timelock.MilestoneIndex, extParas.ConfMsIndex)
	case timelock.UnixTime != 0 && extParas.ConfUnix < timelock.UnixTime:
		return fmt.Errorf("%w: (unix) cond %d vs. ext %d", ErrTimelockNotExpired, timelock.UnixTime, extParas.ConfUnix)
	}

	return nil
}

// DustDepositReturn returns the DustDepositReturnUnlockCondition in the set or nil.
func (f UnlockConditionsSet) DustDepositReturn() *DustDepositReturnUnlockCondition {
	b, has := f[UnlockConditionDustDepositReturn]
	if !has {
		return nil
	}
	return b.(*DustDepositReturnUnlockCondition)
}

// Address returns the AddressUnlockCondition in the set or nil.
func (f UnlockConditionsSet) Address() *AddressUnlockCondition {
	b, has := f[UnlockConditionAddress]
	if !has {
		return nil
	}
	return b.(*AddressUnlockCondition)
}

// GovernorAddress returns the GovernorAddressUnlockCondition in the set or nil.
func (f UnlockConditionsSet) GovernorAddress() *GovernorAddressUnlockCondition {
	b, has := f[UnlockConditionGovernorAddress]
	if !has {
		return nil
	}
	return b.(*GovernorAddressUnlockCondition)
}

// StateControllerAddress returns the StateControllerAddressUnlockCondition in the set or nil.
func (f UnlockConditionsSet) StateControllerAddress() *StateControllerAddressUnlockCondition {
	b, has := f[UnlockConditionStateControllerAddress]
	if !has {
		return nil
	}
	return b.(*StateControllerAddressUnlockCondition)
}

// Timelock returns the TimelockUnlockCondition in the set or nil.
func (f UnlockConditionsSet) Timelock() *TimelockUnlockCondition {
	b, has := f[UnlockConditionTimelock]
	if !has {
		return nil
	}
	return b.(*TimelockUnlockCondition)
}

// Expiration returns the ExpirationUnlockCondition in the set or nil.
func (f UnlockConditionsSet) Expiration() *ExpirationUnlockCondition {
	b, has := f[UnlockConditionExpiration]
	if !has {
		return nil
	}
	return b.(*ExpirationUnlockCondition)
}

// UnlockConditionTypeToString returns the name of an UnlockCondition given the type.
func UnlockConditionTypeToString(ty UnlockConditionType) string {
	switch ty {
	case UnlockConditionAddress:
		return "AddressUnlockCondition"
	case UnlockConditionDustDepositReturn:
		return "DustDepositReturnUnlockCondition"
	case UnlockConditionTimelock:
		return "TimelockUnlockCondition"
	case UnlockConditionExpiration:
		return "ExpirationUnlockCondition"
	case UnlockConditionStateControllerAddress:
		return "StateControllerAddressUnlockCondition"
	case UnlockConditionGovernorAddress:
		return "GovernorAddressUnlockCondition"
	}
	return "unknown unlock condition"
}

// UnlockConditionSelector implements SerializableSelectorFunc for unlock conditions.
func UnlockConditionSelector(unlockCondType uint32) (UnlockCondition, error) {
	var seri UnlockCondition
	switch UnlockConditionType(unlockCondType) {
	case UnlockConditionAddress:
		seri = &AddressUnlockCondition{}
	case UnlockConditionDustDepositReturn:
		seri = &DustDepositReturnUnlockCondition{}
	case UnlockConditionTimelock:
		seri = &TimelockUnlockCondition{}
	case UnlockConditionExpiration:
		seri = &ExpirationUnlockCondition{}
	case UnlockConditionStateControllerAddress:
		seri = &StateControllerAddressUnlockCondition{}
	case UnlockConditionGovernorAddress:
		seri = &GovernorAddressUnlockCondition{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownUnlockConditionType, unlockCondType)
	}
	return seri, nil
}

// selects the json object for the given type.
func jsonUnlockConditionSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch UnlockConditionType(ty) {
	case UnlockConditionAddress:
		obj = &jsonAddressUnlockCondition{}
	case UnlockConditionDustDepositReturn:
		obj = &jsonDustDepositReturnUnlockCondition{}
	case UnlockConditionTimelock:
		obj = &jsonTimelockUnlockCondition{}
	case UnlockConditionExpiration:
		obj = &jsonExpirationUnlockCondition{}
	case UnlockConditionStateControllerAddress:
		obj = &jsonStateControllerAddressUnlockCondition{}
	case UnlockConditionGovernorAddress:
		obj = &jsonGovernorAddressUnlockCondition{}
	default:
		return nil, fmt.Errorf("unable to decode unlock condition type from JSON: %w", ErrUnknownUnlockConditionType)
	}
	return obj, nil
}

func unlockConditionsFromJSONRawMsg(jUnlockConditions []*json.RawMessage) (UnlockConditions, error) {
	blocks, err := jsonRawMsgsToSerializables(jUnlockConditions, jsonUnlockConditionSelector)
	if err != nil {
		return nil, err
	}
	var unlockConditions UnlockConditions
	unlockConditions.FromSerializables(blocks)
	return unlockConditions, nil
}
