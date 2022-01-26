package iotago

import (
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

// Signature is a signature.
type Signature interface {
	serializer.Serializable

	// Type returns the type of the Signature.
	Type() SignatureType
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

func signatureFromJSONRawMsg(jRawMsg *json.RawMessage) (Signature, error) {
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
		return nil, fmt.Errorf("unable to decode signature type from JSON: %w", ErrUnknownUnlockBlockType)
	}
	return obj, nil
}
