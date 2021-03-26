package iotago

import (
	"encoding/binary"
	"encoding/json"
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
	return NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := checkMinByteLength(SignatureUnlockBlockMinSize, len(data)); err != nil {
					return fmt.Errorf("invalid signature unlock block bytes: %w", err)
				}
				if err := checkTypeByte(data, UnlockBlockSignature); err != nil {
					return fmt.Errorf("unable to deserialize signature unlock block: %w", err)
				}
			}
			return nil
		}).
		Skip(SmallTypeDenotationByteSize, func(err error) error {
			return fmt.Errorf("unable to skip milestone payload ID during deserialization: %w", err)
		}).
		ReadObject(func(seri Serializable) { s.Signature = seri }, deSeriMode, TypeDenotationByte, SignatureSelector, func(err error) error {
			return fmt.Errorf("unable to deserialize signature within signature unlock block: %w", err)
		}).Done()
}

func (s *SignatureUnlockBlock) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	return NewSerializer().
		WriteNum(UnlockBlockSignature, func(err error) error {
			return fmt.Errorf("unable to serialize signature unlock block type ID: %w", err)
		}).
		WriteObject(s.Signature, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize signature unlock block signature: %w", err)
		}).
		Serialize()
}

func (s *SignatureUnlockBlock) MarshalJSON() ([]byte, error) {
	jSignatureUnlockBlock := &jsonSignatureUnlockBlock{}
	jSignature, err := s.Signature.MarshalJSON()
	if err != nil {
		return nil, err
	}
	rawMsgJsonSig := json.RawMessage(jSignature)
	jSignatureUnlockBlock.Signature = &rawMsgJsonSig
	jSignatureUnlockBlock.Type = int(UnlockBlockSignature)
	return json.Marshal(jSignatureUnlockBlock)
}

func (s *SignatureUnlockBlock) UnmarshalJSON(bytes []byte) error {
	jSignatureUnlockBlock := &jsonSignatureUnlockBlock{}
	if err := json.Unmarshal(bytes, jSignatureUnlockBlock); err != nil {
		return err
	}
	seri, err := jSignatureUnlockBlock.ToSerializable()
	if err != nil {
		return err
	}
	*s = *seri.(*SignatureUnlockBlock)
	return nil
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

func (r *ReferenceUnlockBlock) MarshalJSON() ([]byte, error) {
	jReferenceUnlockBlock := &jsonReferenceUnlockBlock{}
	jReferenceUnlockBlock.Type = int(UnlockBlockReference)
	jReferenceUnlockBlock.Reference = int(r.Reference)
	return json.Marshal(jReferenceUnlockBlock)
}

func (r *ReferenceUnlockBlock) UnmarshalJSON(bytes []byte) error {
	jReferenceUnlockBlock := &jsonReferenceUnlockBlock{}
	if err := json.Unmarshal(bytes, jReferenceUnlockBlock); err != nil {
		return err
	}
	seri, err := jReferenceUnlockBlock.ToSerializable()
	if err != nil {
		return err
	}
	*r = *seri.(*ReferenceUnlockBlock)
	return nil
}

// UnlockBlockValidatorFunc which given the index of an unlock block and the unlock block itself, runs validations and returns an error if any should fail.
type UnlockBlockValidatorFunc func(index int, unlockBlock Serializable) error

// UnlockBlocksSigUniqueAndRefValidator returns a validator which checks that:
//	1. signature unlock blocks are unique
//	2. reference unlock blocks reference a previous signature unlock block
func UnlockBlocksSigUniqueAndRefValidator() UnlockBlockValidatorFunc {
	seenSigBlocks := map[int]struct{}{}
	seenSigBlocksBytes := map[string]int{}

	return func(index int, unlockBlock Serializable) error {
		switch x := unlockBlock.(type) {
		case *SignatureUnlockBlock:
			if x.Signature == nil {
				return fmt.Errorf("%w: at index %d is nil", ErrSigUnlockBlockHasNilSig, index)
			}

			sigBlockBytes, err := x.Serialize(DeSeriModeNoValidation)
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

// jsonUnlockBlockSelector selects the json unlock block object for the given type.
func jsonUnlockBlockSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch byte(ty) {
	case UnlockBlockSignature:
		obj = &jsonSignatureUnlockBlock{}
	case UnlockBlockReference:
		obj = &jsonReferenceUnlockBlock{}
	default:
		return nil, fmt.Errorf("unable to decode unlock block type from JSON: %w", ErrUnknownUnlockBlockType)
	}
	return obj, nil
}

// jsonSignatureUnlockBlock defines the json representation of a SignatureUnlockBlock.
type jsonSignatureUnlockBlock struct {
	Type      int              `json:"type"`
	Signature *json.RawMessage `json:"signature"`
}

func (j *jsonSignatureUnlockBlock) ToSerializable() (Serializable, error) {
	jsonSig, err := DeserializeObjectFromJSON(j.Signature, jsonSignatureSelector)
	if err != nil {
		return nil, err
	}

	sig, err := jsonSig.ToSerializable()
	if err != nil {
		return nil, err
	}

	return &SignatureUnlockBlock{Signature: sig}, nil
}

// jsonReferenceUnlockBlock defines the json representation of a ReferenceUnlockBlock.
type jsonReferenceUnlockBlock struct {
	Type      int `json:"type"`
	Reference int `json:"reference"`
}

func (j *jsonReferenceUnlockBlock) ToSerializable() (Serializable, error) {
	block := &ReferenceUnlockBlock{Reference: uint16(j.Reference)}
	return block, nil
}
