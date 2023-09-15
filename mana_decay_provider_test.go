//nolint:scopelint
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
	bitsCount                            = 63
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
	manaStruct := &iotago.ManaStructure{
		ManaBitsCount:                    bitsCount,
		ManaGenerationRate:               generationRate,
		ManaGenerationRateExponent:       generationRateExponent,
		ManaDecayFactors:                 testManaDecayFactors,
		ManaDecayFactorsExponent:         decayFactorsExponent,
		ManaDecayFactorEpochsSum:         testManaDecayFactorEpochsSum,
		ManaDecayFactorEpochsSumExponent: decayFactorEpochsSumExponent,
	}
	testManaDecayProvider = iotago.NewManaDecayProvider(testTimeProvider, slotsPerEpochExponent, manaStruct)

	// call the tests
	os.Exit(m.Run())
}

func BenchmarkManaWithDecay_Single(b *testing.B) {
	endIndex := iotago.SlotIndex(300 << slotsPerEpochExponent)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult, _ = testManaDecayProvider.ManaWithDecay(math.MaxUint64, 0, endIndex)
	}
}

func BenchmarkManaWithDecay_Range(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var value iotago.Mana = math.MaxUint64
		for epochIndex := 1; epochIndex <= 5*len(testManaDecayFactors); epochIndex++ {
			value, _ = testManaDecayProvider.ManaWithDecay(value, 0, iotago.SlotIndex(epochIndex)<<slotsPerEpochExponent)
		}
		benchmarkResult = value
	}
}

func BenchmarkManaGenerationWithDecay_Single(b *testing.B) {
	endIndex := iotago.SlotIndex(300 << slotsPerEpochExponent)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult, _ = testManaDecayProvider.ManaGenerationWithDecay(math.MaxUint64, 0, endIndex)
	}
}

func BenchmarkManaGenerationWithDecay_Range(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var value iotago.Mana
		for epochIndex := 1; epochIndex <= 5*len(testManaDecayFactors); epochIndex++ {
			value, _ = testManaDecayProvider.ManaGenerationWithDecay(math.MaxUint64, 0, iotago.SlotIndex(epochIndex)<<slotsPerEpochExponent)
		}
		benchmarkResult = value
	}
}

func TestManaDecay_NoFactorsGiven(t *testing.T) {
	manaStruct := &iotago.ManaStructure{
		ManaBitsCount:                    bitsCount,
		ManaGenerationRate:               generationRate,
		ManaGenerationRateExponent:       decayFactorEpochsSumExponent,
		ManaDecayFactors:                 []uint32{},
		ManaDecayFactorsExponent:         decayFactorsExponent,
		ManaDecayFactorEpochsSum:         testManaDecayFactorEpochsSum,
		ManaDecayFactorEpochsSumExponent: decayFactorEpochsSumExponent,
	}
	manaDecayProvider := iotago.NewManaDecayProvider(testTimeProvider, slotsPerEpochExponent, manaStruct)

	// no mana decay if no decay parameters are given
	value, err := manaDecayProvider.ManaWithDecay(100, testTimeProvider.EpochStart(1), testTimeProvider.EpochStart(100))
	require.NoError(t, err)
	require.Equal(t, iotago.Mana(100), value)
}

func TestManaDecay_NoEpochIndexDiff(t *testing.T) {
	// no decay in the same epoch
	value, err := testManaDecayProvider.ManaWithDecay(100, testTimeProvider.EpochStart(1), testTimeProvider.EpochEnd(1))
	require.NoError(t, err)
	require.Equal(t, iotago.Mana(100), value)
}

