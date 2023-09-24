package iotago

import (
	"context"

	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	canReceiveNativeTokensBitIndex = iota
	canReceiveManaBitIndex
	canReceiveOutputsWithTimelockUnlockConditionBitIndex
	canReceiveOutputsWithExpirationUnlockConditionBitIndex
	canReceiveOutputsWithStorageDepositReturnUnlockConditionBitIndex
	canReceiveAccountOutputsBitIndex
	canReceiveNFTOutputsBitIndex
	canReceiveDelegationOutputsBitIndex
)

// AddressCapabilitiesOptions defines the possible capabilities of an AddressCapabilitiesBitMask.
type AddressCapabilitiesOptions struct {
	canReceiveNativeTokens                                   bool
	canReceiveMana                                           bool
	canReceiveOutputsWithTimelockUnlockCondition             bool
	canReceiveOutputsWithExpirationUnlockCondition           bool
	canReceiveOutputsWithStorageDepositReturnUnlockCondition bool
	canReceiveAccountOutputs                                 bool
	canReceiveNFTOutputs                                     bool
	canReceiveDelegationOutputs                              bool
}

func WithAddressCanReceiveAnything() options.Option[AddressCapabilitiesOptions] {
	return func(o *AddressCapabilitiesOptions) {
		o.canReceiveNativeTokens = true
		o.canReceiveMana = true
		o.canReceiveOutputsWithTimelockUnlockCondition = true
		o.canReceiveOutputsWithExpirationUnlockCondition = true
		o.canReceiveOutputsWithStorageDepositReturnUnlockCondition = true
		o.canReceiveAccountOutputs = true
		o.canReceiveNFTOutputs = true
		o.canReceiveDelegationOutputs = true
	}
}

func WithAddressCanReceiveNativeTokens(canReceiveNativeTokens bool) options.Option[AddressCapabilitiesOptions] {
	return func(o *AddressCapabilitiesOptions) {
		o.canReceiveNativeTokens = canReceiveNativeTokens
	}
}

func WithAddressCanReceiveMana(canReceiveMana bool) options.Option[AddressCapabilitiesOptions] {
	return func(o *AddressCapabilitiesOptions) {
		o.canReceiveMana = canReceiveMana
	}
}

func WithAddressCanReceiveOutputsWithTimelockUnlockCondition(canReceiveOutputsWithTimelockUnlockCondition bool) options.Option[AddressCapabilitiesOptions] {
	return func(o *AddressCapabilitiesOptions) {
		o.canReceiveOutputsWithTimelockUnlockCondition = canReceiveOutputsWithTimelockUnlockCondition
	}
}

func WithAddressCanReceiveOutputsWithExpirationUnlockCondition(canReceiveOutputsWithExpirationUnlockCondition bool) options.Option[AddressCapabilitiesOptions] {
	return func(o *AddressCapabilitiesOptions) {
		o.canReceiveOutputsWithExpirationUnlockCondition = canReceiveOutputsWithExpirationUnlockCondition
	}
}

func WithAddressCanReceiveOutputsWithStorageDepositReturnUnlockCondition(canReceiveOutputsWithStorageDepositReturnUnlockCondition bool) options.Option[AddressCapabilitiesOptions] {
	return func(o *AddressCapabilitiesOptions) {
		o.canReceiveOutputsWithStorageDepositReturnUnlockCondition = canReceiveOutputsWithStorageDepositReturnUnlockCondition
	}
}

func WithAddressCanReceiveAccountOutputs(canReceiveAccountOutputs bool) options.Option[AddressCapabilitiesOptions] {
	return func(o *AddressCapabilitiesOptions) {
		o.canReceiveAccountOutputs = canReceiveAccountOutputs
	}
}

func WithAddressCanReceiveNFTOutputs(canReceiveNFTOutputs bool) options.Option[AddressCapabilitiesOptions] {
	return func(o *AddressCapabilitiesOptions) {
		o.canReceiveNFTOutputs = canReceiveNFTOutputs
	}
}

func WithAddressCanReceiveDelegationOutputs(canReceiveDelegationOutputs bool) options.Option[AddressCapabilitiesOptions] {
	return func(o *AddressCapabilitiesOptions) {
		o.canReceiveDelegationOutputs = canReceiveDelegationOutputs
	}
}

type AddressCapabilitiesBitMask []byte

