package iotago

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

// UnlockBlockType defines a type of unlock block.
type UnlockBlockType = byte

const (
	// UnlockBlockSignature denotes a signature unlock block.
	UnlockBlockSignature UnlockBlockType = iota
	// UnlockBlockReference denotes a reference unlock block.
	UnlockBlockReference

	// SignatureUnlockBlockMinSize defines the minimum size of a signature unlock block.
	SignatureUnlockBlockMinSize = serializer.SmallTypeDenotationByteSize + Ed25519SignatureSerializedBytesSize
	// ReferenceUnlockBlockSize defines the size of a reference unlock block.
	ReferenceUnlockBlockSize = serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
)

var (
	// ErrSigUnlockBlocksNotUnique gets returned if unlock blocks making part of a transaction aren't unique.
	ErrSigUnlockBlocksNotUnique = errors.New("signature unlock blocks must be unique")
	// ErrRefUnlockBlockInvalidRef gets returned if a reference unlock block does not reference a signature unlock block.
	ErrRefUnlockBlockInvalidRef = errors.New("reference unlock block must point to a previous signature unlock block")
	// ErrSigUnlockBlockHasNilSig gets returned if a signature unlock block contains a nil signature.
	ErrSigUnlockBlockHasNilSig = errors.New("signature is nil")
)

// UnlockBlockSelector implements SerializableSelectorFunc for unlock block types.
func UnlockBlockSelector(unlockBlockType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch byte(unlockBlockType) {
	case UnlockBlockSignature:
		seri = &SignatureUnlockBlock{}
	case UnlockBlockReference:
		seri = &ReferenceUnlockBlock{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownUnlockBlockType, unlockBlockType)
	}
	return seri, nil
}

// UnlockBlockValidatorFunc which given the index of an unlock block and the unlock block itself, runs validations and returns an error if any should fail.
type UnlockBlockValidatorFunc func(index int, unlockBlock serializer.Serializable) error

// UnlockBlocksSigUniqueAndRefValidator returns a validator which checks that:
//	1. signature unlock blocks are unique
//	2. reference unlock blocks reference a previous signature unlock block
func UnlockBlocksSigUniqueAndRefValidator() UnlockBlockValidatorFunc {
	seenSigBlocks := map[int]struct{}{}
	seenSigBlocksBytes := map[string]int{}

	return func(index int, unlockBlock serializer.Serializable) error {
		switch x := unlockBlock.(type) {
		case *SignatureUnlockBlock:
			if x.Signature == nil {
				return fmt.Errorf("%w: at index %d is nil", ErrSigUnlockBlockHasNilSig, index)
			}

			sigBlockBytes, err := x.Serialize(serializer.DeSeriModeNoValidation)
			if err != nil {
				return fmt.Errorf("unable to serialize signature unlock block at index %d for dup check: %w", index, err)
			}

			if existingIndex, exists := seenSigBlocksBytes[string(sigBlockBytes)]; exists {
				return fmt.Errorf("%w: signature unlock block at index %d is the same as %d", ErrSigUnlockBlocksNotUnique, index, existingIndex)
			}
			seenSigBlocksBytes[string(sigBlockBytes)] = index

			switch x.Signature.(type) {
			case *Ed25519Signature:
				seenSigBlocks[index] = struct{}{}
			default:
				return fmt.Errorf("%w: signature unblock block at index %d holds unknown signature type %T", ErrUnknownSignatureType, index, x)
			}
		case *ReferenceUnlockBlock:
			reference := int(x.Reference)
			if _, has := seenSigBlocks[reference]; !has {
				return fmt.Errorf("%w: %d references non existent unlock block %d", ErrRefUnlockBlockInvalidRef, index, reference)
			}
		default:
			return fmt.Errorf("%w: unlock block at index %d is of unknown type %T", ErrUnknownUnlockBlockType, index, x)
		}

		return nil
	}
}

// ValidateUnlockBlocks validates the unlock blocks by running them against the given UnlockBlockValidatorFunc.
func ValidateUnlockBlocks(unlockBlocks serializer.Serializables, funcs ...UnlockBlockValidatorFunc) error {
	for i, unlockBlock := range unlockBlocks {
		switch unlockBlock.(type) {
		case *SignatureUnlockBlock:
		case *ReferenceUnlockBlock:
		default:
			return fmt.Errorf("%w: can only validate signature or reference unlock blocks", ErrUnknownInputType)
		}
		for _, f := range funcs {
			if err := f(i, unlockBlock); err != nil {
				return err
			}
		}
	}
	return nil
}
