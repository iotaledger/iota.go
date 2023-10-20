//nolint:scopelint,golint,revive,nosnakecase,stylecheck
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

	testTimeProvider           *iotago.TimeProvider
	testManaDecayProvider      *iotago.ManaDecayProvider
	testFloatManaDecayProvider *TestFloatManaDecayProvider

	// These global variables are needed, otherwise the compiler will optimize away the actual tests.
	benchmarkResult iotago.Mana
)

func TestMain(m *testing.M) {
	testManaDecayFactors = tpkg.ManaDecayFactors(betaPerYear, 1<<slotsPerEpochExponent, slotDurationSeconds, decayFactorsExponent)
	testManaDecayFactorEpochsSum = tpkg.ManaDecayFactorEpochsSum(betaPerYear, 1<<slotsPerEpochExponent, slotDurationSeconds, decayFactorEpochsSumExponent)

	testTimeProvider = iotago.NewTimeProvider(0, slotDurationSeconds, slotsPerEpochExponent)
	manaStruct := &iotago.ManaParameters{
		BitsCount:                    bitsCount,
		GenerationRate:               generationRate,
		GenerationRateExponent:       generationRateExponent,
		DecayFactors:                 testManaDecayFactors,
		DecayFactorsExponent:         decayFactorsExponent,
		DecayFactorEpochsSum:         testManaDecayFactorEpochsSum,
		DecayFactorEpochsSumExponent: decayFactorEpochsSumExponent,
	}
	testManaDecayProvider = iotago.NewManaDecayProvider(testTimeProvider, slotsPerEpochExponent, manaStruct)
	testFloatManaDecayProvider = &TestFloatManaDecayProvider{
		timeProvider:           testTimeProvider,
		generationRate:         uint64(manaStruct.GenerationRate),
		generationRateExponent: uint64(manaStruct.GenerationRateExponent),
	}

	// call the tests
	os.Exit(m.Run())
}

func BenchmarkManaWithDecay_Single(b *testing.B) {
	endSlot := iotago.SlotIndex(300 << slotsPerEpochExponent)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult, _ = testManaDecayProvider.ManaWithDecay(iotago.MaxMana, 0, endSlot)
	}
}

func BenchmarkManaWithDecay_Range(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		value := iotago.MaxMana
		for epoch := 1; epoch <= 5*len(testManaDecayFactors); epoch++ {
			value, _ = testManaDecayProvider.ManaWithDecay(value, 0, iotago.SlotIndex(epoch)<<slotsPerEpochExponent)
		}
		benchmarkResult = value
	}
}

func BenchmarkManaGenerationWithDecay_Single(b *testing.B) {
	endIndex := iotago.SlotIndex(300 << slotsPerEpochExponent)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult, _ = testManaDecayProvider.ManaGenerationWithDecay(iotago.MaxBaseToken, 0, endIndex)
	}
}

func BenchmarkManaGenerationWithDecay_Range(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var value iotago.Mana
		for epoch := 1; epoch <= 5*len(testManaDecayFactors); epoch++ {
			value, _ = testManaDecayProvider.ManaGenerationWithDecay(iotago.MaxBaseToken, 0, iotago.SlotIndex(epoch)<<slotsPerEpochExponent)
		}
		benchmarkResult = value
	}
}

