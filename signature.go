package iota

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"fmt"

	_ "golang.org/x/crypto/blake2b"
)

// Defines the type of signature.
type SignatureType = uint32

const (
	// Denotes a WOTS signature.
	SignatureWOTS SignatureType = iota
	// Denotes an Ed25519 signature.
	SignatureEd25519

	// The size of a serialized Ed25519 signature with its type denoting byte and public key.
	Ed25519SignatureSerializedBytesSize = TypeDenotationByteSize + ed25519.PublicKeySize + ed25519.SignatureSize
)

var (
	// Returned when an Ed25519 address and public key do not correspond to each other.
	ErrEd25519PubKeyAndAddrMismatch = errors.New("public key and address do not correspond to each other (Ed25519)")
	// Returned for invalid Ed25519 signatures.
	ErrEd25519SignatureInvalid = errors.New("signature is invalid (Ed25519")
)

// SignatureSelector implements SerializableSelectorFunc for signature types.
func SignatureSelector(sigType uint32) (Serializable, error) {
	var seri Serializable
	switch sigType {
	case SignatureWOTS:
		seri = &WOTSSignature{}
	case SignatureEd25519:
		seri = &Ed25519Signature{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownSignatureType, sigType)
	}
	return seri, nil
}

// WOTSSignature defines a WOTS signature.
type WOTSSignature struct{}

func (w *WOTSSignature) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkType(data, SignatureWOTS); err != nil {
			return 0, fmt.Errorf("unable to deserialize WOTS signature: %w", err)
		}
	}
	return 0, ErrWOTSNotImplemented
}

func (w *WOTSSignature) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	panic("implement me")
}

// Ed25519Signature defines an Ed25519 signature.
type Ed25519Signature struct {
	// The public key used to verify the given signature.
	PublicKey [ed25519.PublicKeySize]byte `json:"public_key"`
	// The signature.
	Signature [ed25519.SignatureSize]byte `json:"signature"`
}

// Valid verifies whether given the message and Ed25519 address, the signature is valid.
func (e *Ed25519Signature) Valid(msg []byte, addr *Ed25519Address) error {
	// an address is the Blake2b 256 hash of the public key
	addrFromPubKey := AddressFromEd25519PubKey(e.PublicKey[:])
	if !bytes.Equal(addr[:], addrFromPubKey[:]) {
		return fmt.Errorf("%w: address %s, public key %s", ErrEd25519PubKeyAndAddrMismatch, addr[:], addrFromPubKey)
	}
	if valid := ed25519.Verify(e.PublicKey[:], msg, e.Signature[:]); !valid {
		return fmt.Errorf("%w: address %s, public key %s, signature %s ", ErrEd25519SignatureInvalid, addr[:], e.PublicKey, e.Signature)
	}
	return nil
}

func (e *Ed25519Signature) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(Ed25519SignatureSerializedBytesSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid Ed25519 signature bytes: %w", err)
		}
		if err := checkType(data, SignatureEd25519); err != nil {
			return 0, fmt.Errorf("unable to deserialize Ed25519 signature: %w", err)
		}
	}
	// skip type byte
	data = data[TypeDenotationByteSize:]
	copy(e.PublicKey[:], data[:ed25519.PublicKeySize])
	copy(e.Signature[:], data[ed25519.PublicKeySize:])
	return Ed25519SignatureSerializedBytesSize, nil
}

func (e *Ed25519Signature) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	var b [Ed25519SignatureSerializedBytesSize]byte
	binary.LittleEndian.PutUint32(b[:TypeDenotationByteSize], SignatureEd25519)
	copy(b[TypeDenotationByteSize:], e.PublicKey[:])
	copy(b[TypeDenotationByteSize+ed25519.PublicKeySize:], e.Signature[:])
	return b[:], nil
}
