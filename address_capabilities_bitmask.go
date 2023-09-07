package iotago

import "github.com/iotaledger/hive.go/serializer/v2"

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

type AddressCapabilitiesBitMask []byte

func (bm AddressCapabilitiesBitMask) hasBit(bit int) bool {
	byteIndex := bit / 8
	if len(bm) <= byteIndex {
		return false
	}
	bitIndex := bit % 8
	return bm[byteIndex]&(1<<bitIndex) > 0
}

func (bm AddressCapabilitiesBitMask) setBit(bit int) AddressCapabilitiesBitMask {
	newBitmask := bm
	byteIndex := bit / 8
	for len(newBitmask) <= byteIndex {
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
