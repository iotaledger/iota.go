package iotago

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"

	_ "golang.org/x/crypto/blake2b"
)

// SignatureType defines the type of signature.
type SignatureType byte

const (
	// SignatureEd25519 denotes an Ed25519Signature.
	SignatureEd25519 SignatureType = iota
)

func (sigType SignatureType) String() string {
	if int(sigType) >= len(sigNames) {
		return fmt.Sprintf("unknown signature type: %d", sigType)
	}
	return sigNames[sigType]
}

var (
	sigNames = [SignatureEd25519 + 1]string{"Ed25519Signature"}
	// ErrTypeIsNotSupportedSignature gets returned when a serializable was found to not be a supported Signature.
	ErrTypeIsNotSupportedSignature = errors.New("serializable is not a supported signature")
)

// Signatures is a slice of Signature(s).
type Signatures []Signature

func (sigs Signatures) Len() int {
	return len(sigs)
}

func (sigs Signatures) Less(i, j int) bool {
	// change this once there are more signature types
	a, b := sigs[i].(*Ed25519Signature), sigs[j].(*Ed25519Signature)

	cmp := bytes.Compare(a.PublicKey[:], b.PublicKey[:])
	if cmp == 0 {
		return bytes.Compare(a.Signature[:], b.Signature[:]) == -1
	}

	return cmp == -1
}

func (sigs Signatures) Swap(i, j int) {
	sigs[i], sigs[j] = sigs[j], sigs[i]
}

func (sigs Signatures) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(sigs))
	for i, x := range sigs {
		seris[i] = x.(serializer.Serializable)
	}
	return seris
}

func (sigs *Signatures) FromSerializables(seris serializer.Serializables) {
	*sigs = make(Signatures, len(seris))
	for i, seri := range seris {
		(*sigs)[i] = seri.(Signature)
	}
}

func signaturesFromJSONRawMsg(jSignatures []*json.RawMessage) (Signatures, error) {
	sigs, err := jsonRawMsgsToSerializables(jSignatures, jsonSignatureSelector)
	if err != nil {
		return nil, err
	}
	var signatures Signatures
	signatures.FromSerializables(sigs)
	return signatures, nil
}

// Signature is a signature.
type Signature interface {
	serializer.SerializableWithSize

	// Type returns the type of the Signature.
	Type() SignatureType
}

// SignatureTypeSet is a set of SignatureType.
type SignatureTypeSet map[SignatureType]struct{}

// checks whether the given Serializable is a Signature and also supported SignatureType.
func SignatureWriteGuard(supportedSigs SignatureTypeSet) serializer.SerializableWriteGuardFunc {
	return func(seri serializer.Serializable) error {
		if seri == nil {
			return fmt.Errorf("%w: because nil", ErrTypeIsNotSupportedSignature)
		}

		sig, is := seri.(Signature)
		if !is {
			return fmt.Errorf("%w: because not signature", ErrTypeIsNotSupportedSignature)
		}

		if _, supported := supportedSigs[sig.Type()]; !supported {
			return fmt.Errorf("%w: because not in set %v", ErrTypeIsNotSupportedSignature, supported)
		}

		return nil
	}
}

func SignatureReadGuard(supportedSigs SignatureTypeSet) serializer.SerializableReadGuardFunc {
	return func(ty uint32) (serializer.Serializable, error) {
		if _, supported := supportedSigs[SignatureType(ty)]; !supported {
			return nil, fmt.Errorf("%w: because not in set %v (%d)", ErrTypeIsNotSupportedSignature, supportedSigs, ty)
		}
		return SignatureSelector(ty)
	}
}

// SignatureSelector implements SerializableSelectorFunc for signature types.
func SignatureSelector(sigType uint32) (Signature, error) {
	var seri Signature
	switch SignatureType(sigType) {
	case SignatureEd25519:
		seri = &Ed25519Signature{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownSignatureType, sigType)
	}
	return seri, nil
}

func SignatureFromJSONRawMsg(jRawMsg *json.RawMessage) (Signature, error) {
	jsonSignature, err := DeserializeObjectFromJSON(jRawMsg, jsonSignatureSelector)
	if err != nil {
		return nil, fmt.Errorf("can't decode signature type from JSON: %w", err)
	}

	addr, err := jsonSignature.ToSerializable()
	if err != nil {
		return nil, err
	}
	return addr.(Signature), nil
}

// jsonSignatureSelector selects the json signature object for the given type.
func jsonSignatureSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch SignatureType(ty) {
	case SignatureEd25519:
		obj = &jsonEd25519Signature{}
	default:
		return nil, fmt.Errorf("unable to decode signature type from JSON: %w", ErrUnknownUnlockType)
	}
	return obj, nil
}
