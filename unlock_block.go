package iotago

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
)

// UnlockBlockType defines a type of unlock block.
type UnlockBlockType byte

const (
	// UnlockBlockSignature denotes a SignatureUnlockBlock.
	UnlockBlockSignature UnlockBlockType = iota
	// UnlockBlockReference denotes a ReferenceUnlockBlock.
	UnlockBlockReference
	// UnlockBlockAlias denotes an AliasUnlockBlock.
	UnlockBlockAlias
	// UnlockBlockNFT denotes a NFTUnlockBlock.
	UnlockBlockNFT
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
	switch UnlockBlockType(unlockBlockType) {
	case UnlockBlockSignature:
		seri = &SignatureUnlockBlock{}
	case UnlockBlockReference:
		seri = &ReferenceUnlockBlock{}
	case UnlockBlockAlias:
		seri = &AliasUnlockBlock{}
	case UnlockBlockNFT:
		seri = &NFTUnlockBlock{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownUnlockBlockType, unlockBlockType)
	}
	return seri, nil
}

// UnlockBlockTypeToString returns a name for the given UnlockBlock type.
func UnlockBlockTypeToString(ty UnlockBlockType) string {
	switch ty {
	case UnlockBlockSignature:
		return "SignatureUnlockBlock"
	case UnlockBlockReference:
		return "ReferenceUnlockBlock"
	case UnlockBlockAlias:
		return "AliasUnlockBlock"
	case UnlockBlockNFT:
		return "NFTUnlockBlock"
	default:
		return ""
	}
}

type UnlockBlocks []UnlockBlock

func (o UnlockBlocks) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(o))
	for i, x := range o {
		seris[i] = x.(serializer.Serializable)
	}
	return seris
}

func (o *UnlockBlocks) FromSerializables(seris serializer.Serializables) {
	*o = make(UnlockBlocks, len(seris))
	for i, seri := range seris {
		(*o)[i] = seri.(UnlockBlock)
	}
}

// jsonUnlockBlockSelector selects the json unlock block object for the given type.
func jsonUnlockBlockSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch UnlockBlockType(ty) {
	case UnlockBlockSignature:
		obj = &jsonSignatureUnlockBlock{}
	case UnlockBlockReference:
		obj = &jsonReferenceUnlockBlock{}
	case UnlockBlockAlias:
		obj = &jsonAliasUnlockBlock{}
	case UnlockBlockNFT:
		obj = &jsonNFTUnlockBlock{}
	default:
		return nil, fmt.Errorf("unable to decode unlock block type from JSON: %w", ErrUnknownUnlockBlockType)
	}
	return obj, nil
}

func unlockBlocksFromJSONRawMsg(jUnlockBlocks []*json.RawMessage) (UnlockBlocks, error) {
	blocks, err := jsonRawMsgsToSerializables(jUnlockBlocks, jsonUnlockBlockSelector)
	if err != nil {
		return nil, err
	}
	var unlockB UnlockBlocks
	unlockB.FromSerializables(blocks)
	return unlockB, nil
}

// UnlockBlock is a block of data which unlocks inputs of a Transaction.
type UnlockBlock interface {
	serializer.Serializable

	// Type returns the type of the UnlockBlock.
	Type() UnlockBlockType
}

// ReferentialUnlockBlock is an UnlockBlock which references another UnlockBlock.
type ReferentialUnlockBlock interface {
	UnlockBlock

	// Ref returns the index of the UnlockBlock this ReferentialUnlockBlock references.
	Ref() uint16
}

// UnlockBlockValidatorFunc which given the index of an unlock block and the unlock block itself, runs validations and returns an error if any should fail.
type UnlockBlockValidatorFunc func(index int, unlockBlock UnlockBlock) error

// UnlockBlocksSigUniqueAndRefValidator returns a validator which checks that:
//	1. signature unlock blocks are unique
//	2. reference unlock blocks reference a previous signature unlock block
func UnlockBlocksSigUniqueAndRefValidator() UnlockBlockValidatorFunc {
	seenSigBlocks := map[int]struct{}{}
	seenSigBlocksBytes := map[string]int{}
	return func(index int, unlockBlock UnlockBlock) error {
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
func ValidateUnlockBlocks(unlockBlocks UnlockBlocks, funcs ...UnlockBlockValidatorFunc) error {
	for i, unlockBlock := range unlockBlocks {
		for _, f := range funcs {
			if err := f(i, unlockBlock); err != nil {
				return err
			}
		}
	}
	return nil
}
