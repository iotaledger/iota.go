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
	// UnlockNFT denotes a NFTUnlock.
	UnlockNFT
	// UnlockMulti denotes a MultiUnlock.
	UnlockMulti
	// UnlockEmpty denotes an EmptyUnlock.
	UnlockEmpty
	// UnlockAnchor denotes an AnchorUnlock.
	UnlockAnchor
)

func (unlockType UnlockType) String() string {
	if int(unlockType) >= len(unlockNames) {
		return fmt.Sprintf("unknown unlock type: %d", unlockType)
	}

	return unlockNames[unlockType]
}

var (
	unlockNames = [UnlockAnchor + 1]string{
		"SignatureUnlock",
		"ReferenceUnlock",
		"AccountUnlock",
		"NFTUnlock",
		"MultiUnlock",
		"EmptyUnlock",
		"AnchorUnlock",
	}
)

var (
	// ErrSigUnlockNotUnique gets returned if sig unlocks making part of a transaction aren't unique.
	ErrSigUnlockNotUnique = ierrors.New("signature unlock must be unique")
	// ErrMultiUnlockNotUnique gets returned if multi unlocks making part of a transaction aren't unique.
	ErrMultiUnlockNotUnique = ierrors.New("multi unlock must be unique")
	// ErrMultiAddressUnlockThresholdNotReached gets returned if multi address unlock threshold was not reached.
	ErrMultiAddressUnlockThresholdNotReached = ierrors.New("multi address unlock threshold not reached")
	// ErrMultiAddressAndUnlockLengthDoesNotMatch gets returned if multi address length and multi unlock length do not match.
	ErrMultiAddressAndUnlockLengthDoesNotMatch = ierrors.New("multi address length and multi unlock length do not match")
	// ErrReferentialUnlockInvalid gets returned when a ReferentialUnlock is invalid.
	ErrReferentialUnlockInvalid = ierrors.New("invalid referential unlock")
	// ErrSigUnlockHasNilSig gets returned if a signature unlock contains a nil signature.
	ErrSigUnlockHasNilSig = ierrors.New("signature is nil")
	// ErrUnknownUnlockType gets returned for unknown unlock.
	ErrUnknownUnlockType = ierrors.New("unknown unlock type")
	// ErrNestedMultiUnlock gets returned when a MultiUnlock is nested inside a MultiUnlock.
	ErrNestedMultiUnlock = ierrors.New("multi unlocks can't be nested")
	// ErrEmptyUnlockOutsideMultiUnlock gets returned when an empty unlock was not nested inside of a multi unlock.
	ErrEmptyUnlockOutsideMultiUnlock = ierrors.New("empty unlocks are only allowed inside of a multi unlock")
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

func (o Unlocks) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	var workScoreUnlocks WorkScore
	for _, unlock := range o {
		workScoreUnlock, err := unlock.WorkScore(workScoreStructure)
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

	// Ref returns the index of the Unlock this ReferentialUnlock references.
	Ref() uint16
	// Chainable indicates whether this ReferentialUnlock can reference another ReferentialUnlock.
	Chainable() bool
	// SourceAllowed tells whether the given Address is allowed to be the source of this ReferentialUnlock.
	SourceAllowed(address Address) bool
}

// publicKeyBytesFromSignatureBlock returns the bytes of the public key in a signature.
func publicKeyBytesFromSignatureBlock(signature Signature) ([]byte, error) {
	switch sig := signature.(type) {
	case *Ed25519Signature:
		return sig.PublicKey[:], nil
	default:
		return nil, ErrUnknownSignatureType
	}
}

// UnlockValidatorFunc which given the index and the Unlock itself, runs validations and returns an error if any should fail.
type UnlockValidatorFunc func(index int, unlock Unlock) error

// UnlocksSigUniqueAndRefValidator returns a validator which checks that:
//  1. SignatureUnlock(s) are unique (compared by public key)
//     - SignatureUnlock(s) inside different MultiUnlock(s) don't need to be unique,
//     as long as there is no equal SignatureUnlock(s) outside of a MultiUnlock(s).
//  2. ReferenceUnlock(s) reference a previous SignatureUnlock or MultiUnlock
//  3. Following through AccountUnlock(s), NFTUnlock(s), AnchorUnlock(s),  refs results to a SignatureUnlock
//  4. EmptyUnlock(s) are only used inside of MultiUnlock(s)
//  5. MultiUnlock(s) are not nested
//  6. MultiUnlock(s) are unique
//  7. ReferenceUnlock(s) to MultiUnlock(s) are not nested in MultiUnlock(s)
func UnlocksSigUniqueAndRefValidator(api API) UnlockValidatorFunc {
	seenSigUnlocks := map[uint16]struct{}{}
	seenSigBlockPubkeyBytes := map[string]int{}
	seenSigBlockPubkeyBytesInMultiUnlocks := map[string]int{}
	seenRefUnlocks := map[uint16]ReferentialUnlock{}
	seenMultiUnlocks := map[uint16]struct{}{}
	seenMultiUnlockBytes := map[string]int{}

	return func(index int, u Unlock) error {
		switch unlock := u.(type) {
		case *SignatureUnlock:
			if unlock.Signature == nil {
				return ierrors.Wrapf(ErrSigUnlockHasNilSig, "at index %d is nil", index)
			}

			sigBlockPubKeyBytes, err := publicKeyBytesFromSignatureBlock(unlock.Signature)
			if err != nil {
				return ierrors.Wrapf(err, "unable to parse pubkey bytes from signature unlock block at index %d for dup check", index)
			}

			// we check for duplicated pubkeys in SignatureUnlock(s)
			if existingIndex, exists := seenSigBlockPubkeyBytes[string(sigBlockPubKeyBytes)]; exists {
				return ierrors.Wrapf(ErrSigUnlockNotUnique, "signature unlock block at index %d is the same as %d", index, existingIndex)
			}

			// we also need to check for duplicated pubkeys in MultiUnlock(s)
			if existingIndex, exists := seenSigBlockPubkeyBytesInMultiUnlocks[string(sigBlockPubKeyBytes)]; exists {
				return ierrors.Wrapf(ErrSigUnlockNotUnique, "signature unlock block at index %d is the same as in multi unlock at index %d", index, existingIndex)
			}

			seenSigUnlocks[uint16(index)] = struct{}{}
			seenSigBlockPubkeyBytes[string(sigBlockPubKeyBytes)] = index

		case ReferentialUnlock:
			if prevRef := seenRefUnlocks[unlock.Ref()]; prevRef != nil {
				if !unlock.Chainable() {
					return ierrors.Wrapf(ErrReferentialUnlockInvalid, "%d references existing referential unlock %d but it does not support chaining", index, unlock.Ref())
				}
				seenRefUnlocks[uint16(index)] = unlock

				break
			}

			// must reference a sig or multi unlock here
			_, hasSigUnlock := seenSigUnlocks[unlock.Ref()]
			_, hasMultiUnlock := seenMultiUnlocks[unlock.Ref()]
			if !hasSigUnlock && !hasMultiUnlock {
				return ierrors.Wrapf(ErrReferentialUnlockInvalid, "%d references non existent unlock %d", index, unlock.Ref())
			}
			seenRefUnlocks[uint16(index)] = unlock

		case *MultiUnlock:
			multiUnlockBytes, err := api.Encode(unlock)
			if err != nil {
				return ierrors.Errorf("unable to serialize multi unlock block at index %d for dup check: %w", index, err)
			}

			if existingIndex, exists := seenMultiUnlockBytes[string(multiUnlockBytes)]; exists {
				return ierrors.Wrapf(ErrMultiUnlockNotUnique, "multi unlock block at index %d is the same as %d", index, existingIndex)
			}

			for subIndex, subU := range unlock.Unlocks {
				switch subUnlock := subU.(type) {
				case *SignatureUnlock:
					if subUnlock.Signature == nil {
						return ierrors.Wrapf(ErrSigUnlockHasNilSig, "at index %d.%d is nil", index, subIndex)
					}

					sigBlockPubKeyBytes, err := publicKeyBytesFromSignatureBlock(subUnlock.Signature)
					if err != nil {
						return ierrors.Wrapf(err, "unable to parse pubkey bytes from signature unlock block at index %d.%d for dup check", index, subIndex)
					}

					// we check for duplicated pubkeys in SignatureUnlock(s)
					if existingIndex, exists := seenSigBlockPubkeyBytes[string(sigBlockPubKeyBytes)]; exists {
						return ierrors.Wrapf(ErrSigUnlockNotUnique, "signature unlock block at index %d.%d is the same as %d", index, subIndex, existingIndex)
					}

					// we don't set the index here in "seenSigUnlocks" because there is no concept of reference unlocks inside of multi unlocks

					// add the pubkey to "seenSigBlockPubkeyBytesInMultiUnlocks", so we can check that pubkeys from a multi unlock are not reused in a normal SignatureUnlock
					seenSigBlockPubkeyBytesInMultiUnlocks[string(sigBlockPubKeyBytes)] = index

				case ReferentialUnlock:
					if prevRef := seenRefUnlocks[subUnlock.Ref()]; prevRef != nil {
						if !subUnlock.Chainable() {
							return ierrors.Wrapf(ErrReferentialUnlockInvalid, "%d.%d references existing referential unlock %d but it does not support chaining", index, subIndex, subUnlock.Ref())
						}
						// we don't set the index here in "seenRefUnlocks" because it's not allowed to reference an unlock within a multi unlock

						continue
					}
					// must reference a sig unlock here
					// we don't check for "seenMultiUnlocks" here because we don't want to nest "reference unlocks to multi unlocks" in multi unlocks
					if _, has := seenSigUnlocks[subUnlock.Ref()]; !has {
						return ierrors.Wrapf(ErrReferentialUnlockInvalid, "%d.%d references non existent unlock %d", index, subIndex, subUnlock.Ref())
					}
					// we don't set the index here in "seenRefUnlocks" because it's not allowed to reference an unlock within a multi unlock

				case *MultiUnlock:
					return ierrors.Wrapf(ErrNestedMultiUnlock, "unlock at index %d.%d is invalid", index, subIndex)

				case *EmptyUnlock:
					// empty unlocks are allowed inside of multi unlocks
					continue

				default:
					return ierrors.Wrapf(ErrUnknownUnlockType, "unlock at index %d.%d is of unknown type %T", index, subIndex, subUnlock)
				}
			}
			seenMultiUnlocks[uint16(index)] = struct{}{}
			seenMultiUnlockBytes[string(multiUnlockBytes)] = index

		case *EmptyUnlock:
			return ierrors.Wrapf(ErrEmptyUnlockOutsideMultiUnlock, "unlock at index %d is invalid", index)

		default:
			return ierrors.Wrapf(ErrUnknownUnlockType, "unlock at index %d is of unknown type %T", index, unlock)
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
