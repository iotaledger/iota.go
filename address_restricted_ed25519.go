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
	PubKeyHash   [Ed25519AddressBytesLength]byte `serix:"0,mapKey=pubKeyHash"`
	Capabilities AddressCapabilitiesBitMask      `serix:"1,mapKey=capabilities,lengthPrefixType=uint8,maxLen=1"`
}

func (addr *RestrictedEd25519Address) Clone() Address {
	cpy := &RestrictedEd25519Address{}
	copy(cpy.PubKeyHash[:], addr.PubKeyHash[:])
	copy(cpy.Capabilities[:], addr.Capabilities[:])

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
		bytes.Equal(addr.Capabilities, otherAddr.Capabilities)
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
		addr.Capabilities.Size()
}

func (addr *RestrictedEd25519Address) CannotReceiveNativeTokens() bool {
	return addr.Capabilities.CannotReceiveNativeTokens()
}

func (addr *RestrictedEd25519Address) CannotReceiveMana() bool {
	return addr.Capabilities.CannotReceiveMana()
}

func (addr *RestrictedEd25519Address) CannotReceiveOutputsWithTimelockUnlockCondition() bool {
	return addr.Capabilities.CannotReceiveOutputsWithTimelockUnlockCondition()
}

func (addr *RestrictedEd25519Address) CannotReceiveOutputsWithExpirationUnlockCondition() bool {
	return addr.Capabilities.CannotReceiveOutputsWithExpirationUnlockCondition()
}

func (addr *RestrictedEd25519Address) CannotReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return addr.Capabilities.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition()
}

func (addr *RestrictedEd25519Address) CannotReceiveAccountOutputs() bool {
	return addr.Capabilities.CannotReceiveAccountOutputs()
}

func (addr *RestrictedEd25519Address) CannotReceiveNFTOutputs() bool {
	return addr.Capabilities.CannotReceiveNFTOutputs()
}

func (addr *RestrictedEd25519Address) CannotReceiveDelegationOutputs() bool {
	return addr.Capabilities.CannotReceiveDelegationOutputs()
}

func (addr *RestrictedEd25519Address) CapabilitiesBitMask() AddressCapabilitiesBitMask {
	return addr.Capabilities
}

// RestrictedEd25519AddressFromPubKey returns the address belonging to the given Ed25519 public key.
func RestrictedEd25519AddressFromPubKey(pubKey ed25519.PublicKey,
	canReceiveNativeTokens bool,
	canReceiveMana bool,
	canReceiveOutputsWithTimelockUnlockCondition bool,
	canReceiveOutputsWithExpirationUnlockCondition bool,
	canReceiveOutputsWithStorageDepositReturnUnlockCondition bool,
	canReceiveAccountOutputs bool,
	canReceiveNFTOutputs bool,
	canReceiveDelegationOutputs bool) *RestrictedEd25519Address {

	address := blake2b.Sum256(pubKey[:])
	addr := &RestrictedEd25519Address{}
	copy(addr.PubKeyHash[:], address[:])

	if canReceiveNativeTokens {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveNativeTokensBitIndex)
	}

	if canReceiveMana {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveManaBitIndex)
	}

	if canReceiveOutputsWithTimelockUnlockCondition {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveOutputsWithTimelockUnlockConditionBitIndex)
	}

	if canReceiveOutputsWithExpirationUnlockCondition {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveOutputsWithExpirationUnlockConditionBitIndex)
	}

	if canReceiveOutputsWithStorageDepositReturnUnlockCondition {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveOutputsWithStorageDepositReturnUnlockConditionBitIndex)
	}

	if canReceiveAccountOutputs {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveAccountOutputsBitIndex)
	}

	if canReceiveNFTOutputs {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveNFTOutputsBitIndex)
	}

	if canReceiveDelegationOutputs {
		addr.Capabilities = addr.Capabilities.setBit(canReceiveDelegationOutputsBitIndex)
	}

	return addr
}
