//nolint:gosec
package tpkg

import (
	"math"

	iotago "github.com/iotaledger/iota.go/v4"
)

// Must panics if the given error is not nil.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// ReferenceUnlock returns a reference unlock with the given index.
func ReferenceUnlock(index uint16) *iotago.ReferenceUnlock {
	return &iotago.ReferenceUnlock{Reference: index}
}

// ManaDecayFactors calculates mana decay factors that can be used in the tests.
func ManaDecayFactors(betaPerYear float64, slotsPerEpoch int, slotTimeSeconds int, decayFactorsExponent uint64) []uint32 {
	epochsPerYear := ((365.0 * 24.0 * 60.0 * 60.0) / float64(slotTimeSeconds)) / float64(slotsPerEpoch)
	decayFactors := make([]uint32, int(epochsPerYear))

	betaPerEpochIndex := betaPerYear / epochsPerYear

	for epoch := 1; epoch <= int(epochsPerYear); epoch++ {
		decayFactor := math.Exp(-betaPerEpochIndex*float64(epoch)) * (math.Pow(2, float64(decayFactorsExponent)))
		decayFactors[epoch-1] = uint32(decayFactor)
	}

	return decayFactors
}

// ManaDecayFactorEpochsSum calculates mana decay factor epochs sum parameter that can be used in the tests.
func ManaDecayFactorEpochsSum(betaPerYear float64, slotsPerEpoch int, slotTimeSeconds int, decayFactorEpochsSumExponent uint64) uint32 {
	delta := float64(slotsPerEpoch) * (1.0 / (365.0 * 24.0 * 60.0 * 60.0)) * float64(slotTimeSeconds)

	return uint32((math.Exp(-betaPerYear*delta) / (1 - math.Exp(-betaPerYear*delta)) * (math.Pow(2, float64(decayFactorEpochsSumExponent)))))
}
