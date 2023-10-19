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
func fixedPointMultiplication32Splitted(valueHi uint64, valueLo uint64, factor uint64, scale uint64) (uint64, uint64, error) {
	if scale > 32 {
		return 0, 0, ierrors.Errorf("fixed point multiplication with a scaling factor >32 (%d) not allowed", scale)
	}

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
	return valueHi, valueLo, nil
}

// fixedPointMultiplication32 does a fixed point multiplication.
func fixedPointMultiplication32(value uint64, factor uint64, scale uint64) (uint64, error) {
	var remainingScale uint64
	if scale > 32 {
		remainingScale = scale - 32
		scale = 32
	}
	valueHi, valueLo := splitUint64(value)

	resultHi, resultLo, err := fixedPointMultiplication32Splitted(valueHi, valueLo, factor, scale)
	if err != nil {
		return 0, err
	}

	return mergeUint64(resultHi, resultLo) >> remainingScale, nil
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

	// decayFactors is a lookup table of epoch diff to mana decay factor (slice index 0 = 1 epoch).
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
func (p *ManaDecayProvider) decay(value Mana, epochDiff EpochIndex) (Mana, error) {
	if value == 0 || epochDiff == 0 || p.decayFactorsLength == 0 {
		// no need to decay if the epoch didn't change or no decay factors were given
		return value, nil
	}

	// split the value into two uint64 variables to prevent overflows
	valueHi, valueLo := splitUint64(uint64(value))

	// we keep applying the decay as long as epoch diffs are left
	remainingEpochDiff := epochDiff
	for remainingEpochDiff > 0 {
		// we can't decay more than the available epoch diffs
		// in the lookup table in this iteration
		diffsToDecay := remainingEpochDiff
		if diffsToDecay > EpochIndex(p.decayFactorsLength) {
			diffsToDecay = EpochIndex(p.decayFactorsLength)
		}
		remainingEpochDiff -= diffsToDecay

		// slice index 0 equals epoch diff 1
		decayFactor := p.decayFactors[diffsToDecay-1]

		// apply the decay and scale the resulting value (fixed-point arithmetics)
		var err error
		valueHi, valueLo, err = fixedPointMultiplication32Splitted(valueHi, valueLo, decayFactor, p.decayFactorsExponent)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to calculate mana decay")
		}
	}

	// combine both uint64 variables to get the actual value
	return Mana(mergeUint64(valueHi, valueLo)), nil
}

// generateMana calculates the generated mana.
func (p *ManaDecayProvider) generateMana(value BaseToken, slotDiff SlotIndex) (Mana, error) {
	if slotDiff == 0 || p.generationRate == 0 {
		return 0, nil
	}

	result, err := fixedPointMultiplication32(uint64(value), uint64(slotDiff)*p.generationRate, p.generationRateExponent)
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to calculate mana generation")
	}

	return Mana(result), nil
}

// ManaWithDecay applies the decay to the given mana.
func (p *ManaDecayProvider) ManaWithDecay(storedMana Mana, creationSlot SlotIndex, targetSlot SlotIndex) (Mana, error) {
	creationEpoch := p.timeProvider.EpochFromSlot(creationSlot)
	targetEpoch := p.timeProvider.EpochFromSlot(targetSlot)

	if creationEpoch > targetEpoch {
		return 0, ierrors.Wrapf(ErrWrongEpochIndex, "the created epoch was bigger than the target epoch: %d > %d", creationEpoch, targetEpoch)
	}

	return p.decay(storedMana, targetEpoch-creationEpoch)
}

// ManaGenerationWithDecay calculates the generated mana and applies the decay to the result.
func (p *ManaDecayProvider) ManaGenerationWithDecay(amount BaseToken, creationSlot SlotIndex, targetSlot SlotIndex) (Mana, error) {
	if creationSlot > targetSlot {
		return 0, ierrors.Wrapf(ErrWrongEpochIndex, "the created slot was bigger than the target slot: %d > %d", creationSlot, targetSlot)
	}

	creationEpoch := p.timeProvider.EpochFromSlot(creationSlot)
	targetEpoch := p.timeProvider.EpochFromSlot(targetSlot)

	epochDiff := targetEpoch - creationEpoch

	//nolint:exhaustive // false-positive, we have a default case
	switch epochDiff {
	case 0:
		result, err := p.generateMana(amount, targetSlot-creationSlot)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to calculate generated mana")
		}

		return result, nil

	case 1:
		manaGeneratedFirstEpoch, err := p.generateMana(amount, p.timeProvider.SlotsBeforeNextEpoch(creationSlot))
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to calculate generated mana in the first epoch")
		}

		manaDecayedFirstEpoch, err := p.decay(manaGeneratedFirstEpoch, 1)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to decay generated mana in the first epoch")
		}

		manaGeneratedSecondEpoch, err := p.generateMana(amount, p.timeProvider.SlotsSinceEpochStart(targetSlot))
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to calculate generated mana in the second epoch")
		}

		result, err := safemath.SafeAdd(manaDecayedFirstEpoch, manaGeneratedSecondEpoch)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to calculate sum of generated mana")
		}

		return result, nil

	default:
		aux, err := fixedPointMultiplication32(uint64(amount), p.decayFactorEpochsSum*p.generationRate, p.decayFactorEpochsSumExponent+p.generationRateExponent-p.slotsPerEpochExponent)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to calculate auxiliary value")
		}
		c := Mana(aux)

		manaGeneratedFirstEpoch, err := p.generateMana(amount, p.timeProvider.SlotsBeforeNextEpoch(creationSlot))
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to calculate generated mana in the first epoch")
		}

		manaDecayedFirstEpoch, err := p.decay(manaGeneratedFirstEpoch, epochDiff)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to decay generated mana in the first epoch")
		}

		manaDecayedIntermediateEpochs, err := p.decay(c, epochDiff-1)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to decay generated mana in the intermediate epochs")
		}

		manaGeneratedLastEpoch, err := p.generateMana(amount, p.timeProvider.SlotsSinceEpochStart(targetSlot))
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to calculate generated mana in the last epoch")
		}

		result, err := safemath.SafeAdd(c, manaGeneratedLastEpoch)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to calculate sum of generated mana of the last epoch")
		}

		result, err = safemath.SafeSub(result, c>>p.decayFactorsExponent)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to calculate subtraction of generated mana from the rounding term")
		}

		result, err = safemath.SafeSub(result, manaDecayedIntermediateEpochs)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to calculate subtraction of generated mana of intermediate epochs")
		}

		result, err = safemath.SafeAdd(result, manaDecayedFirstEpoch)
		if err != nil {
			return 0, ierrors.Wrap(err, "failed to calculate sum of generated mana of the first epoch")
		}

		return result, nil
	}
}

// RewardsWithDecay applies the decay to the given stored mana.
func (p *ManaDecayProvider) RewardsWithDecay(rewards Mana, rewardEpoch EpochIndex, claimedEpoch EpochIndex) (Mana, error) {
	if rewardEpoch > claimedEpoch {
		return 0, ierrors.Wrapf(ErrWrongEpochIndex, "the reward epoch was bigger than the claiming epoch: %d > %d", rewardEpoch, claimedEpoch)
	}

	return p.decay(rewards, claimedEpoch-rewardEpoch)
}
