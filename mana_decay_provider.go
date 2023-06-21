package iotago

import (
	"fmt"

	"github.com/iotaledger/hive.go/lo"
)

const (
	// ManaDecayScaleFactor is the amount of bits that are used
	// to scale the decay factors for fixed-point arithmetics.
	ManaDecayScaleFactor = 12 // 2^ManaDecayScaleFactor

	// ManaValueScaleThreshold is the maximum mana value that can be decayed safely without overflows.
	// 63 bits are used to support negative values for mana.
	ManaValueScaleThreshold = uint64((1 << (63 - ManaDecayScaleFactor)) - 1)
)

type ManaDecayProvider struct {
	timeProvider            *TimeProvider
	generationRate          uint8
	decayFactors            []uint64 // the decay factors need to be scaled by 2^decayFactorsScaleFactor
	decayFactorsLength      uint64
	decayFactorsScaleFactor uint64
}

func NewManaDecayProvider(timeProvider *TimeProvider, generationRate uint8, decayFactors []uint32, decayFactorsScaleFactor uint8) *ManaDecayProvider {
	return &ManaDecayProvider{
		timeProvider:            timeProvider,
		generationRate:          generationRate,
		decayFactors:            lo.Map(decayFactors, func(factor uint32) uint64 { return uint64(factor) }),
		decayFactorsLength:      uint64(len(decayFactors)),
		decayFactorsScaleFactor: uint64(decayFactorsScaleFactor),
	}
}

func (p *ManaDecayProvider) decay(value uint64, epochDiff EpochIndex) uint64 {
	if value == 0 || epochDiff == 0 || p.decayFactorsLength == 0 {
		// no need to decay if the epoch index didn't change or no decay factors were given
		return value
	}

	// scale the mana value to avoid overflowing of the decay calculation
	bitsDecreased := 0
	for value > ManaValueScaleThreshold {
		value = value >> 1
		bitsDecreased++
	}

	// we keep applying the decay as long as epoch index diffs are left
	remainingEpochIndexDiff := epochDiff
	for remainingEpochIndexDiff > 0 {
		// we can't decay more than the available epoch index diffs
		// in the lookup table in this iteration
		diffsToDecay := remainingEpochIndexDiff
		if diffsToDecay > EpochIndex(p.decayFactorsLength) {
			diffsToDecay = EpochIndex(p.decayFactorsLength)
		}
		remainingEpochIndexDiff -= diffsToDecay

		// slice index 0 equals epoch index diff 1
		decayFactor := p.decayFactors[diffsToDecay-1]

		// apply the decay and scale the resulting value (fixed-point arithmetics)
		value = (value * uint64(decayFactor)) >> ManaDecayScaleFactor
	}

	// scale the mana value back to the correct size
	for i := 0; i < bitsDecreased; i++ {
		value = value << 1
	}

	return value
}

func (p *ManaDecayProvider) StoredManaWithDecay(storedMana uint64, slotIndexLast SlotIndex, slotIndexNew SlotIndex) uint64 {
	epochIndexLast := p.timeProvider.EpochsFromSlot(slotIndexLast)
	epochIndexNew := p.timeProvider.EpochsFromSlot(slotIndexNew)

	if epochIndexLast > epochIndexNew {
		panic(fmt.Sprintf("the last epoch index was bigger than the new epoch index: %d > %d", epochIndexLast, epochIndexNew))
	}

	return p.decay(storedMana, epochIndexNew-epochIndexLast)
}

func (p *ManaDecayProvider) PotentialManaWithDecay(deposit uint64, slotIndexLast SlotIndex, slotIndexNew SlotIndex) uint64 {
	epochIndexLast := p.timeProvider.EpochsFromSlot(slotIndexLast)
	epochIndexNew := p.timeProvider.EpochsFromSlot(slotIndexNew)

	if epochIndexLast > epochIndexNew {
		panic(fmt.Sprintf("the last epoch index was bigger than the new epoch index: %d > %d", epochIndexLast, epochIndexNew))
	}

	// TODO: we need to take mana generation into consideration here
	return p.decay(deposit, epochIndexNew-epochIndexLast)
}

func (p *ManaDecayProvider) RewardsWithDecay(rewards uint64, epochIndexStart EpochIndex, epochIndexEnd EpochIndex) uint64 {
	if epochIndexStart > epochIndexEnd {
		panic(fmt.Sprintf("the start epoch index was bigger than the end epoch index: %d > %d", epochIndexStart, epochIndexEnd))
	}

	return p.decay(rewards, epochIndexEnd-epochIndexStart)
}
