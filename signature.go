package iotago

import (
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"

	_ "golang.org/x/crypto/blake2b"
)

// SignatureType defines the type of signature.
type SignatureType byte

const (
	// SignatureEd25519 denotes an Ed25519Signature.
	SignatureEd25519 SignatureType = iota
	// SignatureBLS denotes a BLSSignature.
	SignatureBLS
)

// SignatureTypeToString returns the name of a Signature given the type.
func SignatureTypeToString(ty SignatureType) string {
	switch ty {
	case SignatureEd25519:
		return "Ed25519Signature"
	case SignatureBLS:
		return "BLSSignature"
	}
	return "unknown signature"
}

// Signature is a signature.
type Signature interface {
	serializer.Serializable

	// Type returns the type of the Signature.
	Type() SignatureType
}

// SignatureSelector implements SerializableSelectorFunc for signature types.
func SignatureSelector(sigType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch SignatureType(sigType) {
	case SignatureEd25519:
		seri = &Ed25519Signature{}
	case SignatureBLS:
		seri = &BLSSignature{}
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
	case SignatureBLS:
		obj = &jsonBLSSignature{}
	default:
		return nil, fmt.Errorf("unable to decode signature type from JSON: %w", ErrUnknownUnlockBlockType)
	}
	return obj, nil
}
