package iotago

import (
	"context"
	"crypto/ed25519"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	ImplicitAccountCreationAddressBytesLength         = Ed25519AddressBytesLength
	ImplicitAccountCreationAddressSerializedBytesSize = Ed25519AddressSerializedBytesSize
)

type ImplicitAccountCreationAddress Ed25519Address

// ParseImplicitAccountCreationAddressFromHexString parses the given hex string into an ImplicitAccountCreationAddress.
func ParseImplicitAccountCreationAddressFromHexString(hexAddr string) (*ImplicitAccountCreationAddress, error) {
	addrBytes, err := hexutil.DecodeHex(hexAddr)
	if err != nil {
		return nil, err
	}

	if len(addrBytes) < ImplicitAccountCreationAddressBytesLength {
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

func (addr *ImplicitAccountCreationAddress) Clone() Address {
	cpy := &ImplicitAccountCreationAddress{}
	copy(cpy[:], addr[:])

	return cpy
}

func (addr *ImplicitAccountCreationAddress) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(VBytes(addr.Size()))
}

func (addr *ImplicitAccountCreationAddress) ID() []byte {
	return lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr))
}

func (addr *ImplicitAccountCreationAddress) Key() string {
	return string(addr.ID())
}

func (addr *ImplicitAccountCreationAddress) Unlock(msg []byte, sig Signature) error {
	edSig, isEdSig := sig.(*Ed25519Signature)
	if !isEdSig {
		return ierrors.Wrapf(ErrSignatureAndAddrIncompatible, "can not unlock ImplicitAccountCreationAddress address with signature of type %s", sig.Type())
	}

	return edSig.Valid(msg, (*Ed25519Address)(addr))
}

func (addr *ImplicitAccountCreationAddress) Equal(other Address) bool {
	otherAddr, is := other.(*ImplicitAccountCreationAddress)
	if !is {
		return false
	}

	return *addr == *otherAddr
}

func (addr *ImplicitAccountCreationAddress) Type() AddressType {
	return AddressImplicitAccountCreation
}

func (addr *ImplicitAccountCreationAddress) Bech32(hrp NetworkPrefix) string {
	return bech32StringAddress(hrp, addr)
}

func (addr *ImplicitAccountCreationAddress) String() string {
	return hexutil.EncodeHex(addr.ID())
}

func (addr *ImplicitAccountCreationAddress) Size() int {
	return ImplicitAccountCreationAddressSerializedBytesSize
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveNativeTokens() bool {
	return true
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveMana() bool {
	return false
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveOutputsWithTimelockUnlockCondition() bool {
	return true
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveOutputsWithExpirationUnlockCondition() bool {
	return true
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return true
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveAccountOutputs() bool {
	return true
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveNFTOutputs() bool {
	return true
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveDelegationOutputs() bool {
	return true
}

// ImplicitAccountCreationAddressFromPubKey returns the address belonging to the given Ed25519 public key.
func ImplicitAccountCreationAddressFromPubKey(pubKey ed25519.PublicKey) *ImplicitAccountCreationAddress {
	address := blake2b.Sum256(pubKey[:])

	return (*ImplicitAccountCreationAddress)(&address)
}
