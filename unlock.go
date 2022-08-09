package iotago

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// UnlockType defines a type of unlock.
type UnlockType byte

const (
	// UnlockSignature denotes a SignatureUnlock.
	UnlockSignature UnlockType = iota
	// UnlockReference denotes a ReferenceUnlock.
	UnlockReference
	// UnlockAlias denotes an AliasUnlock.
	UnlockAlias
	// UnlockNFT denotes a NFTUnlock.
	UnlockNFT
)

func (unlockType UnlockType) String() string {
	if int(unlockType) >= len(unlockNames) {
		return fmt.Sprintf("unknown unlock type: %d", unlockType)
	}
	return unlockNames[unlockType]
}

var (
	unlockNames = [UnlockNFT + 1]string{
		"SignatureUnlock",
		"ReferenceUnlock",
		"AliasUnlock",
		"NFTUnlock",
	}
)

var (
	// ErrSigUnlockNotUnique gets returned if sig unlocks making part of a transaction aren't unique.
	ErrSigUnlockNotUnique = errors.New("signature unlock must be unique")
	// ErrReferentialUnlockInvalid gets returned when a ReferentialUnlock is invalid.
	ErrReferentialUnlockInvalid = errors.New("invalid referential unlock")
	// ErrSigUnlockHasNilSig gets returned if a signature unlock contains a nil signature.
	ErrSigUnlockHasNilSig = errors.New("signature is nil")
	// ErrTypeIsNotSupportedUnlock gets returned when a serializable was found to not be a supported Unlock.
	ErrTypeIsNotSupportedUnlock = errors.New("serializable is not a supported unlock")
)

// UnlockSelector implements SerializableSelectorFunc for unlock types.
func UnlockSelector(unlockType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch UnlockType(unlockType) {
	case UnlockSignature:
		seri = &SignatureUnlock{}
	case UnlockReference:
		seri = &ReferenceUnlock{}
	case UnlockAlias:
		seri = &AliasUnlock{}
	case UnlockNFT:
		seri = &NFTUnlock{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownUnlockType, unlockType)
	}
	return seri, nil
}

type Unlocks []Unlock

func (o Unlocks) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(o))
	for i, x := range o {
		seris[i] = x.(serializer.Serializable)
	}
	return seris
}

func (o *Unlocks) FromSerializables(seris serializer.Serializables) {
	*o = make(Unlocks, len(seris))
	for i, seri := range seris {
		(*o)[i] = seri.(Unlock)
	}
}

// ToUnlockByType converts the Unlocks slice to UnlocksByType.
func (o Unlocks) ToUnlockByType() UnlocksByType {
	unlocksByType := make(UnlocksByType)
	for _, unlock := range o {
		slice, has := unlocksByType[unlock.Type()]
		if !has {
			slice = make(Unlocks, 0)
		}
		unlocksByType[unlock.Type()] = append(slice, unlock)
	}
	return unlocksByType
}

func (o Unlocks) Size() int {
	sum := serializer.UInt16ByteSize
	for _, unlock := range o {
		sum += unlock.Size()
	}
	return sum
}

// UnlocksByType is a map of UnlockType(s) to slice of Unlock(s).
type UnlocksByType map[UnlockType][]Unlock

func unlockWriteGuard() serializer.SerializableWriteGuardFunc {
	return func(seri serializer.Serializable) error {
		if seri == nil {
			return fmt.Errorf("%w: because nil", ErrTypeIsNotSupportedUnlock)
		}
		switch seri.(type) {
		case *SignatureUnlock:
		case *ReferenceUnlock:
		case *AliasUnlock:
		case *NFTUnlock:
		default:
			return ErrTypeIsNotSupportedUnlock
		}
		return nil
	}
}

// jsonUnlockSelector selects the json unlock object for the given type.
func jsonUnlockSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch UnlockType(ty) {
	case UnlockSignature:
		obj = &jsonSignatureUnlock{}
	case UnlockReference:
		obj = &jsonReferenceUnlock{}
	case UnlockAlias:
		obj = &jsonAliasUnlock{}
	case UnlockNFT:
		obj = &jsonNFTUnlock{}
	default:
		return nil, fmt.Errorf("unable to decode unlock type from JSON: %w", ErrUnknownUnlockType)
	}
	return obj, nil
}

