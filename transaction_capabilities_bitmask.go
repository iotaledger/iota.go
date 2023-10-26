package iotago

import (
	"context"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrTxCapabilitiesNativeTokenBurningNotAllowed gets returned when native tokens are burned in a transaction but it was not allowed by the capabilities.
	ErrTxCapabilitiesNativeTokenBurningNotAllowed = ierrors.New("native token burning is not allowed by the transaction capabilities")
	// ErrTxCapabilitiesManaBurningNotAllowed gets returned when mana is burned in a transaction but it was not allowed by the capabilities.
	ErrTxCapabilitiesManaBurningNotAllowed = ierrors.New("mana burning is not allowed by the transaction capabilities")
	// ErrTxCapabilitiesAccountDestructionNotAllowed gets returned when an account is destroyed in a transaction but it was not allowed by the capabilities.
	ErrTxCapabilitiesAccountDestructionNotAllowed = ierrors.New("account destruction is not allowed by the transaction capabilities")
	// ErrTxCapabilitiesAnchorDestructionNotAllowed gets returned when an anchor is destroyed in a transaction but it was not allowed by the capabilities.
	ErrTxCapabilitiesAnchorDestructionNotAllowed = ierrors.New("anchor destruction is not allowed by the transaction capabilities")
	// ErrTxCapabilitiesFoundryDestructionNotAllowed gets returned when a foundry is destroyed in a transaction but it was not allowed by the capabilities.
	ErrTxCapabilitiesFoundryDestructionNotAllowed = ierrors.New("foundry destruction is not allowed by the transaction capabilities")
	// ErrTxCapabilitiesNFTDestructionNotAllowed gets returned when a NFT is destroyed in a transaction but it was not allowed by the capabilities.
	ErrTxCapabilitiesNFTDestructionNotAllowed = ierrors.New("NFT destruction is not allowed by the transaction capabilities")
)

const (
	canBurnNativeTokensBitIndex = iota
	canBurnManaBitIndex
	canDestroyAccountOutputsBitIndex
	canDestroyAnchorOutputsBitIndex
	canDestroyFoundryOutputsBitIndex
	canDestroyNFTOutputsBitIndex
)

// TransactionCapabilitiesOptions defines the possible capabilities of a TransactionCapabilitiesBitMask.
type TransactionCapabilitiesOptions struct {
	canBurnNativeTokens      bool
	canBurnMana              bool
	canDestroyAccountOutputs bool
	canDestroyAnchorOutputs  bool
	canDestroyFoundryOutputs bool
	canDestroyNFTOutputs     bool
}

func WithTransactionCanDoAnything() options.Option[TransactionCapabilitiesOptions] {
	return func(o *TransactionCapabilitiesOptions) {
		o.canBurnNativeTokens = true
		o.canBurnMana = true
		o.canDestroyAccountOutputs = true
		o.canDestroyAnchorOutputs = true
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

func WithTransactionCanDestroyAnchorOutputs(canDestroyAnchorOutputs bool) options.Option[TransactionCapabilitiesOptions] {
	return func(o *TransactionCapabilitiesOptions) {
		o.canDestroyAnchorOutputs = canDestroyAnchorOutputs
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

	if options.canDestroyAnchorOutputs {
		bm = BitMaskSetBit(bm, canDestroyAnchorOutputsBitIndex)
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

func (bm TransactionCapabilitiesBitMask) CannotDestroyAnchorOutputs() bool {
	return !BitMaskHasBit(bm, canDestroyAnchorOutputsBitIndex)
}

func (bm TransactionCapabilitiesBitMask) CannotDestroyFoundryOutputs() bool {
	return !BitMaskHasBit(bm, canDestroyFoundryOutputsBitIndex)
}

func (bm TransactionCapabilitiesBitMask) CannotDestroyNFTOutputs() bool {
	return !BitMaskHasBit(bm, canDestroyNFTOutputsBitIndex)
}
