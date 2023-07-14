package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// UnlockType defines a type of unlock.
type UnlockType byte

const (
	// UnlockSignature denotes a SignatureUnlock.
	UnlockSignature UnlockType = iota
	// UnlockReference denotes a ReferenceUnlock.
	UnlockReference
	// UnlockAccount denotes an AccountUnlock.
	UnlockAccount
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
		"AccountUnlock",
		"NFTUnlock",
	}
)

var (
	// ErrSigUnlockNotUnique gets returned if sig unlocks making part of a transaction aren't unique.
	ErrSigUnlockNotUnique = ierrors.New("signature unlock must be unique")
	// ErrReferentialUnlockInvalid gets returned when a ReferentialUnlock is invalid.
	ErrReferentialUnlockInvalid = ierrors.New("invalid referential unlock")
	// ErrSigUnlockHasNilSig gets returned if a signature unlock contains a nil signature.
	ErrSigUnlockHasNilSig = ierrors.New("signature is nil")
)

type Unlocks []Unlock

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

func (o Unlocks) WorkScore(workScoreStructure *WorkScoreStructure) WorkScore {
	var sumCost WorkScore
	for _, u := range o {
		sumCost += u.WorkScore(workScoreStructure)
	}
	return sumCost
}

// UnlocksByType is a map of UnlockType(s) to slice of Unlock(s).
type UnlocksByType map[UnlockType][]Unlock

// Unlock unlocks inputs of a Transaction.
type Unlock interface {
	Sizer
	ProcessableObject

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
//  3. Following through AccountUnlock(s), NFTUnlock(s) refs results to a SignatureUnlock
func UnlocksSigUniqueAndRefValidator(api API) UnlockValidatorFunc {
	seenSigUnlocks := map[uint16]struct{}{}
	seenRefUnlocks := map[uint16]ReferentialUnlock{}
	seenSigUnlockBytes := map[string]int{}
	return func(index int, unlock Unlock) error {
		switch x := unlock.(type) {
		case *SignatureUnlock:
			if x.Signature == nil {
				return ierrors.Wrapf(ErrSigUnlockHasNilSig, "at index %d is nil", index)
			}

			sigBlockBytes, err := api.Encode(x.Signature)
			if err != nil {
				return ierrors.Errorf("unable to serialize signature unlock block at index %d for dup check: %w", index, err)
			}

			if existingIndex, exists := seenSigUnlockBytes[string(sigBlockBytes)]; exists {
				return ierrors.Wrapf(ErrSigUnlockNotUnique, "signature unlock block at index %d is the same as %d", index, existingIndex)
			}

			seenSigUnlockBytes[string(sigBlockBytes)] = index
			seenSigUnlocks[uint16(index)] = struct{}{}
		case ReferentialUnlock:
			if prevRef := seenRefUnlocks[x.Ref()]; prevRef != nil {
				if !x.Chainable() {
					return ierrors.Wrapf(ErrReferentialUnlockInvalid, "%d references existing referential unlock %d but it does not support chaining", index, x.Ref())
				}
				seenRefUnlocks[uint16(index)] = x
				break
			}
			// must reference a sig unlock here
			if _, has := seenSigUnlocks[x.Ref()]; !has {
				return ierrors.Wrapf(ErrReferentialUnlockInvalid, "%d references non existent unlock %d", index, x.Ref())
			}
			seenRefUnlocks[uint16(index)] = x
		default:
			return ierrors.Wrapf(ErrUnknownUnlockType, "unlock at index %d is of unknown type %T", index, x)
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
