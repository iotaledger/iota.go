package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/constraints"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
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
	// UnlockAnchor denotes an AnchorUnlock.
	UnlockAnchor
	// UnlockNFT denotes a NFTUnlock.
	UnlockNFT
	// UnlockMulti denotes a MultiUnlock.
	UnlockMulti
	// UnlockEmpty denotes an EmptyUnlock.
	UnlockEmpty
)

func (unlockType UnlockType) String() string {
	if int(unlockType) >= len(unlockNames) {
		return fmt.Sprintf("unknown unlock type: %d", unlockType)
	}

	return unlockNames[unlockType]
}

var (
	unlockNames = [UnlockEmpty + 1]string{
		"SignatureUnlock",
		"ReferenceUnlock",
		"AccountUnlock",
		"AnchorUnlock",
		"NFTUnlock",
		"MultiUnlock",
		"EmptyUnlock",
	}
)

var (
	// ErrSignatureUnlockNotUnique gets returned if sig unlocks making part of a transaction aren't unique.
	ErrSignatureUnlockNotUnique = ierrors.New("signature unlock must be unique")
	// ErrUnlockSignatureInvalid gets returned when a signature in an unlock is invalid.
	ErrUnlockSignatureInvalid = ierrors.New("signature in unlock is invalid")
	// ErrMultiUnlockNotUnique gets returned if multi unlocks making part of a transaction aren't unique.
	ErrMultiUnlockNotUnique = ierrors.New("multi unlock must be unique")
	// ErrMultiAddressUnlockThresholdNotReached gets returned if multi address unlock threshold was not reached.
	ErrMultiAddressUnlockThresholdNotReached = ierrors.New("multi address unlock threshold not reached")
	// ErrMultiAddressLengthUnlockLengthMismatch gets returned if multi address length and multi unlock length do not match.
	ErrMultiAddressLengthUnlockLengthMismatch = ierrors.New("multi address length and multi unlock length do not match")
	// ErrReferentialUnlockInvalid gets returned when a ReferentialUnlock is invalid.
	ErrReferentialUnlockInvalid = ierrors.New("invalid referential unlock")
	// ErrSignatureUnlockHasNilSignature gets returned if a signature unlock contains a nil signature.
	ErrSignatureUnlockHasNilSignature = ierrors.New("signature is nil")
	// ErrNestedMultiUnlock gets returned when a MultiUnlock is nested inside a MultiUnlock.
	ErrNestedMultiUnlock = ierrors.New("multi unlocks can't be nested")
	// ErrEmptyUnlockOutsideMultiUnlock gets returned when an empty unlock was not nested inside of a multi unlock.
	ErrEmptyUnlockOutsideMultiUnlock = ierrors.New("empty unlocks are only allowed inside of a multi unlock")
	// ErrChainAddressUnlockInvalid gets returned when an invalid unlock for chain addresses is encountered.
	ErrChainAddressUnlockInvalid = ierrors.New("invalid unlock for chain address")
	// ErrDirectUnlockableAddressUnlockInvalid gets returned when an invalid unlock for direct unlockable addresses is encountered.
	ErrDirectUnlockableAddressUnlockInvalid = ierrors.New("invalid unlock for direct unlockable address")
	// ErrMultiAddressUnlockInvalid gets returned when an invalid unlock for multi addresses is encountered.
	ErrMultiAddressUnlockInvalid = ierrors.New("invalid unlock for multi address")
)

type Unlocks []Unlock

func (o Unlocks) Clone() Unlocks {
	return lo.CloneSlice(o)
}

func (o Unlocks) Size() int {
	sum := serializer.UInt16ByteSize
	for _, unlock := range o {
		sum += unlock.Size()
	}

	return sum
}

func (o Unlocks) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	var workScoreUnlocks WorkScore
	for _, unlock := range o {
		workScoreUnlock, err := unlock.WorkScore(workScoreParameters)
		if err != nil {
			return 0, err
		}

		workScoreUnlocks, err = workScoreUnlocks.Add(workScoreUnlock)
		if err != nil {
			return 0, err
		}
	}

	return workScoreUnlocks, nil
}

