package iotago

import (
	"github.com/iotaledger/hive.go/core/safemath"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
)

// splitUint64 splits a uint64 value into two uint64 that hold the high and the low double-word.
func splitUint64(value uint64) (valueHi uint64, valueLo uint64) {
	return value >> 32, value & 0x00000000FFFFFFFF
}

// mergeUint64 merges two uint64 values that hold the high and the low double-word into one uint64.
func mergeUint64(valueHi uint64, valueLo uint64) (value uint64) {
	return (valueHi << 32) | valueLo
}

// fixedPointMultiplication32Splitted does a fixed point multiplication using two uint64
// containing the high and the low double-word of the value.
// ATTENTION: do not pass factor that use more than 32bits, otherwise this function overflows.
func fixedPointMultiplication32Splitted(valueHi uint64, valueLo uint64, factor uint64, scale uint64) (uint64, uint64) {
	// multiply the integer part of the fixed-point number by the factor
	valueHi *= factor

	// the lower 'scale' bits of the result are extracted and shifted left to form the lower part of the new fraction.
	// the fractional part of the fixed-point number is multiplied by the factor and right-shifted by 'scale' bits.
	// the sum of these two values forms the new lower part (valueLo) of the result.
	valueLo = (valueHi&((1<<scale)-1))<<(32-scale) + (valueLo*factor)>>scale

	// the right-shifted valueHi and the upper 32 bits of valueLo form the new higher part (valueHi) of the result.
	valueHi = (valueHi >> scale) + (valueLo >> 32)

	// the lower 32 bits of valueLo form the new lower part of the result.
	valueLo &= 0x00000000FFFFFFFF

	// return the result as a fixed-point number composed of two 64-bit integers
	return valueHi, valueLo
}

// fixedPointMultiplication32 does a fixed point multiplication.
// ATTENTION: do not pass factor that use more than 32bits, otherwise this function overflows.
func fixedPointMultiplication32(value uint64, factor uint64, scale uint64) uint64 {
	valueHi, valueLo := splitUint64(value)

	return mergeUint64(fixedPointMultiplication32Splitted(valueHi, valueLo, factor, scale))
}

// ManaDecayProvider calculates the mana decay and mana generation
// using fixed point arithmetic and a precomputed lookup table.
type ManaDecayProvider struct {
	timeProvider *TimeProvider

	// slotsPerEpochExponent is the number of slots in an epoch expressed as an exponent of 2.
	// (2**SlotsPerEpochExponent) == slots in an epoch.
	slotsPerEpochExponent uint64

	// bitsCount is the number of bits used to represent Mana.
	bitsCount uint64

	// generationRate is the amount of potential Mana generated by 1 IOTA in 1 slot.
	generationRate uint64 // the generation rate needs to be scaled by 2^-generationRateExponent

	// generationRateExponent is the scaling of generationRate expressed as an exponent of 2.
	generationRateExponent uint64

	// decayFactors is a lookup table of epoch index diff to mana decay factor (slice index 0 = 1 epoch).
	decayFactors []uint64 // the factors need to be scaled by 2^-decayFactorsExponent

	// decayFactorsLength is the length of the decayFactors lookup table.
	decayFactorsLength uint64

	// decayFactorsExponent is the scaling of decayFactors expressed as an exponent of 2.
	decayFactorsExponent uint64

	// decayFactorEpochsSum is an integer approximation of the sum of decay over epochs.
	decayFactorEpochsSum uint64 // the factor needs to be scaled by 2^-decayFactorEpochsSumExponent

	// decayFactorEpochsSumExponent is the scaling of decayFactorEpochsSum expressed as an exponent of 2.
	decayFactorEpochsSumExponent uint64
}

func NewManaDecayProvider(
	timeProvider *TimeProvider,
	slotsPerEpochExponent uint8,
	manaParameters *ManaParameters,
) *ManaDecayProvider {
	return &ManaDecayProvider{
		timeProvider:                 timeProvider,
		slotsPerEpochExponent:        uint64(slotsPerEpochExponent),
		bitsCount:                    uint64(manaParameters.BitsCount),
		generationRate:               uint64(manaParameters.GenerationRate),
		generationRateExponent:       uint64(manaParameters.GenerationRateExponent),
		decayFactors:                 lo.Map(manaParameters.DecayFactors, func(factor uint32) uint64 { return uint64(factor) }),
		decayFactorsLength:           uint64(len(manaParameters.DecayFactors)),
		decayFactorsExponent:         uint64(manaParameters.DecayFactorsExponent),
		decayFactorEpochsSum:         uint64(manaParameters.DecayFactorEpochsSum),
		decayFactorEpochsSumExponent: uint64(manaParameters.DecayFactorEpochsSumExponent),
	}
}