func TestManaDecay_StoredMana(t *testing.T) {
	type test struct {
		name             string
		storedMana       iotago.Mana
		slotIndexCreated iotago.SlotIndex
		slotIndexTarget  iotago.SlotIndex
		result           iotago.Mana
		wantErr          error
	}

	tests := []test{
		{
			name:             "check if mana decay works for 0 mana values",
			storedMana:       0,
			slotIndexCreated: testTimeProvider.EpochStart(1),
			slotIndexTarget:  testTimeProvider.EpochStart(400),
			result:           0,
			wantErr:          nil,
		},
		{
			name:             "check if mana decay works for 0 slot index diffs",
			storedMana:       math.MaxInt64,
			slotIndexCreated: testTimeProvider.EpochStart(1),
			slotIndexTarget:  testTimeProvider.EpochStart(1),
			result:           math.MaxInt64,
			wantErr:          nil,
		},
		{
			name:             "check for error if target index is lower than created index",
			storedMana:       0,
			slotIndexCreated: testTimeProvider.EpochStart(2),
			slotIndexTarget:  testTimeProvider.EpochStart(1),
			result:           0,
			wantErr:          iotago.ErrWrongEpochIndex,
		},
		{
			name:             "check if mana decay works for exactly the amount of epoch indexes in the lookup table",
			storedMana:       math.MaxUint64,
			slotIndexCreated: testTimeProvider.EpochStart(1),
			slotIndexTarget:  testTimeProvider.EpochStart(iotago.EpochIndex(len(testManaDecayFactors) + 1)),
			result:           13228672242897911807,
			wantErr:          nil,
		},
		{
			name:             "check if mana decay works for multiples of the available epoch indexes in the lookup table",
			storedMana:       math.MaxUint64,
			slotIndexCreated: testTimeProvider.EpochStart(1),
			slotIndexTarget:  testTimeProvider.EpochStart(iotago.EpochIndex(3*len(testManaDecayFactors) + 1)),
			result:           6803138682699798504,
			wantErr:          nil,
		},
		{
			name:             "even with the highest possible uint64 number, the calculation should not overflow",
			storedMana:       math.MaxUint64,
			slotIndexCreated: testTimeProvider.EpochStart(1),
			slotIndexTarget:  testTimeProvider.EpochStart(401),
			result:           13046663022640287317,
			wantErr:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testManaDecayProvider.ManaWithDecay(tt.storedMana, tt.slotIndexCreated, tt.slotIndexTarget)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}
			require.Equal(t, tt.result, result)
		})
	}
}

