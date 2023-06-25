package iotago_test

import (
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

const (
	betaPerYear                  float64 = 1 / 3.0
	slotsPerEpochExponent                = 13
	slotDurationSeconds                  = 10
	generationRate                       = 1
	generationRateExponent               = 27
	decayFactorsExponent                 = 32
	decayFactorEpochsSumExponent         = 20
)

var (
	testManaDecayFactors         []uint32
	testManaDecayFactorEpochsSum uint32

	testTimeProvider      *iotago.TimeProvider
	testManaDecayProvider *iotago.ManaDecayProvider

	// These global variables are needed, otherwise the compiler will optimize away the actual tests.
	benchmarkResult iotago.Mana
)

func TestMain(m *testing.M) {
	testManaDecayFactors = tpkg.ManaDecayFactors(betaPerYear, 1<<slotsPerEpochExponent, slotDurationSeconds, decayFactorsExponent)
	testManaDecayFactorEpochsSum = tpkg.ManaDecayFactorEpochsSum(betaPerYear, 1<<slotsPerEpochExponent, slotDurationSeconds, decayFactorEpochsSumExponent)

	testTimeProvider = iotago.NewTimeProvider(0, slotDurationSeconds, slotsPerEpochExponent)
	testManaDecayProvider = iotago.NewManaDecayProvider(testTimeProvider, slotsPerEpochExponent, generationRate, decayFactorEpochsSumExponent, testManaDecayFactors, decayFactorsExponent, testManaDecayFactorEpochsSum, decayFactorEpochsSumExponent)

	// call the tests
	os.Exit(m.Run())
}

func BenchmarkStoredManaWithDecay_Single(b *testing.B) {
	endIndex := iotago.SlotIndex(300 << slotsPerEpochExponent)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult, _ = testManaDecayProvider.StoredManaWithDecay(math.MaxUint64, 0, endIndex)
	}
}

func BenchmarkStoredManaWithDecay_Range(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var value iotago.Mana = math.MaxUint64
		for epochIndex := 1; epochIndex <= 5*len(testManaDecayFactors); epochIndex++ {
			value, _ = testManaDecayProvider.StoredManaWithDecay(value, 0, iotago.SlotIndex(epochIndex)<<slotsPerEpochExponent)
		}
		benchmarkResult = value
	}
}

func BenchmarkPotentialManaWithDecay_Single(b *testing.B) {
	endIndex := iotago.SlotIndex(300 << slotsPerEpochExponent)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult, _ = testManaDecayProvider.PotentialManaWithDecay(math.MaxUint64, 0, endIndex)
	}
}

func BenchmarkPotentialManaWithDecay_Range(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var value iotago.Mana
		for epochIndex := 1; epochIndex <= 5*len(testManaDecayFactors); epochIndex++ {
			value, _ = testManaDecayProvider.PotentialManaWithDecay(math.MaxUint64, 0, iotago.SlotIndex(epochIndex)<<slotsPerEpochExponent)
		}
		benchmarkResult = value
	}
}

func TestManaDecay_NoFactorsGiven(t *testing.T) {
	manaDecayProvider := iotago.NewManaDecayProvider(testTimeProvider, slotsPerEpochExponent, generationRate, decayFactorEpochsSumExponent, []uint32{}, decayFactorsExponent, testManaDecayFactorEpochsSum, decayFactorEpochsSumExponent)

	value, err := manaDecayProvider.StoredManaWithDecay(100, 0, 100<<slotsPerEpochExponent)
	require.NoError(t, err)
	require.Equal(t, iotago.Mana(100), value)
}

func TestManaDecay_DecayIndexDiff(t *testing.T) {
	// no decay in the same decay index
	value, err := testManaDecayProvider.StoredManaWithDecay(100, 1, (1<<slotsPerEpochExponent)-1)
	require.NoError(t, err)
	require.Equal(t, iotago.Mana(100), value)
}

func TestManaDecay_Decay(t *testing.T) {
	{
		// check if mana decay works for multiples of the available decay indexes in the lookup table
		value, err := testManaDecayProvider.StoredManaWithDecay(math.MaxUint64, 0, iotago.SlotIndex(3*len(testManaDecayFactors))<<slotsPerEpochExponent)
		require.NoError(t, err)
		require.Equal(t, iotago.Mana(6803138682699798504), value)
	}

	{
		// check if mana decay works for exactly the amount of decay indexes in the lookup table
		value, err := testManaDecayProvider.StoredManaWithDecay(math.MaxUint64, 0, iotago.SlotIndex(len(testManaDecayFactors))<<slotsPerEpochExponent)
		require.NoError(t, err)
		require.Equal(t, iotago.Mana(13228672242897911807), value)
	}

	{
		// check if mana decay works for 0 mana values
		value, err := testManaDecayProvider.StoredManaWithDecay(0, 0, 400<<slotsPerEpochExponent)
		require.NoError(t, err)
		require.Equal(t, iotago.Mana(0), value)
	}

	{
		// even with the highest possible int64 number, the calculation should not overflow because of the overflow protection
		value, err := testManaDecayProvider.StoredManaWithDecay(math.MaxUint64, 0, 400<<slotsPerEpochExponent)
		require.NoError(t, err)
		require.Equal(t, iotago.Mana(13046663022640287317), value)
	}
}
