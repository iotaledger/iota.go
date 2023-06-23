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
	testDecayFactors = getTestManaDecayFactors(365)
)

func getTestManaDecayFactors(decayIndexes int) []uint32 {
	decayFactors := make([]uint32, decayIndexes)

	betaPerYear := 1 / 3.0
	betaPerDecayIndex := betaPerYear / 365.0

	for decayIndex := 1; decayIndex <= decayIndexes; decayIndex++ {
		decayFactor := math.Exp(-betaPerDecayIndex*float64(decayIndex)) * (math.Pow(2, float64(iotago.ManaDecayScaleFactor)))
		decayFactors[decayIndex-1] = uint32(decayFactor)
	}

	return decayFactors
}

func BenchmarkManaDecay_Single(b *testing.B) {
	timeProvider := iotago.NewTimeProvider(0, 10, 1<<13)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, 0, testDecayFactors, 10)

	endIndex := iotago.SlotIndex(300 << 13)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = manaDecayProvider.StoredManaWithDecay(manaValue53Bit, 0, endIndex)
	}
}

func BenchmarkManaDecay_Range(b *testing.B) {
	timeProvider := iotago.NewTimeProvider(0, 10, 1<<13)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, 0, testDecayFactors, 10)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var v uint64 = manaValue53Bit
		for decayIndex := 1; decayIndex <= 5*365; decayIndex++ {
			v = manaDecayProvider.StoredManaWithDecay(v, 0, iotago.SlotIndex(decayIndex)<<13)
		}
	}
}

func TestManaDecay_NoFactorsGiven(t *testing.T) {
	timeProvider := iotago.NewTimeProvider(0, 10, 1<<13)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, 0, []uint32{}, 10)

	require.Equal(t, uint64(100), manaDecayProvider.StoredManaWithDecay(100, 0, 100<<13))
}

func TestManaDecay_DecayIndexDiff(t *testing.T) {
	timeProvider := iotago.NewTimeProvider(0, 10, 1<<13)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, 0, testDecayFactors, 10)

	// no decay in the same decay index
	require.Equal(t, uint64(100), manaDecayProvider.StoredManaWithDecay(100, 1, 200))

	require.Panics(t, func() {
		manaDecayProvider.StoredManaWithDecay(100, 2<<13, 1<<13)
	})
}

func TestManaDecay_Decay(t *testing.T) {
	timeProvider := iotago.NewTimeProvider(0, 10, 1<<13)
	manaDecayProvider := iotago.NewManaDecayProvider(timeProvider, 0, testDecayFactors, 10)

	// check if mana decay works for multiples of the available decay indexes in the lookup table
	require.Equal(t, uint64(3310474560012284), manaDecayProvider.StoredManaWithDecay(manaValue53Bit, 0, iotago.SlotIndex(3*len(testDecayFactors))<<13))

	// check if mana decay works for exactly the  amount of decay indexes in the lookup table
	require.Equal(t, uint64(6451934231789564), manaDecayProvider.StoredManaWithDecay(manaValue53Bit, 0, iotago.SlotIndex(len(testDecayFactors))<<13))

	// check if mana decay works for 0 mana values
	require.Equal(t, uint64(0), manaDecayProvider.StoredManaWithDecay(0, 0, 400<<13))

	// even with the highest possible int64 number, the calculation should not overflow because of the overflow protection
	require.Equal(t, uint64(6398705774377299968), manaDecayProvider.StoredManaWithDecay(math.MaxInt64, 0, 400<<13))
}