func TestManaDecay_PotentialMana(t *testing.T) {
	type test struct {
		name             string
		amount           iotago.BaseToken
		slotIndexCreated iotago.SlotIndex
		slotIndexTarget  iotago.SlotIndex
		result           iotago.Mana
		wantErr          error
	}

	tests := []test{
		{
			name:             "check if mana decay works for 0 base token values",
			amount:           0,
			slotIndexCreated: testTimeProvider.EpochStart(1),
			slotIndexTarget:  testTimeProvider.EpochStart(400),
			result:           0,
			wantErr:          nil,
		},
		{
			name:             "check if mana decay works for 0 slot index diffs",
			amount:           math.MaxInt64,
			slotIndexCreated: testTimeProvider.EpochStart(1),
			slotIndexTarget:  testTimeProvider.EpochStart(1),
			result:           0,
			wantErr:          nil,
		},
		{
			name:             "check for error if target index is lower than created index",
			amount:           0,
			slotIndexCreated: testTimeProvider.EpochStart(2),
			slotIndexTarget:  testTimeProvider.EpochStart(1),
			result:           0,
			wantErr:          iotago.ErrWrongEpochIndex,
		},
		{
			name:             "check if mana decay works for exactly the amount of epoch indexes in the lookup table",
			amount:           math.MaxInt64,
			slotIndexCreated: testTimeProvider.EpochStart(1),
			slotIndexTarget:  testTimeProvider.EpochStart(iotago.EpochIndex(len(testManaDecayFactors) + 1)),
			result:           183827294847826527,
			wantErr:          nil,
		},
		{
			name:             "check if mana decay works for multiples of the available epoch indexes in the lookup table",
			amount:           math.MaxInt64,
			slotIndexCreated: testTimeProvider.EpochStart(1),
			slotIndexTarget:  testTimeProvider.EpochStart(iotago.EpochIndex(3*len(testManaDecayFactors) + 1)),
			result:           410192222442040018,
			wantErr:          nil,
		},
		{
			name:             "check if mana generation works for 0 epoch index diffs",
			amount:           math.MaxInt64,
			slotIndexCreated: testTimeProvider.EpochStart(1),
			slotIndexTarget:  testTimeProvider.EpochEnd(1),
			result:           562881233944575,
			wantErr:          nil,
		},
		{
			name:             "check if mana generation works for 1 epoch index diffs",
			amount:           math.MaxInt64,
			slotIndexCreated: testTimeProvider.EpochStart(1),
			slotIndexTarget:  testTimeProvider.EpochEnd(2),
			result:           1125343946211326,
			wantErr:          nil,
		},
		{
			name:             "check if mana generation works for >=2 epoch index diffs",
			amount:           math.MaxInt64,
			slotIndexCreated: testTimeProvider.EpochStart(1),
			slotIndexTarget:  testTimeProvider.EpochEnd(3),
			result:           1687319975062367,
			wantErr:          nil,
		},
		{
			name:             "even with the highest possible int64 number, the calculation should not overflow",
			amount:           math.MaxInt64,
			slotIndexCreated: testTimeProvider.EpochStart(1),
			slotIndexTarget:  testTimeProvider.EpochStart(401),
			result:           190239292158065300,
			wantErr:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testManaDecayProvider.ManaGenerationWithDecay(tt.amount, tt.slotIndexCreated, tt.slotIndexTarget)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}
			require.Equal(t, tt.result, result)
		})
	}
}

func TestManaDecay_Rewards(t *testing.T) {
	type test struct {
		name              string
		rewards           iotago.Mana
		epochIndexReward  iotago.EpochIndex
		epochIndexClaimed iotago.EpochIndex
		result            iotago.Mana
		wantErr           error
	}

	tests := []test{
		{
			name:              "check if mana decay works for 0 mana values",
			rewards:           0,
			epochIndexReward:  1,
			epochIndexClaimed: 400,
			result:            0,
			wantErr:           nil,
		},
		{
			name:              "check if mana decay works for 0 slot index diffs",
			rewards:           math.MaxInt64,
			epochIndexReward:  1,
			epochIndexClaimed: 1,
			result:            math.MaxInt64,
			wantErr:           nil,
		},
		{
			name:              "check for error if target index is lower than created index",
			rewards:           0,
			epochIndexReward:  2,
			epochIndexClaimed: 1,
			result:            0,
			wantErr:           iotago.ErrWrongEpochIndex,
		},
		{
			name:              "check if mana decay works for exactly the amount of epoch indexes in the lookup table",
			rewards:           math.MaxUint64,
			epochIndexReward:  1,
			epochIndexClaimed: iotago.EpochIndex(len(testManaDecayFactors) + 1),
			result:            13228672242897911807,
			wantErr:           nil,
		},
		{
			name:              "check if mana decay works for multiples of the available epoch indexes in the lookup table",
			rewards:           math.MaxUint64,
			epochIndexReward:  1,
			epochIndexClaimed: iotago.EpochIndex(3*len(testManaDecayFactors) + 1),
			result:            6803138682699798504,
			wantErr:           nil,
		},
		{
			name:              "even with the highest possible uint64 number, the calculation should not overflow",
			rewards:           math.MaxUint64,
			epochIndexReward:  1,
			epochIndexClaimed: 401,
			result:            13046663022640287317,
			wantErr:           nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testManaDecayProvider.RewardsWithDecay(tt.rewards, tt.epochIndexReward, tt.epochIndexClaimed)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}
			require.Equal(t, tt.result, result)
		})
	}
}
