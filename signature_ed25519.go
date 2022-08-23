package iotago

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotagoEd25519 "github.com/iotaledger/iota.go/v3/ed25519"
)

const (
	// Ed25519SignatureSerializedBytesSize defines the size of a serialized Ed25519Signature with its type denoting byte and public key.
	Ed25519SignatureSerializedBytesSize = serializer.SmallTypeDenotationByteSize + ed25519.PublicKeySize + ed25519.SignatureSize
)

var (
	// ErrEd25519PubKeyAndAddrMismatch gets returned when an Ed25519Address and public key do not correspond to each other.
	ErrEd25519PubKeyAndAddrMismatch = errors.New("public key and address do not correspond to each other (Ed25519)")
	// ErrEd25519SignatureInvalid gets returned for invalid an Ed25519Signature.
	ErrEd25519SignatureInvalid = errors.New("signature is invalid (Ed25519)")
)

// Ed25519Signature defines an Ed25519 signature.
type Ed25519Signature struct {
	// The public key used to verify the given signature.
	PublicKey [ed25519.PublicKeySize]byte
	// The signature.
	Signature [ed25519.SignatureSize]byte
}

func (e *Ed25519Signature) Type() SignatureType {
	return SignatureEd25519
}

func (e *Ed25519Signature) String() string {
	return fmt.Sprintf("public key: %s, signature: %s", EncodeHex(e.PublicKey[:]), EncodeHex(e.Signature[:]))
}

// Valid verifies whether given the message and Ed25519 address, the signature is valid.
func (e *Ed25519Signature) Valid(msg []byte, addr *Ed25519Address) error {
	// an address is the Blake2b 256 hash of the public key
	addrFromPubKey := Ed25519AddressFromPubKey(e.PublicKey[:])
	if !bytes.Equal(addr[:], addrFromPubKey[:]) {
		return fmt.Errorf("%w: address %s, address from public key %v", ErrEd25519PubKeyAndAddrMismatch, EncodeHex(addr[:]), addrFromPubKey)
	}
	if valid := iotagoEd25519.Verify(e.PublicKey[:], msg, e.Signature[:]); !valid {
		return fmt.Errorf("%w: address %s, public key %v, signature %v", ErrEd25519SignatureInvalid, EncodeHex(addr[:]), EncodeHex(e.PublicKey[:]), EncodeHex(e.Signature[:]))
	}

	return nil
}

func (e *Ed25519Signature) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
		if err := serializer.CheckMinByteLength(Ed25519SignatureSerializedBytesSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid Ed25519 signature bytes: %w", err)
		}
		if err := serializer.CheckTypeByte(data, byte(SignatureEd25519)); err != nil {
			return 0, fmt.Errorf("unable to deserialize Ed25519 signature: %w", err)
		}
	}
	// skip type byte
	data = data[serializer.SmallTypeDenotationByteSize:]
	copy(e.PublicKey[:], data[:ed25519.PublicKeySize])
	copy(e.Signature[:], data[ed25519.PublicKeySize:])

	return Ed25519SignatureSerializedBytesSize, nil
}

func (e *Ed25519Signature) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	var b [Ed25519SignatureSerializedBytesSize]byte
	b[0] = byte(SignatureEd25519)
	copy(b[serializer.SmallTypeDenotationByteSize:], e.PublicKey[:])
	copy(b[serializer.SmallTypeDenotationByteSize+ed25519.PublicKeySize:], e.Signature[:])

	return b[:], nil
}

func (e *Ed25519Signature) Size() int {
	return serializer.SmallTypeDenotationByteSize + ed25519.PublicKeySize + ed25519.SignatureSize
}

func (e *Ed25519Signature) MarshalJSON() ([]byte, error) {
	jEd25519Signature := &jsonEd25519Signature{}
	jEd25519Signature.Type = int(SignatureEd25519)
	jEd25519Signature.PublicKey = EncodeHex(e.PublicKey[:])
	jEd25519Signature.Signature = EncodeHex(e.Signature[:])

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

// jsonEd25519Signature defines the json representation of an Ed25519Signature.
type jsonEd25519Signature struct {
	Type      int    `json:"type"`
	PublicKey string `json:"publicKey"`
	Signature string `json:"signature"`
}

func (j *jsonEd25519Signature) ToSerializable() (serializer.Serializable, error) {
	sig := &Ed25519Signature{}

	pubKeyBytes, err := DecodeHex(j.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("unable to decode public key from JSON for Ed25519 signature: %w", err)
	}

	sigBytes, err := DecodeHex(j.Signature)
	if err != nil {
		return nil, fmt.Errorf("unable to decode signature from JSON for Ed25519 signature: %w", err)
	}

	copy(sig.PublicKey[:], pubKeyBytes)
	copy(sig.Signature[:], sigBytes)

	return sig, nil
}
