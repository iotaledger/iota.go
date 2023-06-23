package iotago_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
)

const (
	betaPerYear float64 = 1 / 3.0
)

var (
	testManaDecayFactors         = getTestManaDecayFactors(betaPerYear, 1<<13, 10, 32)
	testManaDecayFactorEpochsSum = getTestManaDecayFactorEpochsSum(betaPerYear, 1<<13, 10, 20)
)

func getTestManaDecayFactors(betaPerYear float64, slotsPerEpoch int, slotTimeSeconds int, decayFactorsShiftFactor uint64) []uint32 {
	epochsPerYear := ((365.0 * 24.0 * 60.0 * 60.0) / float64(slotTimeSeconds)) / float64(slotsPerEpoch)
	decayFactors := make([]uint32, int(epochsPerYear))

	betaPerDecayIndex := betaPerYear / epochsPerYear

	for epochIndex := 1; epochIndex <= int(epochsPerYear); epochIndex++ {
		decayFactor := math.Exp(-betaPerDecayIndex*float64(epochIndex)) * (math.Pow(2, float64(decayFactorsShiftFactor)))
		decayFactors[epochIndex-1] = uint32(decayFactor)
	}

	return decayFactors
}

func getTestManaDecayFactorEpochsSum(betaPerYear float64, slotsPerEpoch int, slotTimeSeconds int, decayFactorEpochsSumShiftFactor uint64) uint32 {
	delta := float64(slotsPerEpoch) * (1.0 / (365.0 * 24.0 * 60.0 * 60.0)) * float64(slotTimeSeconds)
	return uint32((math.Exp(-betaPerYear*delta) / (1 - math.Exp(-betaPerYear*delta)) * (math.Pow(2, float64(decayFactorEpochsSumShiftFactor)))))
}

func BenchmarkManaDecay_Single(b *testing.B) {
	timeProvider := iotago.NewTimeProvider(0, 10, 1<<13)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, 13, 0, 27, testManaDecayFactors, 32, testManaDecayFactorEpochsSum, 20)

	endIndex := iotago.SlotIndex(300 << 13)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = manaDecayProvider.StoredManaWithDecay(math.MaxUint64, 0, endIndex)
	}
}

func BenchmarkManaDecay_Range(b *testing.B) {
	timeProvider := iotago.NewTimeProvider(0, 10, 1<<13)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, 13, 0, 27, testManaDecayFactors, 32, testManaDecayFactorEpochsSum, 20)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var value iotago.Mana = (1 << 64) - 1
		for decayIndex := 1; decayIndex <= 5*365; decayIndex++ {
			value, _ = manaDecayProvider.StoredManaWithDecay(value, 0, iotago.SlotIndex(decayIndex)<<13)
		}
	}
}

func TestManaDecay_NoFactorsGiven(t *testing.T) {
	timeProvider := iotago.NewTimeProvider(0, 10, 1<<13)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, 13, 0, 27, []uint32{}, 0, 32, 20)

	value, err := manaDecayProvider.StoredManaWithDecay(100, 0, 100<<13)
	require.NoError(t, err)
	require.Equal(t, iotago.Mana(100), value)
}

func TestManaDecay_DecayIndexDiff(t *testing.T) {
	timeProvider := iotago.NewTimeProvider(0, 10, 1<<13)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, 13, 0, 27, testManaDecayFactors, 32, testManaDecayFactorEpochsSum, 20)

	// no decay in the same decay index
	value, err := manaDecayProvider.StoredManaWithDecay(100, 1, 200)
	require.NoError(t, err)
	require.Equal(t, iotago.Mana(100), value)
}

func TestManaDecay_Decay(t *testing.T) {
	timeProvider := iotago.NewTimeProvider(0, 10, 1<<13)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, 13, 0, 27, testManaDecayFactors, 32, testManaDecayFactorEpochsSum, 20)

	{
		// check if mana decay works for multiples of the available decay indexes in the lookup table
		value, err := manaDecayProvider.StoredManaWithDecay(math.MaxUint64, 0, iotago.SlotIndex(3*len(testManaDecayFactors))<<13)
		require.NoError(t, err)
		require.Equal(t, iotago.Mana(6803138682699798504), value)
	}

	{
		// check if mana decay works for exactly the amount of decay indexes in the lookup table
		value, err := manaDecayProvider.StoredManaWithDecay(math.MaxUint64, 0, iotago.SlotIndex(len(testManaDecayFactors))<<13)
		require.NoError(t, err)
		require.Equal(t, iotago.Mana(13228672242897911807), value)
	}

	{
		// check if mana decay works for 0 mana values
		value, err := manaDecayProvider.StoredManaWithDecay(0, 0, 400<<13)
		require.NoError(t, err)
		require.Equal(t, iotago.Mana(0), value)
	}

	{
		// even with the highest possible int64 number, the calculation should not overflow because of the overflow protection
		value, err := manaDecayProvider.StoredManaWithDecay(math.MaxUint64, 0, 400<<13)
		require.NoError(t, err)
		require.Equal(t, iotago.Mana(13046663022640287317), value)
	}
}
