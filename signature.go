package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/serializer"

	_ "golang.org/x/crypto/blake2b"
)

// SignatureType defines the type of signature.
type SignatureType = byte

const (
	// SignatureEd25519 denotes an Ed25519Signature.
	SignatureEd25519 SignatureType = iota
	// SignatureBLS denotes a BLSSignature.
	SignatureBLS
)

// SignatureSelector implements SerializableSelectorFunc for signature types.
func SignatureSelector(sigType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch byte(sigType) {
	case SignatureEd25519:
		seri = &Ed25519Signature{}
	case SignatureBLS:
		seri = &BLSSignature{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownSignatureType, sigType)
	}
	return seri, nil
}

// jsonSignatureSelector selects the json signature object for the given type.
func jsonSignatureSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch byte(ty) {
	case SignatureEd25519:
		obj = &jsonEd25519Signature{}
	case SignatureBLS:
		obj = &jsonBLSSignature{}
	default:
		return nil, fmt.Errorf("unable to decode signature type from JSON: %w", ErrUnknownUnlockBlockType)
	}
	return obj, nil
}