func TestManaDecay_NoFactorsGiven(t *testing.T) {
	manaStruct := &iotago.ManaParameters{
		BitsCount:                    bitsCount,
		GenerationRate:               generationRate,
		GenerationRateExponent:       decayFactorEpochsSumExponent,
		DecayFactors:                 []uint32{},
		DecayFactorsExponent:         decayFactorsExponent,
		DecayFactorEpochsSum:         testManaDecayFactorEpochsSum,
		DecayFactorEpochsSumExponent: decayFactorEpochsSumExponent,
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
		name        string
		storedMana  iotago.Mana
		createdSlot iotago.SlotIndex
		targetSlot  iotago.SlotIndex
		result      iotago.Mana
		wantErr     error
	}

	tests := []*test{
		{
			name:        "check if mana decay works for 0 mana values",
			storedMana:  0,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(400),
			result:      0,
			wantErr:     nil,
		},
		{
			name:        "check if mana decay works for 0 slot index diffs",
			storedMana:  iotago.MaxMana,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(1),
			result:      iotago.MaxMana,
			wantErr:     nil,
		},
		{
			name:        "check for error if target index is lower than created index",
			storedMana:  0,
			createdSlot: testTimeProvider.EpochStart(2),
			targetSlot:  testTimeProvider.EpochStart(1),
			result:      0,
			wantErr:     iotago.ErrWrongEpochIndex,
		},
		{
			name:        "check if mana decay works for exactly the amount of epochs in the lookup table",
			storedMana:  iotago.MaxMana,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(iotago.EpochIndex(len(testManaDecayFactors) + 1)),
			result:      13228672242897911807,
			wantErr:     nil,
		},
		{
			name:        "check if mana decay works for multiples of the available epochs in the lookup table",
			storedMana:  iotago.MaxMana,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(iotago.EpochIndex(3*len(testManaDecayFactors) + 1)),
			result:      6803138682699798504,
			wantErr:     nil,
		},
		{
			name:        "even with the highest possible uint64 number, the calculation should not overflow",
			storedMana:  iotago.MaxMana,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(401),
			result:      13046663022640287317,
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testManaDecayProvider.ManaWithDecay(tt.storedMana, tt.createdSlot, tt.targetSlot)
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
		name        string
		amount      iotago.BaseToken
		createdSlot iotago.SlotIndex
		targetSlot  iotago.SlotIndex
		result      iotago.Mana
		wantErr     error
	}

	tests := []*test{
		{
			name:        "check if mana decay works for 0 base token values",
			amount:      0,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(400),
			result:      0,
			wantErr:     nil,
		},
		{
			name:        "check if mana decay works for 0 slot index diffs",
			amount:      math.MaxInt64,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(1),
			result:      0,
			wantErr:     nil,
		},
		{
			name:        "check for error if target index is lower than created index",
			amount:      0,
			createdSlot: testTimeProvider.EpochStart(2),
			targetSlot:  testTimeProvider.EpochStart(1),
			result:      0,
			wantErr:     iotago.ErrWrongEpochIndex,
		},
		{
			name:        "check if mana decay works for exactly the amount of epochs in the lookup table",
			amount:      math.MaxInt64,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(iotago.EpochIndex(len(testManaDecayFactors) + 1)),
			result:      183827295065703076,
			wantErr:     nil,
		},
		{
			name:        "check if mana decay works for multiples of the available epochs in the lookup table",
			amount:      math.MaxInt64,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(iotago.EpochIndex(3*len(testManaDecayFactors) + 1)),
			result:      410192223115924783,
			wantErr:     nil,
		},
		{
			name:        "check if mana generation works for 0 epoch diffs",
			amount:      math.MaxInt64,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochEnd(1),
			result:      562881233944575,
			wantErr:     nil,
		},
		{
			name:        "check if mana generation works for 1 epoch diffs",
			amount:      math.MaxInt64,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochEnd(2),
			result:      1125343946211326,
			wantErr:     nil,
		},
		{
			name:        "check if mana generation works for >=2 epoch diffs",
			amount:      math.MaxInt64,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochEnd(3),
			result:      1687319824887185,
			wantErr:     nil,
		},
		{
			name:        "even with the highest possible int64 number, the calculation should not overflow",
			amount:      math.MaxInt64,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(401),
			result:      190239292388858706,
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// calculate the result
			result, err := testManaDecayProvider.ManaGenerationWithDecay(tt.amount, tt.createdSlot, tt.targetSlot)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			// calculate the bounds
			upperBound := testFloatManaDecayProvider.UpperBoundPotentialMana(tt.amount, tt.createdSlot, tt.targetSlot)
			lowerBound := testFloatManaDecayProvider.LowerBoundPotentialMana(tt.amount, tt.createdSlot, tt.targetSlot)
			floatResult := testFloatManaDecayProvider.ManaGenerationWithDecay(tt.amount, tt.createdSlot, tt.targetSlot)

			// check if the result is in the bounds
			require.LessOrEqual(t, float64(result), upperBound)
			require.GreaterOrEqual(t, float64(result), lowerBound)

			// for epsilon check it must not be zero
			if result != 0 {
				require.InEpsilon(t, floatResult, float64(result), float64(0.001))
			}

			// check for the exact precomputed result
			require.Equal(t, tt.result, result)
		})
	}
}

func TestManaDecay_Rewards(t *testing.T) {
	type test struct {
		name         string
		rewards      iotago.Mana
		rewardEpoch  iotago.EpochIndex
		claimedEpoch iotago.EpochIndex
		result       iotago.Mana
		wantErr      error
	}

	tests := []*test{
		{
			name:         "check if mana decay works for 0 mana values",
			rewards:      0,
			rewardEpoch:  1,
			claimedEpoch: 400,
			result:       0,
			wantErr:      nil,
		},
		{
			name:         "check if mana decay works for 0 slot index diffs",
			rewards:      iotago.MaxMana,
			rewardEpoch:  1,
			claimedEpoch: 1,
			result:       iotago.MaxMana,
			wantErr:      nil,
		},
		{
			name:         "check for error if target index is lower than created index",
			rewards:      0,
			rewardEpoch:  2,
			claimedEpoch: 1,
			result:       0,
			wantErr:      iotago.ErrWrongEpochIndex,
		},
		{
			name:         "check if mana decay works for exactly the amount of epochs in the lookup table",
			rewards:      iotago.MaxMana,
			rewardEpoch:  1,
			claimedEpoch: iotago.EpochIndex(len(testManaDecayFactors) + 1),
			result:       13228672242897911807,
			wantErr:      nil,
		},
		{
			name:         "check if mana decay works for multiples of the available epochs in the lookup table",
			rewards:      iotago.MaxMana,
			rewardEpoch:  1,
			claimedEpoch: iotago.EpochIndex(3*len(testManaDecayFactors) + 1),
			result:       6803138682699798504,
			wantErr:      nil,
		},
		{
			name:         "even with the highest possible uint64 number, the calculation should not overflow",
			rewards:      iotago.MaxMana,
			rewardEpoch:  1,
			claimedEpoch: 401,
			result:       13046663022640287317,
			wantErr:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testManaDecayProvider.RewardsWithDecay(tt.rewards, tt.rewardEpoch, tt.claimedEpoch)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}
			require.Equal(t, tt.result, result)
		})
	}
}

