package iota

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Defines a type of unlock block.
type UnlockBlockType = byte

const (
	// Denotes a signature unlock block.
	UnlockBlockSignature UnlockBlockType = iota
	// Denotes a reference unlock block.
	UnlockBlockReference

	// Defines the minimum size of a signature unlock block.
	SignatureUnlockBlockMinSize = SmallTypeDenotationByteSize + Ed25519SignatureSerializedBytesSize
	// Defines the size of a reference unlock block.
	ReferenceUnlockBlockSize = SmallTypeDenotationByteSize + UInt16ByteSize
)

var (
	// Returned if unlock blocks making part of a transaction aren't unique.
	ErrSigUnlockBlocksNotUnique = errors.New("signature unlock blocks must be unique")
	// Returned if a reference unlock block does not reference a signature unlock block.
	ErrRefUnlockBlockInvalidRef = errors.New("reference unlock block must point to a previous signature unlock block")
	// Returned if a signature unlock block contains a nil signature.
	ErrSigUnlockBlockHasNilSig = errors.New("signature is nil")
)

// UnlockBlockSelector implements SerializableSelectorFunc for unlock block types.
func UnlockBlockSelector(unlockBlockType uint32) (Serializable, error) {
	var seri Serializable
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

// SignatureUnlockBlock holds a signature which unlocks inputs.
type SignatureUnlockBlock struct {
	// The signature of this unlock block.
	Signature Serializable `json:"signature"`
}

func (s *SignatureUnlockBlock) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(SignatureUnlockBlockMinSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid signature unlock block bytes: %w", err)
		}
		if err := checkTypeByte(data, UnlockBlockSignature); err != nil {
			return 0, fmt.Errorf("unable to deserialize signature unlock block: %w", err)
		}
	}

	// skip type byte
	bytesReadTotal := SmallTypeDenotationByteSize
	data = data[SmallTypeDenotationByteSize:]

	sig, sigBytesRead, err := DeserializeObject(data, deSeriMode, TypeDenotationByte, SignatureSelector)
	if err != nil {
		return 0, fmt.Errorf("%w: unable to deserialize signature within signature unlock block", err)
	}
	bytesReadTotal += sigBytesRead
	s.Signature = sig

	return bytesReadTotal, nil
}

func (s *SignatureUnlockBlock) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	sigBytes, err := s.Signature.Serialize(deSeriMode)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to serialize signature within signature unlock block", err)
	}
	return append([]byte{UnlockBlockSignature}, sigBytes...), nil
}

// ReferenceUnlockBlock is an unlock block which references a previous unlock block.
type ReferenceUnlockBlock struct {
	// The other unlock block this reference unlock block references to.
	Reference uint16 `json:"reference"`
}

func (r *ReferenceUnlockBlock) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(ReferenceUnlockBlockSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid reference unlock block bytes: %w", err)
		}
		if err := checkTypeByte(data, UnlockBlockReference); err != nil {
			return 0, fmt.Errorf("unable to deserialize reference unlock block: %w", err)
		}
	}
	data = data[SmallTypeDenotationByteSize:]
	r.Reference = binary.LittleEndian.Uint16(data)
	return ReferenceUnlockBlockSize, nil
}

func (r *ReferenceUnlockBlock) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	var b [ReferenceUnlockBlockSize]byte
	b[0] = UnlockBlockReference
	binary.LittleEndian.PutUint16(b[SmallTypeDenotationByteSize:], r.Reference)
	return b[:], nil
}

// UnlockBlockValidatorFunc which given the index of an unlock block and the unlock block itself, runs validations and returns an error if any should fail.
type UnlockBlockValidatorFunc func(index int, unlockBlock Serializable) error

// UnlockBlocksSigUniqueAndRefValidator returns a validator which checks that:
//	1. signature unlock blocks are unique
//	2. reference unlock blocks reference a previous signature unlock block
func UnlockBlocksSigUniqueAndRefValidator() UnlockBlockValidatorFunc {
	seenEdPubKeys := map[string]int{}
	seenSigBlocks := map[int]struct{}{}
	return func(index int, unlockBlock Serializable) error {
		switch x := unlockBlock.(type) {
		case *SignatureUnlockBlock:
			if x.Signature == nil {
				return fmt.Errorf("%w: at index %d is nil", ErrSigUnlockBlockHasNilSig, index)
			}
			switch y := x.Signature.(type) {
			case *WOTSSignature:
				// TODO: implement
			case *Ed25519Signature:
				k := string(y.PublicKey[:])
				j, has := seenEdPubKeys[k]
				if has {
					return fmt.Errorf("%w: unlock block %d has the same Ed25519 public key as %d", ErrSigUnlockBlocksNotUnique, index, j)
				}
				seenEdPubKeys[k] = index
				seenSigBlocks[index] = struct{}{}
			}
		case *ReferenceUnlockBlock:
			reference := int(x.Reference)
			if _, has := seenSigBlocks[reference]; !has {
				return fmt.Errorf("%w: %d references non existent unlock block %d", ErrRefUnlockBlockInvalidRef, index, reference)
			}
		default:
			return fmt.Errorf("%w: %T", ErrUnknownUnlockBlockType, x)
		}

		return nil
	}
}

// ValidateUnlockBlocks validates the unlock blocks by running them against the given UnlockBlockValidatorFunc.
func ValidateUnlockBlocks(unlockBlocks Serializables, funcs ...UnlockBlockValidatorFunc) error {
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
