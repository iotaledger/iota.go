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
	canDestroyAccountOutputsBitIndex
	canDestroyFoundryOutputsBitIndex
	canDestroyNFTOutputsBitIndex
)

// TransactionCapabilitiesOptions defines the possible capabilities of a TransactionCapabilitiesBitMask.
type TransactionCapabilitiesOptions struct {
	canBurnNativeTokens      bool
	canBurnMana              bool
	canDestroyAccountOutputs bool
	canDestroyFoundryOutputs bool
	canDestroyNFTOutputs     bool
}

func WithTransactionCanDoAnything() options.Option[TransactionCapabilitiesOptions] {
	return func(o *TransactionCapabilitiesOptions) {
		o.canBurnNativeTokens = true
		o.canBurnMana = true
		o.canDestroyAccountOutputs = true
		o.canDestroyFoundryOutputs = true
		o.canDestroyNFTOutputs = true
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

func WithTransactionCanDestroyAccountOutputs(canDestroyAccountOutputs bool) options.Option[TransactionCapabilitiesOptions] {
	return func(o *TransactionCapabilitiesOptions) {
		o.canDestroyAccountOutputs = canDestroyAccountOutputs
	}
}

func WithTransactionCanDestroyFoundryOutputs(canDestroyFoundryOutputs bool) options.Option[TransactionCapabilitiesOptions] {
	return func(o *TransactionCapabilitiesOptions) {
		o.canDestroyFoundryOutputs = canDestroyFoundryOutputs
	}
}

func WithTransactionCanDestroyNFTOutputs(canDestroyNFTOutputs bool) options.Option[TransactionCapabilitiesOptions] {
	return func(o *TransactionCapabilitiesOptions) {
		o.canDestroyNFTOutputs = canDestroyNFTOutputs
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

	if options.canDestroyAccountOutputs {
		bm = BitMaskSetBit(bm, canDestroyAccountOutputsBitIndex)
	}

	if options.canDestroyFoundryOutputs {
		bm = BitMaskSetBit(bm, canDestroyFoundryOutputsBitIndex)
	}

	if options.canDestroyNFTOutputs {
		bm = BitMaskSetBit(bm, canDestroyNFTOutputsBitIndex)
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

func (bm TransactionCapabilitiesBitMask) CannotDestroyAccountOutputs() bool {
	return !BitMaskHasBit(bm, canDestroyAccountOutputsBitIndex)
}

func (bm TransactionCapabilitiesBitMask) CannotDestroyFoundryOutputs() bool {
	return !BitMaskHasBit(bm, canDestroyFoundryOutputsBitIndex)
}

func (bm TransactionCapabilitiesBitMask) CannotDestroyNFTOutputs() bool {
	return !BitMaskHasBit(bm, canDestroyNFTOutputsBitIndex)
}
