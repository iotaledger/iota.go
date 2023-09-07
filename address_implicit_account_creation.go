package iotago

import (
	"context"
	"crypto/ed25519"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

type ImplicitAccountCreationAddress Ed25519Address

// ParseImplicitAccountCreationAddressFromHexString parses the given hex string into an ImplicitAccountCreationAddress.
func ParseImplicitAccountCreationAddressFromHexString(hexAddr string) (*ImplicitAccountCreationAddress, error) {
	addrBytes, err := hexutil.DecodeHex(hexAddr)
	if err != nil {
		return nil, err
	}

	if len(addrBytes) < Ed25519AddressBytesLength {
		return nil, ierrors.New("invalid ImplicitAccountCreationAddress length")
	}

	addr := &ImplicitAccountCreationAddress{}
	copy(addr[:], addrBytes)

	return addr, nil
}

// MustParseImplicitAccountCreationAddressFromHexString parses the given hex string into an ImplicitAccountCreationAddress.
// It panics if the hex address is invalid.
func MustParseImplicitAccountCreationAddressFromHexString(hexAddr string) *ImplicitAccountCreationAddress {
	addr, err := ParseImplicitAccountCreationAddressFromHexString(hexAddr)
	if err != nil {
		panic(err)
	}

	return addr
}

func (iacAddr *ImplicitAccountCreationAddress) Decode(b []byte) (int, error) {
	copy(iacAddr[:], b)

	return Ed25519AddressSerializedBytesSize - 1, nil
}

func (iacAddr *ImplicitAccountCreationAddress) Encode() ([]byte, error) {
	var b [Ed25519AddressSerializedBytesSize - 1]byte
	copy(b[:], iacAddr[:])

	return b[:], nil
}

func (iacAddr *ImplicitAccountCreationAddress) Clone() Address {
	cpy := &ImplicitAccountCreationAddress{}
	copy(cpy[:], iacAddr[:])

	return cpy
}

func (iacAddr *ImplicitAccountCreationAddress) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(VBytes(iacAddr.Size()))
}

func (iacAddr *ImplicitAccountCreationAddress) Key() string {
	return string(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), iacAddr)))
}

func (iacAddr *ImplicitAccountCreationAddress) Unlock(msg []byte, sig Signature) error {
	edSig, isEdSig := sig.(*Ed25519Signature)
	if !isEdSig {
		return ierrors.Wrapf(ErrSignatureAndAddrIncompatible, "can not unlock ImplicitAccountCreationAddress address with signature of type %s", sig.Type())
	}

	return edSig.Valid(msg, (*Ed25519Address)(iacAddr))
}

func (iacAddr *ImplicitAccountCreationAddress) Equal(other Address) bool {
	otherAddr, is := other.(*ImplicitAccountCreationAddress)
	if !is {
		return false
	}

	return *iacAddr == *otherAddr
}

func (iacAddr *ImplicitAccountCreationAddress) Type() AddressType {
	return AddressImplicitAccountCreation
}

func (iacAddr *ImplicitAccountCreationAddress) Bech32(hrp NetworkPrefix) string {
	return bech32String(hrp, iacAddr)
}

func (iacAddr *ImplicitAccountCreationAddress) String() string {
	return hexutil.EncodeHex(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), iacAddr)))
}

func (iacAddr *ImplicitAccountCreationAddress) Size() int {
	return Ed25519AddressSerializedBytesSize
}

func (iacAddr *ImplicitAccountCreationAddress) CanReceiveNativeTokens() bool {
	return false
}

func (iacAddr *ImplicitAccountCreationAddress) CanReceiveMana() bool {
	return true
}

func (iacAddr *ImplicitAccountCreationAddress) CanReceiveOutputsWithTimelockUnlockCondition() bool {
	return false
}

func (iacAddr *ImplicitAccountCreationAddress) CanReceiveOutputsWithExpirationUnlockCondition() bool {
	return false
}

func (iacAddr *ImplicitAccountCreationAddress) CanReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return false
}

func (iacAddr *ImplicitAccountCreationAddress) CanReceiveAccountOutputs() bool {
	return false
}

func (iacAddr *ImplicitAccountCreationAddress) CanReceiveNFTOutputs() bool {
	return false
}

func (iacAddr *ImplicitAccountCreationAddress) CanReceiveDelegationOutputs() bool {
	return false
}

// ImplicitAccountCreationAddressFromPubKey returns the address belonging to the given Ed25519 public key.
func ImplicitAccountCreationAddressFromPubKey(pubKey ed25519.PublicKey) *ImplicitAccountCreationAddress {
	address := blake2b.Sum256(pubKey[:])

	return (*ImplicitAccountCreationAddress)(&address)
}