// decay performs mana decay without mana generation.
func (p *ManaDecayProvider) decay(value Mana, epochDiff EpochIndex) Mana {
	if value == 0 || epochDiff == 0 || p.decayFactorsLength == 0 {
		// no need to decay if the epoch index didn't change or no decay factors were given
		return value
	}

	// split the value into two uint64 variables to prevent overflows
	valueHi, valueLo := splitUint64(uint64(value))

	// we keep applying the decay as long as epoch index diffs are left
	remainingEpochDiff := epochDiff
	for remainingEpochDiff > 0 {
		// we can't decay more than the available epoch index diffs
		// in the lookup table in this iteration
		diffsToDecay := remainingEpochDiff
		if diffsToDecay > EpochIndex(p.decayFactorsLength) {
			diffsToDecay = EpochIndex(p.decayFactorsLength)
		}
		remainingEpochDiff -= diffsToDecay

		// slice index 0 equals epoch index diff 1
		decayFactor := p.decayFactors[diffsToDecay-1]

		// apply the decay and scale the resulting value (fixed-point arithmetics)
		valueHi, valueLo = fixedPointMultiplication32Splitted(valueHi, valueLo, decayFactor, p.decayFactorsExponent)
	}

	// combine both uint64 variables to get the actual value
	return Mana(mergeUint64(valueHi, valueLo))
}

// generateMana calculates the generated mana.
func (p *ManaDecayProvider) generateMana(value BaseToken, slotDiff SlotIndex) Mana {
	if slotDiff == 0 || p.generationRate == 0 {
		return 0
	}

	return Mana(fixedPointMultiplication32(uint64(value), uint64(slotDiff)*p.generationRate, p.generationRateExponent))
}

// ManaWithDecay applies the decay to the given mana.
func (p *ManaDecayProvider) ManaWithDecay(storedMana Mana, creationSlot SlotIndex, targetSlot SlotIndex) (Mana, error) {
	creationEpoch := p.timeProvider.EpochFromSlot(creationSlot)
	targetEpoch := p.timeProvider.EpochFromSlot(targetSlot)

	if creationEpoch > targetEpoch {
		return 0, ierrors.Wrapf(ErrWrongEpochIndex, "the created epoch index was bigger than the target epoch index: %d > %d", creationEpoch, targetEpoch)
	}

	return p.decay(storedMana, targetEpoch-creationEpoch), nil
}

// ManaGenerationWithDecay calculates the generated mana and applies the decay to the result.
func (p *ManaDecayProvider) ManaGenerationWithDecay(amount BaseToken, creationSlot SlotIndex, targetSlot SlotIndex) (Mana, error) {
	creationEpoch := p.timeProvider.EpochFromSlot(creationSlot)
	targetEpoch := p.timeProvider.EpochFromSlot(targetSlot)

	if creationEpoch > targetEpoch {
		return 0, ierrors.Wrapf(ErrWrongEpochIndex, "the created epoch index was bigger than the target epoch index: %d > %d", creationEpoch, targetEpoch)
	}

	epochDiff := targetEpoch - creationEpoch

	//nolint:exhaustive // false-positive, we have default case
	switch epochDiff {
	case 0:
		return p.generateMana(amount, targetSlot-creationSlot), nil

	case 1:
		manaDecayed := p.decay(p.generateMana(amount, p.timeProvider.SlotsBeforeNextEpoch(creationSlot)), 1)
		manaGenerated := p.generateMana(amount, p.timeProvider.SlotsSinceEpochStart(targetSlot))
		return safemath.SafeAdd(manaDecayed, manaGenerated)

	default:
		c := Mana(fixedPointMultiplication32(uint64(amount), p.decayFactorEpochsSum, p.decayFactorEpochsSumExponent+p.generationRateExponent-p.slotsPerEpochExponent))

		//nolint:golint,revive,nosnakecase,stylecheck // taken from the formula, lets keep it that way
		potentialMana_n := p.decay(p.generateMana(amount, p.timeProvider.SlotsBeforeNextEpoch(creationSlot)), epochDiff)

		//nolint:golint,revive,nosnakecase,stylecheck // taken from the formula, lets keep it that way
		potentialMana_n_1 := p.decay(c, epochDiff-1)

		//nolint:golint,revive,nosnakecase,stylecheck // taken from the formula, lets keep it that way
		potentialMana_0, err := safemath.SafeAdd(c, p.generateMana(amount, p.timeProvider.SlotsSinceEpochStart(targetSlot)))
		if err != nil {
			return 0, err
		}

		// result = potentialMana_0 - potentialMana_n_1 + potentialMana_n
		//nolint:golint,revive,nosnakecase,stylecheck // taken from the formula, lets keep it that way
		result, err := safemath.SafeSub(potentialMana_0, potentialMana_n_1)
		if err != nil {
			return 0, err
		}

		//nolint:golint,revive,nosnakecase,stylecheck // taken from the formula, lets keep it that way
		return safemath.SafeAdd(result, potentialMana_n)
	}
}

// RewardsWithDecay applies the decay to the given stored mana.
func (p *ManaDecayProvider) RewardsWithDecay(rewards Mana, rewardEpoch EpochIndex, claimedEpoch EpochIndex) (Mana, error) {
	if rewardEpoch > claimedEpoch {
		return 0, ierrors.Wrapf(ErrWrongEpochIndex, "the reward epoch index was bigger than the claiming epoch index: %d > %d", rewardEpoch, claimedEpoch)
	}

	return p.decay(rewards, claimedEpoch-rewardEpoch), nil
}