func unlocksFromJSONRawMsg(jUnlocks []*json.RawMessage) (Unlocks, error) {
	unlocks, err := jsonRawMsgsToSerializables(jUnlocks, jsonUnlockSelector)
	if err != nil {
		return nil, err
	}
	var unlockB Unlocks
	unlockB.FromSerializables(unlocks)
	return unlockB, nil
}

// Unlock unlocks inputs of a Transaction.
type Unlock interface {
	serializer.SerializableWithSize

	// Type returns the type of the Unlock.
	Type() UnlockType
}

// ReferentialUnlock is an Unlock which references another Unlock.
type ReferentialUnlock interface {
	Unlock

	// Ref returns the index of the Unlock this ReferentialUnlock references.
	Ref() uint16
	// Chainable indicates whether this ReferentialUnlock can reference another ReferentialUnlock.
	Chainable() bool
	// SourceAllowed tells whether the given Address is allowed to be the source of this ReferentialUnlock.
	SourceAllowed(address Address) bool
}

// UnlockValidatorFunc which given the index and the Unlock itself, runs validations and returns an error if any should fail.
type UnlockValidatorFunc func(index int, unlock Unlock) error

// UnlocksSigUniqueAndRefValidator returns a validator which checks that:
//  1. SignatureUnlock(s) are unique
//  2. ReferenceUnlock(s) reference a previous SignatureUnlock
//  3. Following through AliasUnlock(s), NFTUnlock(s) refs results to a SignatureUnlock
func UnlocksSigUniqueAndRefValidator() UnlockValidatorFunc {
	seenSigUnlocks := map[uint16]struct{}{}
	seenRefUnlocks := map[uint16]ReferentialUnlock{}
	seenSigUnlockBytes := map[string]int{}
	return func(index int, unlock Unlock) error {
		switch x := unlock.(type) {
		case *SignatureUnlock:
			if x.Signature == nil {
				return fmt.Errorf("%w: at index %d is nil", ErrSigUnlockHasNilSig, index)
			}

			sigUnlockBytes, err := x.Serialize(serializer.DeSeriModeNoValidation, nil)
			if err != nil {
				return fmt.Errorf("unable to serialize signature unlock at index %d for dup check: %w", index, err)
			}

			if existingIndex, exists := seenSigUnlockBytes[string(sigUnlockBytes)]; exists {
				return fmt.Errorf("%w: signature unlock at index %d is the same as %d", ErrSigUnlockNotUnique, index, existingIndex)
			}

			seenSigUnlockBytes[string(sigUnlockBytes)] = index
			seenSigUnlocks[uint16(index)] = struct{}{}
		case ReferentialUnlock:
			if prevRef := seenRefUnlocks[x.Ref()]; prevRef != nil {
				if !x.Chainable() {
					return fmt.Errorf("%w: %d references existing referential unlock %d but it does not support chaining", ErrReferentialUnlockInvalid, index, x.Ref())
				}
				seenRefUnlocks[uint16(index)] = x
				break
			}
			// must reference a sig unlock here
			if _, has := seenSigUnlocks[x.Ref()]; !has {
				return fmt.Errorf("%w: %d references non existent unlock %d", ErrReferentialUnlockInvalid, index, x.Ref())
			}
			seenRefUnlocks[uint16(index)] = x
		default:
			return fmt.Errorf("%w: unlock at index %d is of unknown type %T", ErrUnknownUnlockType, index, x)
		}

		return nil
	}
}

// ValidateUnlocks validates the unlocks by running them against the given UnlockValidatorFunc.
func ValidateUnlocks(unlocks Unlocks, funcs ...UnlockValidatorFunc) error {
	for i, unlock := range unlocks {
		for _, f := range funcs {
			if err := f(i, unlock); err != nil {
				return err
			}
		}
	}
	return nil
}
