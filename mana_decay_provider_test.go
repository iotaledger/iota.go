package iotago_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
)

const (
	manaValue53Bit = 9007199254740991 // 0x1FFFFFFFFFFFFF, 53 bits set to 1 (maximum mana value)
)

var (
	testManaDecayFactors         = getTestManaDecayFactors(365, 32)
	testManaDecayFactorEpochsSum = getTestManaDecayFactorEpochsSum(32)
)

func getTestManaDecayFactors(decayIndexes int, decayFactorsScaleFactor uint64) []uint32 {
	decayFactors := make([]uint32, decayIndexes)

	betaPerYear := 1 / 3.0
	betaPerDecayIndex := betaPerYear / 365.0

	for decayIndex := 1; decayIndex <= decayIndexes; decayIndex++ {
		decayFactor := math.Exp(-betaPerDecayIndex*float64(decayIndex)) * (math.Pow(2, float64(decayFactorsScaleFactor)))
		decayFactors[decayIndex-1] = uint32(decayFactor)
	}

	return decayFactors
}

func getTestManaDecayFactorEpochsSum(decayFactorsScaleFactor uint64) uint32 {
	betaPerYear := 1 / 3.0
	delta := float64(1<<13) * (1.0 / (365.0 * 24.0 * 60.0 * 60.0)) * 10.0
	return uint32((math.Exp(-betaPerYear*delta) / (1 - math.Exp(-betaPerYear*delta)) * (math.Pow(2, float64(decayFactorsScaleFactor)))))
}

func BenchmarkManaDecay_Single(b *testing.B) {
	timeProvider := iotago.NewTimeProvider(0, 10, 1<<13)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, 13, 0, 27, testManaDecayFactors, 32, testManaDecayFactorEpochsSum, 20)

	endIndex := iotago.SlotIndex(300 << 13)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = manaDecayProvider.StoredManaWithDecay(manaValue53Bit, 0, endIndex)
	}
}

func BenchmarkManaDecay_Range(b *testing.B) {
	timeProvider := iotago.NewTimeProvider(0, 10, 1<<13)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, 13, 0, 27, testManaDecayFactors, 32, testManaDecayFactorEpochsSum, 20)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var value iotago.Mana = manaValue53Bit
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
	require.Equal(t, uint64(100), value)
}

func TestManaDecay_DecayIndexDiff(t *testing.T) {
	timeProvider := iotago.NewTimeProvider(0, 10, 1<<13)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, 13, 0, 27, testManaDecayFactors, 32, testManaDecayFactorEpochsSum, 20)

	// no decay in the same decay index
	value, err := manaDecayProvider.StoredManaWithDecay(100, 1, 200)
	require.NoError(t, err)
	require.Equal(t, uint64(100), value)

	require.Panics(t, func() {
		manaDecayProvider.StoredManaWithDecay(100, 2<<13, 1<<13)
	})
}

func TestManaDecay_Decay(t *testing.T) {
	timeProvider := iotago.NewTimeProvider(0, 10, 1<<13)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, 13, 0, 27, testManaDecayFactors, 32, testManaDecayFactorEpochsSum, 20)

	{
		// check if mana decay works for multiples of the available decay indexes in the lookup table
		value, err := manaDecayProvider.StoredManaWithDecay(manaValue53Bit, 0, iotago.SlotIndex(3*len(testManaDecayFactors))<<13)
		require.NoError(t, err)
		require.Equal(t, uint64(3310474560012284), value)
	}

	{
		// check if mana decay works for exactly the  amount of decay indexes in the lookup table
		value, err := manaDecayProvider.StoredManaWithDecay(manaValue53Bit, 0, iotago.SlotIndex(len(testManaDecayFactors))<<13)
		require.NoError(t, err)
		require.Equal(t, uint64(6451934231789564), value)
	}

	{
		// check if mana decay works for 0 mana values
		value, err := manaDecayProvider.StoredManaWithDecay(0, 0, 400<<13)
		require.NoError(t, err)
		require.Equal(t, uint64(0), value)
	}

	{
		// even with the highest possible int64 number, the calculation should not overflow because of the overflow protection
		value, err := manaDecayProvider.StoredManaWithDecay(math.MaxInt64, 0, 400<<13)
		require.NoError(t, err)
		require.Equal(t, uint64(6398705774377299968), value)
	}
}
