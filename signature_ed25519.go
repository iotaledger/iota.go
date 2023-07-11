package iotago

import (
	"bytes"
	"crypto/ed25519"
	"fmt"

	hiveEd25519 "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	// Ed25519SignatureSerializedBytesSize defines the size of a serialized Ed25519Signature with its type denoting byte and public key.
	Ed25519SignatureSerializedBytesSize = serializer.SmallTypeDenotationByteSize + ed25519.PublicKeySize + ed25519.SignatureSize
)

var (
	// ErrEd25519PubKeyAndAddrMismatch gets returned when an Ed25519Address and public key do not correspond to each other.
	ErrEd25519PubKeyAndAddrMismatch = ierrors.New("public key and address do not correspond to each other (Ed25519)")
	// ErrEd25519SignatureInvalid gets returned for invalid an Ed25519Signature.
	ErrEd25519SignatureInvalid = ierrors.New("signature is invalid (Ed25519)")
)

// Ed25519Signature defines an Ed25519 signature.
type Ed25519Signature struct {
	// The public key used to verify the given signature.
	PublicKey [ed25519.PublicKeySize]byte `serix:"0,mapKey=publicKey"`
	// The signature.
	Signature [ed25519.SignatureSize]byte `serix:"1,mapKey=signature"`
}

func (e *Ed25519Signature) Decode(b []byte) (int, error) {
	copy(e.PublicKey[:], b[:ed25519.PublicKeySize])
	copy(e.Signature[:], b[ed25519.PublicKeySize:])
	return Ed25519SignatureSerializedBytesSize - 1, nil
}

func (e *Ed25519Signature) Encode() ([]byte, error) {
	var b [Ed25519SignatureSerializedBytesSize - 1]byte
	copy(b[:], e.PublicKey[:])
	copy(b[ed25519.PublicKeySize:], e.Signature[:])
	return b[:], nil
}

func (e *Ed25519Signature) Type() SignatureType {
	return SignatureEd25519
}

func (e *Ed25519Signature) String() string {
	return fmt.Sprintf("public key: %s, signature: %s", hexutil.EncodeHex(e.PublicKey[:]), hexutil.EncodeHex(e.Signature[:]))
}

// Valid verifies whether given the message and Ed25519 address, the signature is valid.
func (e *Ed25519Signature) Valid(msg []byte, addr *Ed25519Address) error {
	// an address is the Blake2b 256 hash of the public key
	addrFromPubKey := Ed25519AddressFromPubKey(e.PublicKey[:])
	if !bytes.Equal(addr[:], addrFromPubKey[:]) {
		return ierrors.Wrapf(ErrEd25519PubKeyAndAddrMismatch, "address %s, address from public key %v", hexutil.EncodeHex(addr[:]), hexutil.EncodeHex(addrFromPubKey[:]))
	}
	if valid := hiveEd25519.Verify(e.PublicKey[:], msg, e.Signature[:]); !valid {
		return ierrors.Wrapf(ErrEd25519SignatureInvalid, "address %s, public key %v, signature %v", hexutil.EncodeHex(addr[:]), hexutil.EncodeHex(e.PublicKey[:]), hexutil.EncodeHex(e.Signature[:]))
	}
	return nil
}

func (e *Ed25519Signature) Size() int {
	return serializer.SmallTypeDenotationByteSize + ed25519.PublicKeySize + ed25519.SignatureSize
}