// TestFloatManaDecayProvider calculates the mana decay and mana generation
// using floating point arithmetic.
type TestFloatManaDecayProvider struct {
	timeProvider *iotago.TimeProvider

	// generationRate is the amount of potential Mana generated by 1 IOTA in 1 slot.
	generationRate uint64 // the generation rate needs to be scaled by 2^-generationRateExponent

	// generationRateExponent is the scaling of generationRate expressed as an exponent of 2.
	generationRateExponent uint64
}

func (p *TestFloatManaDecayProvider) ManaGenerationWithDecay(amount iotago.BaseToken, creationSlot iotago.SlotIndex, targetSlot iotago.SlotIndex) float64 {
	creationEpoch := p.timeProvider.EpochFromSlot(creationSlot)
	targetEpoch := p.timeProvider.EpochFromSlot(targetSlot)
	floatAmount := float64(amount)
	floatGenerationRate := float64(p.generationRate) * math.Pow(2, -float64(p.generationRateExponent))
	delta := float64(1<<13) * (1.0 / (365.0 * 24.0 * 60.0 * 60.0)) * float64(10)
	epochDiff := targetEpoch - creationEpoch

	//nolint:exhaustive // false-positive, we have a default case
	switch epochDiff {
	case 0:
		floatSlotDiff := float64(targetSlot - creationSlot)

		return floatSlotDiff * floatAmount * floatGenerationRate

	case 1:
		slotsBeforeNextEpoch := p.timeProvider.SlotsBeforeNextEpoch(creationSlot)
		slotsSinceEpochStart := p.timeProvider.SlotsSinceEpochStart(targetSlot)
		manaDecayed := float64(slotsBeforeNextEpoch) * floatAmount * floatGenerationRate * math.Exp(-delta/3)
		manaGenerated := float64(slotsSinceEpochStart) * floatAmount * floatGenerationRate

		return manaDecayed + manaGenerated

	default:
		slotsBeforeNextEpoch := p.timeProvider.SlotsBeforeNextEpoch(creationSlot)
		slotsSinceEpochStart := p.timeProvider.SlotsSinceEpochStart(targetSlot)
		epochDiffFloat := float64(epochDiff)
		slotsPerEpochFloat := math.Pow(2, 13)
		constant := math.Exp(-delta/3) * (1 - math.Exp(-(epochDiffFloat-1)*delta/3)) / (1 - math.Exp(-delta/3))
		potentialMana_n := float64(slotsBeforeNextEpoch) * floatAmount * floatGenerationRate * math.Exp(-epochDiffFloat*delta/3)
		potentialMana_n_1 := constant * floatAmount * floatGenerationRate * slotsPerEpochFloat
		potentialMana_0 := float64(slotsSinceEpochStart) * floatAmount * floatGenerationRate

		return potentialMana_n + potentialMana_n_1 + potentialMana_0
	}
}

func (p *TestFloatManaDecayProvider) LowerBoundPotentialMana(amount iotago.BaseToken, creationSlot iotago.SlotIndex, targetSlot iotago.SlotIndex) float64 {
	delta := float64(1<<13) * (1.0 / (365.0 * 24.0 * 60.0 * 60.0)) * float64(10)
	constant := math.Exp(-delta/3) / (1 - math.Exp(-delta/3))
	return p.ManaGenerationWithDecay(amount, creationSlot, targetSlot) - (4 + float64(amount)*math.Pow(2, float64(13-27))*(1+constant*math.Pow(2, -float64(32))))
}

func (p *TestFloatManaDecayProvider) UpperBoundPotentialMana(amount iotago.BaseToken, creationSlot iotago.SlotIndex, targetSlot iotago.SlotIndex) float64 {
	return p.ManaGenerationWithDecay(amount, creationSlot, targetSlot) + 2 - math.Pow(2, -float64(32-1))
}
