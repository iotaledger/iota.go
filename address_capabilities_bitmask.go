package iotago

import (
	"context"

	"github.com/iotaledger/hive.go/lo"
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

	bm := []byte{}

	if options.canReceiveNativeTokens {
		bm = BitMaskSetBit(bm, canReceiveNativeTokensBitIndex)
	}

	if options.canReceiveMana {
		bm = BitMaskSetBit(bm, canReceiveManaBitIndex)
	}

	if options.canReceiveOutputsWithTimelockUnlockCondition {
		bm = BitMaskSetBit(bm, canReceiveOutputsWithTimelockUnlockConditionBitIndex)
	}

	if options.canReceiveOutputsWithExpirationUnlockCondition {
		bm = BitMaskSetBit(bm, canReceiveOutputsWithExpirationUnlockConditionBitIndex)
	}

	if options.canReceiveOutputsWithStorageDepositReturnUnlockCondition {
		bm = BitMaskSetBit(bm, canReceiveOutputsWithStorageDepositReturnUnlockConditionBitIndex)
	}

	if options.canReceiveAccountOutputs {
		bm = BitMaskSetBit(bm, canReceiveAccountOutputsBitIndex)
	}

	if options.canReceiveNFTOutputs {
		bm = BitMaskSetBit(bm, canReceiveNFTOutputsBitIndex)
	}

	if options.canReceiveDelegationOutputs {
		bm = BitMaskSetBit(bm, canReceiveDelegationOutputsBitIndex)
	}

	return bm
}

func (bm AddressCapabilitiesBitMask) Clone() AddressCapabilitiesBitMask {
	return lo.CopySlice(bm)
}

func (bm AddressCapabilitiesBitMask) Size() int {
	return serializer.SmallTypeDenotationByteSize + len(bm)
}

func (bm AddressCapabilitiesBitMask) CannotReceiveNativeTokens() bool {
	return !BitMaskHasBit(bm, canReceiveNativeTokensBitIndex)
}

func (bm AddressCapabilitiesBitMask) CannotReceiveMana() bool {
	return !BitMaskHasBit(bm, canReceiveManaBitIndex)
}

func (bm AddressCapabilitiesBitMask) CannotReceiveOutputsWithTimelockUnlockCondition() bool {
	return !BitMaskHasBit(bm, canReceiveOutputsWithTimelockUnlockConditionBitIndex)
}

func (bm AddressCapabilitiesBitMask) CannotReceiveOutputsWithExpirationUnlockCondition() bool {
	return !BitMaskHasBit(bm, canReceiveOutputsWithExpirationUnlockConditionBitIndex)
}

func (bm AddressCapabilitiesBitMask) CannotReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return !BitMaskHasBit(bm, canReceiveOutputsWithStorageDepositReturnUnlockConditionBitIndex)
}

func (bm AddressCapabilitiesBitMask) CannotReceiveAccountOutputs() bool {
	return !BitMaskHasBit(bm, canReceiveAccountOutputsBitIndex)
}

func (bm AddressCapabilitiesBitMask) CannotReceiveNFTOutputs() bool {
	return !BitMaskHasBit(bm, canReceiveNFTOutputsBitIndex)
}

func (bm AddressCapabilitiesBitMask) CannotReceiveDelegationOutputs() bool {
	return !BitMaskHasBit(bm, canReceiveDelegationOutputsBitIndex)
}
