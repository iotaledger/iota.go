package iotago

import "github.com/iotaledger/hive.go/lo"

// Mana Structure defines the parameters used in mana calculations.
type ManaStructure struct {
	// ManaBitsCount is the number of bits used to represent Mana.
	ManaBitsCount uint8 `serix:"0,mapKey=manaBitsCount"`
	// ManaGenerationRate is the amount of potential Mana generated by 1 IOTA in 1 slot.
	ManaGenerationRate uint8 `serix:"1,mapKey=manaGenerationRate"`
	// ManaGenerationRateExponent is the scaling of ManaGenerationRate expressed as an exponent of 2.
	ManaGenerationRateExponent uint8 `serix:"2,mapKey=manaGenerationRateExponent"`
	// ManaDecayFactors is a lookup table of epoch index diff to mana decay factor (slice index 0 = 1 epoch).
	ManaDecayFactors []uint32 `serix:"3,lengthPrefixType=uint16,mapKey=manaDecayFactors"`
	// ManaDecayFactorsExponent is the scaling of ManaDecayFactors expressed as an exponent of 2.
	ManaDecayFactorsExponent uint8 `serix:"4,mapKey=manaDecayFactorsExponent"`
	// ManaDecayFactorEpochsSum is an integer approximation of the sum of decay over epochs.
	ManaDecayFactorEpochsSum uint32 `serix:"5,mapKey=manaDecayFactorEpochsSum"`
	// ManaDecayFactorEpochsSumExponent is the scaling of ManaDecayFactorEpochsSum expressed as an exponent of 2.
	ManaDecayFactorEpochsSumExponent uint8 `serix:"6,mapKey=manaDecayFactorEpochsSumExponent"`
}

func (m ManaStructure) Equals(other ManaStructure) bool {
	return m.ManaBitsCount == other.ManaBitsCount &&
		m.ManaGenerationRate == other.ManaGenerationRate &&
		m.ManaGenerationRateExponent == other.ManaGenerationRateExponent &&
		lo.Equal(m.ManaDecayFactors, other.ManaDecayFactors) &&
		m.ManaDecayFactorsExponent == other.ManaDecayFactorsExponent &&
		m.ManaDecayFactorEpochsSum == other.ManaDecayFactorEpochsSum &&
		m.ManaDecayFactorEpochsSumExponent == other.ManaDecayFactorEpochsSumExponent
}