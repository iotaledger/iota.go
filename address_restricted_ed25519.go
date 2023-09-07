package iotago

import (
	"bytes"
	"context"
	"crypto/ed25519"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

const (
	// RestrictedEd25519AddressMinBytesLength is the min length of a restricted Ed25519 address.
	RestrictedEd25519AddressMinBytesLength        = Ed25519AddressBytesLength + serializer.OneByte
	RestrictedEd25519AddressMaxCapabilitiesLength = 1
)

type RestrictedEd25519Address struct {
	PubKeyHash   [Ed25519AddressBytesLength]byte `serix:"0"`
	Capabilities AddressCapabilitiesBitMask      `serix:"1,lengthPrefixType=uint8,maxLen=1"`
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
	return string(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), redAddr)))
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
	return hexutil.EncodeHex(lo.PanicOnErr(CommonSerixAPI().Encode(context.TODO(), redAddr)))
}

func (redAddr *RestrictedEd25519Address) Size() int {
	return Ed25519AddressSerializedBytesSize +
		redAddr.Capabilities.Size()
}

func (redAddr *RestrictedEd25519Address) CanReceiveNativeTokens() bool {
	return redAddr.Capabilities.CanReceiveNativeTokens()
}

func (redAddr *RestrictedEd25519Address) CanReceiveMana() bool {
	return redAddr.Capabilities.CanReceiveMana()
}

func (redAddr *RestrictedEd25519Address) CanReceiveOutputsWithTimelockUnlockCondition() bool {
	return redAddr.Capabilities.CanReceiveOutputsWithTimelockUnlockCondition()
}

func (redAddr *RestrictedEd25519Address) CanReceiveOutputsWithExpirationUnlockCondition() bool {
	return redAddr.Capabilities.CanReceiveOutputsWithExpirationUnlockCondition()
}

func (redAddr *RestrictedEd25519Address) CanReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return redAddr.Capabilities.CanReceiveOutputsWithStorageDepositReturnUnlockCondition()
}

func (redAddr *RestrictedEd25519Address) CanReceiveAccountOutputs() bool {
	return redAddr.Capabilities.CanReceiveAccountOutputs()
}

func (redAddr *RestrictedEd25519Address) CanReceiveNFTOutputs() bool {
	return redAddr.Capabilities.CanReceiveNFTOutputs()
}

func (redAddr *RestrictedEd25519Address) CanReceiveDelegationOutputs() bool {
	return redAddr.Capabilities.CanReceiveDelegationOutputs()
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
	redAddr := &RestrictedEd25519Address{}
	copy(redAddr.PubKeyHash[:], address[:])

	if canReceiveNativeTokens {
		redAddr.Capabilities = redAddr.Capabilities.setBit(canReceiveNativeTokensBitIndex)
	}

	if canReceiveMana {
		redAddr.Capabilities = redAddr.Capabilities.setBit(canReceiveManaBitIndex)
	}

	if canReceiveOutputsWithTimelockUnlockCondition {
		redAddr.Capabilities = redAddr.Capabilities.setBit(canReceiveOutputsWithTimelockUnlockConditionBitIndex)
	}

	if canReceiveOutputsWithExpirationUnlockCondition {
		redAddr.Capabilities = redAddr.Capabilities.setBit(canReceiveOutputsWithExpirationUnlockConditionBitIndex)
	}

	if canReceiveOutputsWithStorageDepositReturnUnlockCondition {
		redAddr.Capabilities = redAddr.Capabilities.setBit(canReceiveOutputsWithStorageDepositReturnUnlockConditionBitIndex)
	}

	if canReceiveAccountOutputs {
		redAddr.Capabilities = redAddr.Capabilities.setBit(canReceiveAccountOutputsBitIndex)
	}

	if canReceiveNFTOutputs {
		redAddr.Capabilities = redAddr.Capabilities.setBit(canReceiveNFTOutputsBitIndex)
	}

	if canReceiveDelegationOutputs {
		redAddr.Capabilities = redAddr.Capabilities.setBit(canReceiveDelegationOutputsBitIndex)
	}

	return redAddr
}
