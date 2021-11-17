package iotago

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/crypto/bls"
	"github.com/iotaledger/hive.go/serializer"
)

const (
	// BLSSignatureSerializedBytesSize defines the size of a serialized BLSSignature with its type denoting byte and public key.
	BLSSignatureSerializedBytesSize = serializer.SmallTypeDenotationByteSize + bls.PublicKeySize + bls.SignatureSize
)

var (
	// ErrBLSPubKeyAndAddrMismatch gets returned when a BLSAddress and public key do not correspond to each other.
	ErrBLSPubKeyAndAddrMismatch = errors.New("public key and address do not correspond to each other (BLS)")
	// ErrBLSSignatureInvalid gets returned for invalid an BLSSignature.
	ErrBLSSignatureInvalid = errors.New("signature is invalid (BLS)")
)

// BLSSignature defines a BLS signature.
type BLSSignature struct {
	// The public key used to verify the given signature.
	PublicKey [bls.PublicKeySize]byte
	// The signature.
	Signature [bls.SignatureSize]byte
}

func (blsSig *BLSSignature) Type() SignatureType {
	return SignatureBLS
}

// Valid verifies whether given the message and BLS address, the signature is valid.
func (blsSig *BLSSignature) Valid(msg []byte, addr *BLSAddress) error {
	pubKey, _, err := bls.PublicKeyFromBytes(blsSig.PublicKey[:])
	if err != nil {
		return fmt.Errorf("unable to build BLS public keys for validation: %w", err)
	}
	addrFromPubKey := BLSAddressFromPubKey(pubKey)
	if !bytes.Equal(addr[:], addrFromPubKey[:]) {
		return fmt.Errorf("%w: address %s, public key %v", ErrBLSPubKeyAndAddrMismatch, addr[:], addrFromPubKey)
	}
	sig, _, err := bls.SignatureFromBytes(blsSig.Signature[:])
	if err != nil {
		return fmt.Errorf("unable to build BLS signature for validation: %w", err)
	}
	if valid := pubKey.SignatureValid(msg, sig); !valid {
		return fmt.Errorf("%w: address %s, public key %v, signature %s", ErrBLSSignatureInvalid, addr[:], blsSig.PublicKey, blsSig.Signature)
	}
	return nil
}

func (blsSig *BLSSignature) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
		if err := serializer.CheckMinByteLength(BLSSignatureSerializedBytesSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid BLS signature bytes: %w", err)
		}
		if err := serializer.CheckTypeByte(data, byte(SignatureBLS)); err != nil {
			return 0, fmt.Errorf("unable to deserialize BLS signature: %w", err)
		}
	}
	// skip type byte
	data = data[serializer.SmallTypeDenotationByteSize:]
	copy(blsSig.PublicKey[:], data[:bls.PublicKeySize])
	copy(blsSig.Signature[:], data[bls.PublicKeySize:])
	return BLSSignatureSerializedBytesSize, nil
}

func (blsSig *BLSSignature) Serialize(_ serializer.DeSerializationMode) ([]byte, error) {
	var b [BLSSignatureSerializedBytesSize]byte
	b[0] = byte(SignatureBLS)
	copy(b[serializer.SmallTypeDenotationByteSize:], blsSig.PublicKey[:])
	copy(b[serializer.SmallTypeDenotationByteSize+bls.PublicKeySize:], blsSig.Signature[:])
	return b[:], nil
}

func (blsSig *BLSSignature) MarshalJSON() ([]byte, error) {
	jBLSSignature := &jsonBLSSignature{}
	jBLSSignature.Type = int(SignatureBLS)
	jBLSSignature.PublicKey = hex.EncodeToString(blsSig.PublicKey[:])
	jBLSSignature.Signature = hex.EncodeToString(blsSig.Signature[:])
	return json.Marshal(jBLSSignature)
}

func (blsSig *BLSSignature) UnmarshalJSON(bytes []byte) error {
	jBLSSignature := &jsonBLSSignature{}
	if err := json.Unmarshal(bytes, jBLSSignature); err != nil {
		return err
	}
	seri, err := jBLSSignature.ToSerializable()
	if err != nil {
		return err
	}
	*blsSig = *seri.(*BLSSignature)
	return nil
}

// jsonBLSSignature defines the json representation of a BLSSignature.
type jsonBLSSignature struct {
	Type      int    `json:"type"`
	PublicKey string `json:"publicKey"`
	Signature string `json:"signature"`
}

func (j *jsonBLSSignature) ToSerializable() (serializer.Serializable, error) {
	sig := &BLSSignature{}

	pubKeyBytes, err := hex.DecodeString(j.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("unable to decode public key from JSON for BLS signature: %w", err)
	}

	sigBytes, err := hex.DecodeString(j.Signature)
	if err != nil {
		return nil, fmt.Errorf("unable to decode signature from JSON for BLS signature: %w", err)
	}

	copy(sig.PublicKey[:], pubKeyBytes)
	copy(sig.Signature[:], sigBytes)
	return sig, nil
}
