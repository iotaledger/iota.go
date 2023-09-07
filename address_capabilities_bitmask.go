package iotago

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

func (bm AddressCapabilitiesBitMask) CanReceiveNativeTokens() bool {
	return bm.hasBit(canReceiveNativeTokensBitIndex)
}

func (bm AddressCapabilitiesBitMask) CanReceiveMana() bool {
	return bm.hasBit(canReceiveManaBitIndex)
}

func (bm AddressCapabilitiesBitMask) CanReceiveOutputsWithTimelockUnlockCondition() bool {
	return bm.hasBit(canReceiveOutputsWithTimelockUnlockConditionBitIndex)
}

func (bm AddressCapabilitiesBitMask) CanReceiveOutputsWithExpirationUnlockCondition() bool {
	return bm.hasBit(canReceiveOutputsWithExpirationUnlockConditionBitIndex)
}

func (bm AddressCapabilitiesBitMask) CanReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return bm.hasBit(canReceiveOutputsWithStorageDepositReturnUnlockConditionBitIndex)
}

func (bm AddressCapabilitiesBitMask) CanReceiveAccountOutputs() bool {
	return bm.hasBit(canReceiveAccountOutputsBitIndex)
}

func (bm AddressCapabilitiesBitMask) CanReceiveNFTOutputs() bool {
	return bm.hasBit(canReceiveNFTOutputsBitIndex)
}

func (bm AddressCapabilitiesBitMask) CanReceiveDelegationOutputs() bool {
	return bm.hasBit(canReceiveDelegationOutputsBitIndex)
}