// Unlock unlocks inputs of a SignedTransaction.
type Unlock interface {
	Sizer
	ProcessableObject
	constraints.Cloneable[Unlock]

	// Type returns the type of the Unlock.
	Type() UnlockType
}

// ReferentialUnlock is an Unlock which references another Unlock.
type ReferentialUnlock interface {
	Unlock

	// ReferencedInputIndex returns the index of the Input/Unlock this ReferentialUnlock references.
	ReferencedInputIndex() uint16
	// Chainable indicates whether this ReferentialUnlock can reference another ReferentialUnlock.
	Chainable() bool
	// SourceAllowed tells whether the given Address is allowed to be the source of this ReferentialUnlock.
	SourceAllowed(address Address) bool
}

// UnlockValidatorFunc which given the index and the Unlock itself, runs validations and returns an error if any should fail.
type UnlockValidatorFunc func(index int, unlock Unlock) error

// SignaturesUniqueAndReferenceUnlocksValidator returns a validator which checks that:
//  1. SignatureUnlock(s) are unique (compared by signer UID)
//     - SignatureUnlock(s) inside different MultiUnlock(s) don't need to be unique,
//     as long as there is no equal SignatureUnlock(s) outside of a MultiUnlock(s).
//  2. ReferenceUnlock(s) reference a previous SignatureUnlock or MultiUnlock
//  3. Following through AccountUnlock(s), AnchorUnlock(s), NFTUnlock(s) refs results to a SignatureUnlock
//  4. EmptyUnlock(s) are only used inside of MultiUnlock(s)
//  5. MultiUnlock(s) are not nested
//  6. MultiUnlock(s) are unique
//  7. ReferenceUnlock(s) to MultiUnlock(s) are not nested in MultiUnlock(s)
func SignaturesUniqueAndReferenceUnlocksValidator(api API) UnlockValidatorFunc {
	// seen signature unlocks and their unlock index
	seenSignatureUnlocks := map[uint16]struct{}{}
	// seen reference unlocks and their unlock index
	seenReferentialUnlocks := map[uint16]ReferentialUnlock{}
	// seen signerUIDs in the signature unlocks
	seenSignerUIDs := map[Identifier]int{}
	// seen signerUIDs in the signature unlocks inside of multi unlocks
	seenSignerUIDsInMultiUnlocks := map[Identifier]int{}
	// seen multi unlocks and their unlock index
	seenMultiUnlocks := map[uint16]struct{}{}
	// seen multi unlock bytes and their unlock index
	seenMultiUnlockBytes := map[string]int{}

	return func(index int, u Unlock) error {
		switch unlock := u.(type) {
		case *SignatureUnlock:
			if unlock.Signature == nil {
				return ierrors.WithMessagef(ErrSignatureUnlockHasNilSignature, "signature at unlock index %d is nil", index)
			}

			signerUID := unlock.Signature.SignerUID()

			// we check for duplicated signer UIDs in SignatureUnlock(s)
			if existingIndex, exists := seenSignerUIDs[signerUID]; exists {
				return ierrors.WithMessagef(ErrSignatureUnlockNotUnique, "signature unlock block at index %d is the same as %d", index, existingIndex)
			}

			// we also need to check for duplicated signer UIDs in MultiUnlock(s)
			if existingIndex, exists := seenSignerUIDsInMultiUnlocks[signerUID]; exists {
				return ierrors.WithMessagef(ErrSignatureUnlockNotUnique, "signature unlock block at index %d is the same as in multi unlock at index %d", index, existingIndex)
			}

			seenSignatureUnlocks[uint16(index)] = struct{}{}
			seenSignerUIDs[signerUID] = index

		case ReferentialUnlock:
			if prevReferentialUnlock := seenReferentialUnlocks[unlock.ReferencedInputIndex()]; prevReferentialUnlock != nil {
				if !unlock.Chainable() {
					return ierrors.WithMessagef(ErrReferentialUnlockInvalid, "unlock at index %d references existing referential unlock %d but it does not support chaining", index, unlock.ReferencedInputIndex())
				}
				seenReferentialUnlocks[uint16(index)] = unlock

				break
			}

			// must reference a sig or multi unlock here
			_, hasSignatureUnlock := seenSignatureUnlocks[unlock.ReferencedInputIndex()]
			_, hasMultiUnlock := seenMultiUnlocks[unlock.ReferencedInputIndex()]
			if !hasSignatureUnlock && !hasMultiUnlock {
				return ierrors.WithMessagef(ErrReferentialUnlockInvalid, "unlock at index %d references non existent unlock %d", index, unlock.ReferencedInputIndex())
			}
			seenReferentialUnlocks[uint16(index)] = unlock

		case *MultiUnlock:
			multiUnlockBytes, err := api.Encode(unlock)
			if err != nil {
				return ierrors.Wrapf(err, "unable to serialize multi unlock block at index %d for dup check", index)
			}

			if existingIndex, exists := seenMultiUnlockBytes[string(multiUnlockBytes)]; exists {
				return ierrors.WithMessagef(ErrMultiUnlockNotUnique, "multi unlock block at index %d is the same as %d", index, existingIndex)
			}

			for subIndex, subU := range unlock.Unlocks {
				switch subUnlock := subU.(type) {
				case *SignatureUnlock:
					if subUnlock.Signature == nil {
						return ierrors.WithMessagef(ErrSignatureUnlockHasNilSignature, "unlock at index %d.%d is nil", index, subIndex)
					}

					signerUID := subUnlock.Signature.SignerUID()

					// we check for duplicated signer UIDs in SignatureUnlock(s)
					if existingIndex, exists := seenSignerUIDs[signerUID]; exists {
						return ierrors.WithMessagef(ErrSignatureUnlockNotUnique, "signature unlock block at index %d.%d is the same as %d", index, subIndex, existingIndex)
					}

					// we don't set the index here in "seenSignatureUnlocks" because there is no concept of reference unlocks inside of multi unlocks

					// add the pubkey to "seenSignerUIDsInMultiUnlocks", so we can check that signer UIDs from a multi unlock are not reused in a normal SignatureUnlock
					seenSignerUIDsInMultiUnlocks[signerUID] = index

				case ReferentialUnlock:
					if prevRef := seenReferentialUnlocks[subUnlock.ReferencedInputIndex()]; prevRef != nil {
						if !subUnlock.Chainable() {
							return ierrors.WithMessagef(ErrReferentialUnlockInvalid, "%d.%d references existing referential unlock %d but it does not support chaining", index, subIndex, subUnlock.ReferencedInputIndex())
						}
						// we don't set the index here in "seenReferentialUnlocks" because it's not allowed to reference an unlock within a multi unlock

						continue
					}
					// must reference a sig unlock here
					// we don't check for "seenMultiUnlocks" here because we don't want to nest "reference unlocks to multi unlocks" in multi unlocks
					if _, has := seenSignatureUnlocks[subUnlock.ReferencedInputIndex()]; !has {
						return ierrors.WithMessagef(ErrReferentialUnlockInvalid, "%d.%d references non existent unlock %d", index, subIndex, subUnlock.ReferencedInputIndex())
					}
					// we don't set the index here in "seenReferentialUnlocks" because it's not allowed to reference an unlock within a multi unlock

				case *MultiUnlock:
					return ierrors.WithMessagef(ErrNestedMultiUnlock, "unlock at index %d.%d is invalid", index, subIndex)

				case *EmptyUnlock:
					// empty unlocks are allowed inside of multi unlocks
					continue

				default:
					panic("all supported unlock types should be handled above")
				}
			}
			seenMultiUnlocks[uint16(index)] = struct{}{}
			seenMultiUnlockBytes[string(multiUnlockBytes)] = index

		case *EmptyUnlock:
			return ierrors.WithMessagef(ErrEmptyUnlockOutsideMultiUnlock, "unlock at index %d is invalid", index)

		default:
			panic("all supported unlock types should be handled above")
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
