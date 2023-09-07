package iotago

import (
	"bytes"
	"context"
	"crypto/ed25519"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

type RestrictedEd25519Address struct {
	PubKeyHash          [Ed25519AddressBytesLength]byte `serix:"0,mapKey=pubKeyHash"`
	AllowedCapabilities AddressCapabilitiesBitMask      `serix:"1,mapKey=allowedCapabilities,lengthPrefixType=uint8,maxLen=1"`
}

func (addr *RestrictedEd25519Address) Clone() Address {
	cpy := &RestrictedEd25519Address{}
	copy(cpy.PubKeyHash[:], addr.PubKeyHash[:])
	copy(cpy.AllowedCapabilities[:], addr.AllowedCapabilities[:])

	return cpy
}

func (addr *RestrictedEd25519Address) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return rentStruct.VBFactorData.Multiply(VBytes(addr.Size()))
}

func (addr *RestrictedEd25519Address) Key() string {
	return string(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr)))
}

func (addr *RestrictedEd25519Address) Unlock(msg []byte, sig Signature) error {
	edSig, isEdSig := sig.(*Ed25519Signature)
	if !isEdSig {
		return ierrors.Wrapf(ErrSignatureAndAddrIncompatible, "can not unlock RestrictedEd25519Address address with signature of type %s", sig.Type())
	}

	ed25519Addr := Ed25519Address(addr.PubKeyHash)
	return edSig.Valid(msg, &ed25519Addr)
}

func (addr *RestrictedEd25519Address) Equal(other Address) bool {
	otherAddr, is := other.(*RestrictedEd25519Address)
	if !is {
		return false
	}

	return addr.PubKeyHash == otherAddr.PubKeyHash &&
		bytes.Equal(addr.AllowedCapabilities, otherAddr.AllowedCapabilities)
}

func (addr *RestrictedEd25519Address) Type() AddressType {
	return AddressRestrictedEd25519
}

func (addr *RestrictedEd25519Address) Bech32(hrp NetworkPrefix) string {
	return bech32String(hrp, addr)
}

func (addr *RestrictedEd25519Address) String() string {
	return hexutil.EncodeHex(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), addr)))
}

func (addr *RestrictedEd25519Address) Size() int {
	return Ed25519AddressSerializedBytesSize +
		addr.AllowedCapabilities.Size()
}

func (addr *RestrictedEd25519Address) CannotReceiveNativeTokens() bool {
	return addr.AllowedCapabilities.CannotReceiveNativeTokens()
}

func (addr *RestrictedEd25519Address) CannotReceiveMana() bool {
	return addr.AllowedCapabilities.CannotReceiveMana()
}

func (addr *RestrictedEd25519Address) CannotReceiveOutputsWithTimelockUnlockCondition() bool {
	return addr.AllowedCapabilities.CannotReceiveOutputsWithTimelockUnlockCondition()
}

func (addr *RestrictedEd25519Address) CannotReceiveOutputsWithExpirationUnlockCondition() bool {
	return addr.AllowedCapabilities.CannotReceiveOutputsWithExpirationUnlockCondition()
}

func (addr *RestrictedEd25519Address) CannotReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return addr.AllowedCapabilities.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition()
}

func (addr *RestrictedEd25519Address) CannotReceiveAccountOutputs() bool {
	return addr.AllowedCapabilities.CannotReceiveAccountOutputs()
}

func (addr *RestrictedEd25519Address) CannotReceiveNFTOutputs() bool {
	return addr.AllowedCapabilities.CannotReceiveNFTOutputs()
}

func (addr *RestrictedEd25519Address) CannotReceiveDelegationOutputs() bool {
	return addr.AllowedCapabilities.CannotReceiveDelegationOutputs()
}

func (addr *RestrictedEd25519Address) AllowedCapabilitiesBitMask() AddressCapabilitiesBitMask {
	return addr.AllowedCapabilities
}

// RestrictedEd25519AddressFromPubKey returns the address belonging to the given Ed25519 public key.
func RestrictedEd25519AddressFromPubKey(pubKey ed25519.PublicKey) *RestrictedEd25519Address {
	address := blake2b.Sum256(pubKey[:])
	addr := &RestrictedEd25519Address{}
	copy(addr.PubKeyHash[:], address[:])

	return addr
}

// RestrictedEd25519AddressFromPubKeyWithCapabilities returns the address belonging to the given Ed25519 public key.
func RestrictedEd25519AddressFromPubKeyWithCapabilities(pubKey ed25519.PublicKey,
	canReceiveNativeTokens bool,
	canReceiveMana bool,
	canReceiveOutputsWithTimelockUnlockCondition bool,
	canReceiveOutputsWithExpirationUnlockCondition bool,
	canReceiveOutputsWithStorageDepositReturnUnlockCondition bool,
	canReceiveAccountOutputs bool,
	canReceiveNFTOutputs bool,
	canReceiveDelegationOutputs bool) *RestrictedEd25519Address {
	addr := RestrictedEd25519AddressFromPubKey(pubKey)
	addr.AllowedCapabilities = AddressCapabilitiesBitMaskWithCapabilities(
		canReceiveNativeTokens,
		canReceiveMana,
		canReceiveOutputsWithTimelockUnlockCondition,
		canReceiveOutputsWithExpirationUnlockCondition,
		canReceiveOutputsWithStorageDepositReturnUnlockCondition,
		canReceiveAccountOutputs,
		canReceiveNFTOutputs,
		canReceiveDelegationOutputs,
	)

	return addr
}
