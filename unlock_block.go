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
	// ErrRefUnlockBlockInvalidRef gets returned if a ReferentialUnlockBlock does not reference a SignatureUnlockBlock.
	ErrRefUnlockBlockInvalidRef = errors.New("referential unlock block must point to a previous signature unlock block")
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

// ToUnlockBlocksByType converts the UnlockBlocks slice to UnlockBlocksByType.
func (o UnlockBlocks) ToUnlockBlocksByType() UnlockBlocksByType {
	unlockBlocksByType := make(UnlockBlocksByType)
	for _, unlockBlock := range o {
		slice, has := unlockBlocksByType[unlockBlock.Type()]
		if !has {
			slice = make(UnlockBlocks, 0)
		}
		unlockBlocksByType[unlockBlock.Type()] = append(slice, unlockBlock)
	}
	return unlockBlocksByType
}

// UnlockBlocksByType is a map of UnlockBlockType(s) to slice of UnlockBlock(s).
type UnlockBlocksByType map[UnlockBlockType][]UnlockBlock

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

// UnlockBlockValidatorFunc which given the index and the UnlockBlock itself, runs validations and returns an error if any should fail.
type UnlockBlockValidatorFunc func(index int, unlockBlock UnlockBlock) error

// UnlockBlocksSigUniqueAndRefValidator returns a validator which checks that:
//	1. SignatureUnlockBlock(s) are unique
//	2. ReferenceUnlockBlock(s), AliasUnlockBlock(s), NFTUnlockBlock(s) reference a previous SignatureUnlockBlock
func UnlockBlocksSigUniqueAndRefValidator() UnlockBlockValidatorFunc {
	seenSigBlocks := map[uint16]struct{}{}
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
			seenSigBlocks[uint16(index)] = struct{}{}
		case ReferentialUnlockBlock:
			if _, has := seenSigBlocks[x.Ref()]; !has {
				return fmt.Errorf("%w: %d references non existent unlock block %d", ErrRefUnlockBlockInvalidRef, index, x.Ref())
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
