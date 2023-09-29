package iotago

import (
	"context"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	canBurnNativeTokensBitIndex = iota
	canBurnManaBitIndex
	canBurnAccountOutputsBitIndex
	canBurnTokenFoundryOutputsBitIndex
	canBurnNFTOutputsBitIndex
)

// TransactionCapabilitiesOptions defines the possible capabilities of a TransactionCapabilitiesBitMask.
type TransactionCapabilitiesOptions struct {
	canBurnNativeTokens        bool
	canBurnMana                bool
	canBurnAccountOutputs      bool
	canBurnTokenFoundryOutputs bool
	canBurnNFTOutputs          bool
}

func WithTransactionCanDoAnything() options.Option[TransactionCapabilitiesOptions] {
	return func(o *TransactionCapabilitiesOptions) {
		o.canBurnNativeTokens = true
		o.canBurnMana = true
		o.canBurnAccountOutputs = true
		o.canBurnTokenFoundryOutputs = true
		o.canBurnNFTOutputs = true
	}
}

func WithTransactionCanBurnNativeTokens(canBurnNativeTokens bool) options.Option[TransactionCapabilitiesOptions] {
	return func(o *TransactionCapabilitiesOptions) {
		o.canBurnNativeTokens = canBurnNativeTokens
	}
}

func WithTransactionCanBurnMana(canBurnMana bool) options.Option[TransactionCapabilitiesOptions] {
	return func(o *TransactionCapabilitiesOptions) {
		o.canBurnMana = canBurnMana
	}
}

func WithTransactionCanBurnAccountOutputs(canBurnAccountOutputs bool) options.Option[TransactionCapabilitiesOptions] {
	return func(o *TransactionCapabilitiesOptions) {
		o.canBurnAccountOutputs = canBurnAccountOutputs
	}
}

func WithTransactionCanBurnTokenFoundryOutputs(canBurnTokenFoundryOutputs bool) options.Option[TransactionCapabilitiesOptions] {
	return func(o *TransactionCapabilitiesOptions) {
		o.canBurnTokenFoundryOutputs = canBurnTokenFoundryOutputs
	}
}

func WithTransactionCanBurnNFTOutputs(canBurnNFTOutputs bool) options.Option[TransactionCapabilitiesOptions] {
	return func(o *TransactionCapabilitiesOptions) {
		o.canBurnNFTOutputs = canBurnNFTOutputs
	}
}

type TransactionCapabilitiesBitMask []byte

func TransactionCapabilitiesBitMaskFromBytes(bytes []byte) (TransactionCapabilitiesBitMask, int, error) {
	var result TransactionCapabilitiesBitMask
	consumed, err := CommonSerixAPI().Decode(context.TODO(), bytes, &result)
	return result, consumed, err
}

func TransactionCapabilitiesBitMaskWithCapabilities(opts ...options.Option[TransactionCapabilitiesOptions]) TransactionCapabilitiesBitMask {
	options := options.Apply(new(TransactionCapabilitiesOptions), opts)

	bm := []byte{}

	if options.canBurnNativeTokens {
		bm = BitMaskSetBit(bm, canBurnNativeTokensBitIndex)
	}

	if options.canBurnMana {
		bm = BitMaskSetBit(bm, canBurnManaBitIndex)
	}

	if options.canBurnAccountOutputs {
		bm = BitMaskSetBit(bm, canBurnAccountOutputsBitIndex)
	}

	if options.canBurnTokenFoundryOutputs {
		bm = BitMaskSetBit(bm, canBurnTokenFoundryOutputsBitIndex)
	}

	if options.canBurnNFTOutputs {
		bm = BitMaskSetBit(bm, canBurnNFTOutputsBitIndex)
	}

	return bm
}

func (bm TransactionCapabilitiesBitMask) Clone() TransactionCapabilitiesBitMask {
	return lo.CopySlice(bm)
}

func (bm TransactionCapabilitiesBitMask) Size() int {
	return serializer.SmallTypeDenotationByteSize + len(bm)
}

func (bm TransactionCapabilitiesBitMask) CannotBurnNativeTokens() bool {
	return !BitMaskHasBit(bm, canBurnNativeTokensBitIndex)
}

func (bm TransactionCapabilitiesBitMask) CannotBurnMana() bool {
	return !BitMaskHasBit(bm, canBurnManaBitIndex)
}

func (bm TransactionCapabilitiesBitMask) CannotBurnAccountOutputs() bool {
	return !BitMaskHasBit(bm, canBurnAccountOutputsBitIndex)
}

func (bm TransactionCapabilitiesBitMask) CannotBurnTokenFoundryOutputs() bool {
	return !BitMaskHasBit(bm, canBurnTokenFoundryOutputsBitIndex)
}

func (bm TransactionCapabilitiesBitMask) CannotBurnNFTOutputs() bool {
	return !BitMaskHasBit(bm, canBurnNFTOutputsBitIndex)
}
