package iotago

import (
	"bytes"
	"crypto/ed25519"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	// RestrictedEd25519AddressMinBytesLength is the min length of a restricted Ed25519 address.
	RestrictedEd25519AddressMinBytesLength = Ed25519AddressBytesLength + serializer.OneByte
)

type RestrictedEd25519Address struct {
	PubKeyHash   [Ed25519AddressBytesLength]byte `serix:"0"`
	Capabilities []byte                          `serix:"1,lengthPrefixType=uint8"`
}

// ParseRestrictedEd25519AddressFromHexString parses the given hex string into an RestrictedEd25519Address.
func ParseRestrictedEd25519AddressFromHexString(hexAddr string) (*RestrictedEd25519Address, error) {
	addrBytes, err := hexutil.DecodeHex(hexAddr)
	if err != nil {
		return nil, err
	}

	if len(addrBytes) < RestrictedEd25519AddressMinBytesLength {
		return nil, ierrors.New("invalid RestrictedEd25519Address length")
	}

	addr := &RestrictedEd25519Address{}
	copy(addr.PubKeyHash[:], addrBytes[:Ed25519AddressBytesLength])

	capabilitiesLength := int(addrBytes[Ed25519AddressBytesLength])
	if len(addrBytes) < RestrictedEd25519AddressMinBytesLength+capabilitiesLength {
		return nil, ierrors.New("invalid RestrictedEd25519Address length")
	}

	copy(addr.Capabilities[:], addrBytes[Ed25519AddressBytesLength:capabilitiesLength])

	return addr, nil
}

// MustParseRestrictedEd25519AddressFromHexString parses the given hex string into an RestrictedEd25519Address.
// It panics if the hex address is invalid.
func MustParseRestrictedEd25519AddressFromHexString(hexAddr string) *RestrictedEd25519Address {
	addr, err := ParseRestrictedEd25519AddressFromHexString(hexAddr)
	if err != nil {
		panic(err)
	}

	return addr
}

func (redAddr *RestrictedEd25519Address) Clone() Address {
	cpy := &RestrictedEd25519Address{}
	copy(cpy.PubKeyHash[:], redAddr.PubKeyHash[:])
	copy(cpy.Capabilities[:], redAddr.Capabilities[:])

	return cpy
}

func (redAddr *RestrictedEd25519Address) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(VBytes(redAddr.Size()))
}

func (redAddr *RestrictedEd25519Address) Key() string {
	return string(
		byteutils.ConcatBytes(
			[]byte{byte(AddressRestrictedEd25519)},
			redAddr.PubKeyHash[:],
			[]byte{byte(len(redAddr.Capabilities))},
			redAddr.Capabilities,
		),
	)
}

func (redAddr *RestrictedEd25519Address) Unlock(msg []byte, sig Signature) error {
	edSig, isEdSig := sig.(*Ed25519Signature)
	if !isEdSig {
		return ierrors.Wrapf(ErrSignatureAndAddrIncompatible, "can not unlock RestrictedEd25519Address address with signature of type %s", sig.Type())
	}

	addr := Ed25519Address(redAddr.PubKeyHash)
	return edSig.Valid(msg, &addr)
}

func (redAddr *RestrictedEd25519Address) Equal(other Address) bool {
	otherAddr, is := other.(*RestrictedEd25519Address)
	if !is {
		return false
	}

	return redAddr.PubKeyHash == otherAddr.PubKeyHash &&
		bytes.Equal(redAddr.Capabilities, otherAddr.Capabilities)
}

func (redAddr *RestrictedEd25519Address) Type() AddressType {
	return AddressRestrictedEd25519
}

func (redAddr *RestrictedEd25519Address) Bech32(hrp NetworkPrefix) string {
	return bech32String(hrp, redAddr)
}

func (redAddr *RestrictedEd25519Address) String() string {
	return hexutil.EncodeHex(
		byteutils.ConcatBytes(
			redAddr.PubKeyHash[:],
			[]byte{byte(len(redAddr.Capabilities))},
			redAddr.Capabilities,
		),
	)
}

func (redAddr *RestrictedEd25519Address) Size() int {
	return Ed25519AddressSerializedBytesSize +
		serializer.SmallTypeDenotationByteSize +
		len(redAddr.Capabilities)
}

// RestrictedEd25519AddressFromPubKey returns the address belonging to the given Ed25519 public key.
func RestrictedEd25519AddressFromPubKey(pubKey ed25519.PublicKey) *RestrictedEd25519Address {
	address := blake2b.Sum256(pubKey[:])
	redAddr := &RestrictedEd25519Address{}
	copy(redAddr.PubKeyHash[:], address[:])

	return redAddr
}