func AddressCapabilitiesBitMaskFromBytes(bytes []byte) (AddressCapabilitiesBitMask, int, error) {
	var result AddressCapabilitiesBitMask
	consumed, err := CommonSerixAPI().Decode(context.TODO(), bytes, &result)
	return result, consumed, err
}

func AddressCapabilitiesBitMaskWithCapabilities(opts ...options.Option[AddressCapabilitiesOptions]) AddressCapabilitiesBitMask {
	options := options.Apply(new(AddressCapabilitiesOptions), opts)

	bm := AddressCapabilitiesBitMask{}

	if options.canReceiveNativeTokens {
		bm = bm.setBit(canReceiveNativeTokensBitIndex)
	}

	if options.canReceiveMana {
		bm = bm.setBit(canReceiveManaBitIndex)
	}

	if options.canReceiveOutputsWithTimelockUnlockCondition {
		bm = bm.setBit(canReceiveOutputsWithTimelockUnlockConditionBitIndex)
	}

	if options.canReceiveOutputsWithExpirationUnlockCondition {
		bm = bm.setBit(canReceiveOutputsWithExpirationUnlockConditionBitIndex)
	}

	if options.canReceiveOutputsWithStorageDepositReturnUnlockCondition {
		bm = bm.setBit(canReceiveOutputsWithStorageDepositReturnUnlockConditionBitIndex)
	}

	if options.canReceiveAccountOutputs {
		bm = bm.setBit(canReceiveAccountOutputsBitIndex)
	}

	if options.canReceiveNFTOutputs {
		bm = bm.setBit(canReceiveNFTOutputsBitIndex)
	}

	if options.canReceiveDelegationOutputs {
		bm = bm.setBit(canReceiveDelegationOutputsBitIndex)
	}

	return bm
}

func (bm AddressCapabilitiesBitMask) Clone() AddressCapabilitiesBitMask {
	cpy := make(AddressCapabilitiesBitMask, 0, len(bm))
	copy(cpy, bm)
	return cpy
}

func (bm AddressCapabilitiesBitMask) hasBit(bit uint) bool {
	byteIndex := bit / 8
	if uint(len(bm)) <= byteIndex {
		return false
	}
	bitIndex := bit % 8

	return bm[byteIndex]&(1<<bitIndex) > 0
}

func (bm AddressCapabilitiesBitMask) setBit(bit uint) AddressCapabilitiesBitMask {
	newBitmask := bm
	byteIndex := bit / 8
	for uint(len(newBitmask)) <= byteIndex {
		newBitmask = append(newBitmask, 0)
	}
	bitIndex := bit % 8
	newBitmask[byteIndex] |= 1 << bitIndex

	return newBitmask
}

func (bm AddressCapabilitiesBitMask) CannotReceiveNativeTokens() bool {
	return !bm.hasBit(canReceiveNativeTokensBitIndex)
}

func (bm AddressCapabilitiesBitMask) CannotReceiveMana() bool {
	return !bm.hasBit(canReceiveManaBitIndex)
}

func (bm AddressCapabilitiesBitMask) CannotReceiveOutputsWithTimelockUnlockCondition() bool {
	return !bm.hasBit(canReceiveOutputsWithTimelockUnlockConditionBitIndex)
}

func (bm AddressCapabilitiesBitMask) CannotReceiveOutputsWithExpirationUnlockCondition() bool {
	return !bm.hasBit(canReceiveOutputsWithExpirationUnlockConditionBitIndex)
}

func (bm AddressCapabilitiesBitMask) CannotReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return !bm.hasBit(canReceiveOutputsWithStorageDepositReturnUnlockConditionBitIndex)
}

func (bm AddressCapabilitiesBitMask) CannotReceiveAccountOutputs() bool {
	return !bm.hasBit(canReceiveAccountOutputsBitIndex)
}

func (bm AddressCapabilitiesBitMask) CannotReceiveNFTOutputs() bool {
	return !bm.hasBit(canReceiveNFTOutputsBitIndex)
}

func (bm AddressCapabilitiesBitMask) CannotReceiveDelegationOutputs() bool {
	return !bm.hasBit(canReceiveDelegationOutputsBitIndex)
}

func (bm AddressCapabilitiesBitMask) Size() int {
	return serializer.SmallTypeDenotationByteSize + len(bm)
}
