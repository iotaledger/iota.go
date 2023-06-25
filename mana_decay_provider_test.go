package iotago_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

const (
	betaPerYear                     float64 = 1 / 3.0
	slotsPerEpochShiftFactor                = 13
	slotDurationSeconds                     = 10
	generationRate                          = 1
	generationRateShiftFactor               = 27
	decayFactorsShiftFactor                 = 32
	decayFactorEpochsSumShiftFactor         = 20
)

var (
	testManaDecayFactors         = tpkg.ManaDecayFactors(betaPerYear, 1<<slotsPerEpochShiftFactor, slotDurationSeconds, decayFactorsShiftFactor)
	testManaDecayFactorEpochsSum = tpkg.ManaDecayFactorEpochsSum(betaPerYear, 1<<slotsPerEpochShiftFactor, slotDurationSeconds, decayFactorEpochsSumShiftFactor)
)

func BenchmarkManaDecay_Single(b *testing.B) {
	timeProvider := iotago.NewTimeProvider(0, slotDurationSeconds, slotsPerEpochShiftFactor)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, slotsPerEpochShiftFactor, generationRate, decayFactorEpochsSumShiftFactor, testManaDecayFactors, decayFactorsShiftFactor, testManaDecayFactorEpochsSum, decayFactorEpochsSumShiftFactor)

	endIndex := iotago.SlotIndex(300 << slotsPerEpochShiftFactor)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = manaDecayProvider.StoredManaWithDecay(math.MaxUint64, 0, endIndex)
	}
}

func BenchmarkManaDecay_Range(b *testing.B) {
	timeProvider := iotago.NewTimeProvider(0, slotDurationSeconds, slotsPerEpochShiftFactor)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, slotsPerEpochShiftFactor, generationRate, decayFactorEpochsSumShiftFactor, testManaDecayFactors, decayFactorsShiftFactor, testManaDecayFactorEpochsSum, decayFactorEpochsSumShiftFactor)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var value iotago.Mana = (1 << 64) - 1
		for decayIndex := 1; decayIndex <= 5*365; decayIndex++ {
			value, _ = manaDecayProvider.StoredManaWithDecay(value, 0, iotago.SlotIndex(decayIndex)<<13)
		}
	}
}

func TestManaDecay_NoFactorsGiven(t *testing.T) {
	timeProvider := iotago.NewTimeProvider(0, slotDurationSeconds, slotsPerEpochShiftFactor)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, slotsPerEpochShiftFactor, generationRate, decayFactorEpochsSumShiftFactor, []uint32{}, decayFactorsShiftFactor, testManaDecayFactorEpochsSum, decayFactorEpochsSumShiftFactor)

	value, err := manaDecayProvider.StoredManaWithDecay(100, 0, 100<<13)
	require.NoError(t, err)
	require.Equal(t, iotago.Mana(100), value)
}

func TestManaDecay_DecayIndexDiff(t *testing.T) {
	timeProvider := iotago.NewTimeProvider(0, slotDurationSeconds, slotsPerEpochShiftFactor)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, slotsPerEpochShiftFactor, generationRate, decayFactorEpochsSumShiftFactor, testManaDecayFactors, decayFactorsShiftFactor, testManaDecayFactorEpochsSum, decayFactorEpochsSumShiftFactor)

	// no decay in the same decay index
	value, err := manaDecayProvider.StoredManaWithDecay(100, 1, 200)
	require.NoError(t, err)
	require.Equal(t, iotago.Mana(100), value)
}

func TestManaDecay_Decay(t *testing.T) {
	timeProvider := iotago.NewTimeProvider(0, slotDurationSeconds, slotsPerEpochShiftFactor)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, slotsPerEpochShiftFactor, generationRate, decayFactorEpochsSumShiftFactor, testManaDecayFactors, decayFactorsShiftFactor, testManaDecayFactorEpochsSum, decayFactorEpochsSumShiftFactor)

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
