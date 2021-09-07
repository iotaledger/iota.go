package iotago

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/hive.go/serializer"

	"github.com/iotaledger/iota.go/v2/ed25519"
	_ "golang.org/x/crypto/blake2b"
)

// SignatureType defines the type of signature.
type SignatureType = byte

const (
	// SignatureEd25519 denotes an Ed25519Signature.
	SignatureEd25519 SignatureType = iota

	// Ed25519SignatureSerializedBytesSize defines the size of a serialized Ed25519Signature with its type denoting byte and public key.
	Ed25519SignatureSerializedBytesSize = serializer.SmallTypeDenotationByteSize + ed25519.PublicKeySize + ed25519.SignatureSize
)

var (
	// ErrEd25519PubKeyAndAddrMismatch gets returned when an Ed25519Address and public key do not correspond to each other.
	ErrEd25519PubKeyAndAddrMismatch = errors.New("public key and address do not correspond to each other (Ed25519)")
	// ErrEd25519SignatureInvalid gets returned for invalid an Ed25519Signature.
	ErrEd25519SignatureInvalid = errors.New("signature is invalid (Ed25519")
)

// SignatureSelector implements SerializableSelectorFunc for signature types.
func SignatureSelector(sigType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch byte(sigType) {
	case SignatureEd25519:
		seri = &Ed25519Signature{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownSignatureType, sigType)
	}
	return seri, nil
}

// Ed25519Signature defines an Ed25519 signature.
type Ed25519Signature struct {
	// The public key used to verify the given signature.
	PublicKey [ed25519.PublicKeySize]byte
	// The signature.
	Signature [ed25519.SignatureSize]byte
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

func (e *Ed25519Signature) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
		if err := serializer.CheckMinByteLength(Ed25519SignatureSerializedBytesSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid Ed25519 signature bytes: %w", err)
		}
		if err := serializer.CheckTypeByte(data, SignatureEd25519); err != nil {
			return 0, fmt.Errorf("unable to deserialize Ed25519 signature: %w", err)
		}
	}
	// skip type byte
	data = data[serializer.SmallTypeDenotationByteSize:]
	copy(e.PublicKey[:], data[:ed25519.PublicKeySize])
	copy(e.Signature[:], data[ed25519.PublicKeySize:])
	return Ed25519SignatureSerializedBytesSize, nil
}

func (e *Ed25519Signature) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	var b [Ed25519SignatureSerializedBytesSize]byte
	b[0] = SignatureEd25519
	copy(b[serializer.SmallTypeDenotationByteSize:], e.PublicKey[:])
	copy(b[serializer.SmallTypeDenotationByteSize+ed25519.PublicKeySize:], e.Signature[:])
	return b[:], nil
}

func (e *Ed25519Signature) MarshalJSON() ([]byte, error) {
	jEd25519Signature := &jsonEd25519Signature{}
	jEd25519Signature.Type = int(SignatureEd25519)
	jEd25519Signature.PublicKey = hex.EncodeToString(e.PublicKey[:])
	jEd25519Signature.Signature = hex.EncodeToString(e.Signature[:])
	return json.Marshal(jEd25519Signature)
}

func (e *Ed25519Signature) UnmarshalJSON(bytes []byte) error {
	jEd25519Signature := &jsonEd25519Signature{}
	if err := json.Unmarshal(bytes, jEd25519Signature); err != nil {
		return err
	}
	seri, err := jEd25519Signature.ToSerializable()
	if err != nil {
		return err
	}
	*e = *seri.(*Ed25519Signature)
	return nil
}

// jsonSignatureSelector selects the json signature object for the given type.
func jsonSignatureSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch byte(ty) {
	case SignatureEd25519:
		obj = &jsonEd25519Signature{}
	default:
		return nil, fmt.Errorf("unable to decode signature type from JSON: %w", ErrUnknownUnlockBlockType)
	}
	return obj, nil
}

// jsonEd25519Signature defines the json representation of an Ed25519Signature.
type jsonEd25519Signature struct {
	Type      int    `json:"type"`
	PublicKey string `json:"publicKey"`
	Signature string `json:"signature"`
}

func (j *jsonEd25519Signature) ToSerializable() (serializer.Serializable, error) {
	sig := &Ed25519Signature{}

	pubKeyBytes, err := hex.DecodeString(j.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("unable to decode public key from JSON for Ed25519 signature: %w", err)
	}

	sigBytes, err := hex.DecodeString(j.Signature)
	if err != nil {
		return nil, fmt.Errorf("unable to decode signature from JSON for Ed25519 signature: %w", err)
	}

	copy(sig.PublicKey[:], pubKeyBytes)
	copy(sig.Signature[:], sigBytes)
	return sig, nil
}
