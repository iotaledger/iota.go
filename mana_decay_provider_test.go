//nolint:golint,revive,stylecheck
package iotago_test

import (
	"math"
	"os"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/safemath"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

var (
	testProtoParams            = tpkg.IOTAMainnetV3TestProtocolParameters
	testAPI                    = iotago.V3API(testProtoParams)
	testTimeProvider           = testAPI.TimeProvider()
	testManaDecayProvider      *iotago.ManaDecayProvider
	testFloatManaDecayProvider *TestReferenceManaDecayProvider

	// These global variables are needed, otherwise the compiler will optimize away the actual tests.
	benchmarkResult iotago.Mana
)

func TestMain(m *testing.M) {
	manaParams := testProtoParams.ManaParameters()
	testManaDecayProvider = iotago.NewManaDecayProvider(testTimeProvider, testProtoParams.SlotsPerEpochExponent(), manaParams)
	testFloatManaDecayProvider = &TestReferenceManaDecayProvider{
		timeProvider:                 testTimeProvider,
		generationRate:               uint64(manaParams.GenerationRate),
		generationRateExponent:       uint64(manaParams.GenerationRateExponent),
		decayFactorEpochsSum:         uint64(manaParams.DecayFactorEpochsSum),
		decayFactorEpochsSumExponent: uint64(manaParams.DecayFactorEpochsSumExponent),
		decayFactorsExponent:         uint64(manaParams.DecayFactorsExponent),
		decayFactors:                 manaParams.DecayFactors,
		decayFactorsLength:           uint64(len(manaParams.DecayFactors)),
		annualDecayFactorPercentage:  uint64(manaParams.AnnualDecayFactorPercentage),
		tokenSupply:                  testProtoParams.TokenSupply(),
	}

	// call the tests
	os.Exit(m.Run())
}

func BenchmarkManaWithDecay_Single(b *testing.B) {
	endSlot := iotago.SlotIndex(300 << testProtoParams.SlotsPerEpochExponent())

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult, _ = testManaDecayProvider.DecayManaBySlots(iotago.MaxMana, 0, endSlot)
	}
}

func BenchmarkManaWithDecay_Range(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		value := iotago.MaxMana
		for epoch := 1; epoch <= 5*len(testProtoParams.ManaParameters().DecayFactors); epoch++ {
			value, _ = testManaDecayProvider.DecayManaBySlots(value, 0, iotago.SlotIndex(epoch)<<testProtoParams.SlotsPerEpochExponent())
		}
		benchmarkResult = value
	}
}

func BenchmarkManaGenerationWithDecay_Single(b *testing.B) {
	endIndex := iotago.SlotIndex(300 << testProtoParams.SlotsPerEpochExponent())

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkResult, _ = testManaDecayProvider.GenerateManaAndDecayBySlots(iotago.MaxBaseToken, 0, endIndex)
	}
}

func BenchmarkManaGenerationWithDecay_Range(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var value iotago.Mana
		for epoch := 1; epoch <= 5*len(testProtoParams.ManaParameters().DecayFactors); epoch++ {
			value, _ = testManaDecayProvider.GenerateManaAndDecayBySlots(iotago.MaxBaseToken, 0, iotago.SlotIndex(epoch)<<testProtoParams.SlotsPerEpochExponent())
		}
		benchmarkResult = value
	}
}

func TestManaDecay_NoEpochIndexDiff(t *testing.T) {
	// no decay in the same epoch
	value, err := testManaDecayProvider.DecayManaBySlots(100, testTimeProvider.EpochStart(1), testTimeProvider.EpochEnd(1))
	require.NoError(t, err)
	require.Equal(t, iotago.Mana(100), value)
}

func TestManaDecay_StoredMana(t *testing.T) {
	type test struct {
		name        string
		storedMana  iotago.Mana
		createdSlot iotago.SlotIndex
		targetSlot  iotago.SlotIndex
		wantErr     error
	}

	tests := []*test{
		{
			name:        "check if mana decay works for 0 mana values",
			storedMana:  0,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(400),
			wantErr:     nil,
		},
		{
			name:        "check if mana decay works for 0 slot index diffs",
			storedMana:  iotago.MaxMana,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(1),
			wantErr:     nil,
		},
		{
			name:        "check for error if target index is lower than created index",
			storedMana:  0,
			createdSlot: testTimeProvider.EpochStart(2),
			targetSlot:  testTimeProvider.EpochStart(1),
			wantErr:     iotago.ErrManaDecayCreationIndexExceedsTargetIndex,
		},
		{
			name:        "check if mana decay works for exactly the amount of epochs in the lookup table",
			storedMana:  iotago.MaxMana,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(iotago.EpochIndex(len(testProtoParams.ManaParameters().DecayFactors) + 1)),
			wantErr:     nil,
		},
		{
			name:        "check if mana decay works for multiples of the available epochs in the lookup table",
			storedMana:  iotago.MaxMana,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(iotago.EpochIndex(3*len(testProtoParams.ManaParameters().DecayFactors) + 1)),
			wantErr:     nil,
		},
		{
			name:        "even with the highest possible uint64 number, the calculation should not overflow",
			storedMana:  iotago.MaxMana,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(401),
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixedPointResult, err := testManaDecayProvider.DecayManaBySlots(tt.storedMana, tt.createdSlot, tt.targetSlot)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			// calculate the bounds of the float result.
			upperBound := testFloatManaDecayProvider.UpperBoundManaDecay(tt.storedMana, tt.createdSlot, tt.targetSlot)
			lowerBound := testFloatManaDecayProvider.LowerBoundManaDecay(tt.storedMana, tt.createdSlot, tt.targetSlot)

			// check if the fixed point result is in the bounds.
			require.LessOrEqual(t, float64(fixedPointResult), upperBound)
			require.GreaterOrEqual(t, float64(fixedPointResult), lowerBound)

			// check the fixed point result is exactly equal to the 256-bit integer result.
			result256 := testFloatManaDecayProvider.ManaDecay256(tt.storedMana, tt.createdSlot, tt.targetSlot)
			require.Equal(t, uint64(fixedPointResult), uint64(result256))

		})
	}
}

func TestManaDecay_PotentialMana(t *testing.T) {
	type test struct {
		name        string
		amount      iotago.BaseToken
		createdSlot iotago.SlotIndex
		targetSlot  iotago.SlotIndex
		wantErr     error
	}

	tests := []*test{
		{
			name:        "check if mana decay works for 0 base token values",
			amount:      0,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(400),
			wantErr:     nil,
		},
		{
			name:        "check if mana decay works for 0 slot index diffs",
			amount:      math.MaxInt64,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(1),
			wantErr:     nil,
		},
		{
			name:        "check for error if target index is lower than created index",
			amount:      0,
			createdSlot: testTimeProvider.EpochStart(2),
			targetSlot:  testTimeProvider.EpochStart(1),
			wantErr:     iotago.ErrManaDecayCreationIndexExceedsTargetIndex,
		},
		{
			name:        "check if mana decay works for exactly the amount of epochs in the lookup table",
			amount:      testFloatManaDecayProvider.tokenSupply,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(iotago.EpochIndex(len(testProtoParams.ManaParameters().DecayFactors) + 1)),
			wantErr:     nil,
		},
		{
			name:        "check if mana decay works for multiples of the available epochs in the lookup table",
			amount:      testFloatManaDecayProvider.tokenSupply,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochStart(iotago.EpochIndex(3*len(testProtoParams.ManaParameters().DecayFactors) + 1)),
			wantErr:     nil,
		},
		{
			name:        "check if mana generation works for 0 epoch diffs",
			amount:      testFloatManaDecayProvider.tokenSupply,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochEnd(1),
			wantErr:     nil,
		},
		{
			name:        "check if mana generation works for 1 epoch diffs",
			amount:      testFloatManaDecayProvider.tokenSupply,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochEnd(2),
			wantErr:     nil,
		},
		{
			name:        "check if mana generation works for >=2 epoch diffs",
			amount:      testFloatManaDecayProvider.tokenSupply,
			createdSlot: testTimeProvider.EpochStart(1),
			targetSlot:  testTimeProvider.EpochEnd(3),
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// calculate the fixedPointResult
			fixedPointResult, err := testManaDecayProvider.GenerateManaAndDecayBySlots(tt.amount, tt.createdSlot, tt.targetSlot)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}

			// calculate the bounds of the float result.
			upperBound := testFloatManaDecayProvider.UpperBoundManaGenerationWithDecay(tt.amount, tt.createdSlot, tt.targetSlot)
			lowerBound := testFloatManaDecayProvider.LowerBoundManaGenerationWithDecay(tt.amount, tt.createdSlot, tt.targetSlot)

			// check if the fixed point result is in the bounds.
			require.LessOrEqual(t, float64(fixedPointResult), upperBound)
			require.GreaterOrEqual(t, float64(fixedPointResult), lowerBound)

			// check the fixed point result is within epsilon of float result (must not be zero)
			if fixedPointResult != 0 {
				floatResult := testFloatManaDecayProvider.ManaGenerationWithDecayFloat(tt.amount, tt.createdSlot, tt.targetSlot)
				require.InEpsilon(t, floatResult, float64(fixedPointResult), float64(0.001))
			}

			// check the fixed point result is exactly equal to the 256-bit integer result
			result256 := testFloatManaDecayProvider.ManaGenerationWithDecay256(tt.amount, tt.createdSlot, tt.targetSlot)
			require.Equal(t, fixedPointResult, result256)

		})
	}
}

func TestManaDecay_Rewards(t *testing.T) {
	type test struct {
		name         string
		rewards      iotago.Mana
		rewardEpoch  iotago.EpochIndex
		claimedEpoch iotago.EpochIndex
		wantErr      error
	}

	tests := []*test{
		{
			name:         "check if mana decay works for 0 mana values",
			rewards:      0,
			rewardEpoch:  1,
			claimedEpoch: 400,
			wantErr:      nil,
		},
		{
			name:         "check if mana decay works for 0 slot index diffs",
			rewards:      iotago.MaxMana,
			rewardEpoch:  1,
			claimedEpoch: 1,
			wantErr:      nil,
		},
		{
			name:         "check for error if target index is lower than created index",
			rewards:      0,
			rewardEpoch:  2,
			claimedEpoch: 1,
			wantErr:      iotago.ErrManaDecayCreationIndexExceedsTargetIndex,
		},
		{
			name:         "check if mana decay works for exactly the amount of epochs in the lookup table",
			rewards:      iotago.MaxMana,
			rewardEpoch:  1,
			claimedEpoch: iotago.EpochIndex(len(testProtoParams.ManaParameters().DecayFactors) + 1),
			wantErr:      nil,
		},
		{
			name:         "check if mana decay works for multiples of the available epochs in the lookup table",
			rewards:      iotago.MaxMana,
			rewardEpoch:  1,
			claimedEpoch: iotago.EpochIndex(3*len(testProtoParams.ManaParameters().DecayFactors) + 1),
			wantErr:      nil,
		},
		{
			name:         "even with the highest possible uint64 number, the calculation should not overflow",
			rewards:      iotago.MaxMana,
			rewardEpoch:  1,
			claimedEpoch: 401,
			wantErr:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixedPointResult, err := testManaDecayProvider.DecayManaByEpochs(tt.rewards, tt.rewardEpoch, tt.claimedEpoch)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}

			// calculate bounds for the float result.
			createdSlot := testTimeProvider.EpochStart(tt.rewardEpoch)
			targetSlot := testTimeProvider.EpochStart(tt.claimedEpoch)
			testTimeProvider.EpochStart(1)
			upperBound := testFloatManaDecayProvider.UpperBoundManaDecay(tt.rewards, createdSlot, targetSlot)
			lowerBound := testFloatManaDecayProvider.LowerBoundManaDecay(tt.rewards, createdSlot, targetSlot)

			// check if the fixed point result is in the bounds.
			require.LessOrEqual(t, float64(fixedPointResult), upperBound)
			require.GreaterOrEqual(t, float64(fixedPointResult), lowerBound)

			// check the fixed point result is exactly equal to the 256-bit integer result.
			result256 := testFloatManaDecayProvider.decay256(uint64(tt.rewards), tt.claimedEpoch-tt.rewardEpoch)
			require.Equal(t, fixedPointResult, result256)
		})
	}
}

// TestReferenceManaDecayProvider calculates reference values for the mana decay and mana generation
// using either floating point arithmetic or 256-bit integer to compare against our fixed point decay provider.
type TestReferenceManaDecayProvider struct {
	timeProvider *iotago.TimeProvider

	// generationRate is the amount of potential Mana generated by 1 IOTA in 1 slot.
	generationRate uint64 // the generation rate needs to be scaled by 2^-generationRateExponent

	// generationRateExponent is the scaling of generationRate expressed as an exponent of 2.
	generationRateExponent uint64

	decayFactorEpochsSum uint64

	decayFactorEpochsSumExponent uint64

	decayFactorsExponent uint64

	decayFactorsLength uint64

	decayFactors []uint32 // the factors need to be scaled by 2^-decayFactorsExponent

	annualDecayFactorPercentage uint64

	tokenSupply iotago.BaseToken
}

// decay256 applies the decay to the given value using 256-bit integer arithmetic.
func (p *TestReferenceManaDecayProvider) decay256(value uint64, epochDiff iotago.EpochIndex) iotago.Mana {
	value256 := uint256.NewInt(value)

	// we keep applying the decay as long as epoch diffs are left
	remainingEpochDiff := epochDiff
	for remainingEpochDiff > 0 {
		// we can't decay more than the available epoch diffs
		// in the lookup table in this iteration
		diffsToDecay := remainingEpochDiff
		if diffsToDecay > iotago.EpochIndex(p.decayFactorsLength) {
			diffsToDecay = iotago.EpochIndex(p.decayFactorsLength)
		}
		remainingEpochDiff -= diffsToDecay
		// slice index 0 equals epoch diff 1
		decayFactor256 := uint256.NewInt(uint64(p.decayFactors[diffsToDecay-1]))
		// apply the decay and scale the resulting value (fixed-point arithmetics)
		value256 = new(uint256.Int).Mul(value256, decayFactor256)
		value256 = new(uint256.Int).Rsh(value256, uint(p.decayFactorsExponent))
	}

	return iotago.Mana(value256.Uint64())
}

// ManaDecay256 applies the decay to the given value using 256-bit integer arithmetic.
func (p *TestReferenceManaDecayProvider) ManaDecay256(value iotago.Mana, creationSlot iotago.SlotIndex, targetSlot iotago.SlotIndex) iotago.Mana {
	creationEpoch := p.timeProvider.EpochFromSlot(creationSlot)
	targetEpoch := p.timeProvider.EpochFromSlot(targetSlot)
	if value == 0 || targetEpoch-creationEpoch == 0 || p.decayFactorsLength == 0 {
		// no need to decay if the epoch didn't change or no decay factors were given
		return value
	}

	return p.decay256(uint64(value), targetEpoch-creationEpoch)
}

// ManaDecayFloat applies the decay to the given value using floating point arithmetic.
func (p *TestReferenceManaDecayProvider) ManaDecayFloat(mana iotago.Mana, creationSlot iotago.SlotIndex, targetSlot iotago.SlotIndex) float64 {
	creationEpoch := p.timeProvider.EpochFromSlot(creationSlot)
	targetEpoch := p.timeProvider.EpochFromSlot(targetSlot)
	floatAmount := float64(mana)
	epochsPerYear := (365.0 * 24.0 * 60.0 * 60.0) / float64(p.timeProvider.EpochDurationSeconds())
	decayPerEpoch := math.Pow(float64(p.annualDecayFactorPercentage)/100.0, 1/epochsPerYear)
	manaDecayed := floatAmount * math.Pow(decayPerEpoch, float64(targetEpoch-creationEpoch))

	return manaDecayed
}

// ManaGenerationWithDecay256 calculates mana generation with decay (for potential Mana) using 256-bit integer arithmetic.
func (p *TestReferenceManaDecayProvider) ManaGenerationWithDecay256(amount iotago.BaseToken, creationSlot iotago.SlotIndex, targetSlot iotago.SlotIndex) iotago.Mana {
	if targetSlot-creationSlot == 0 || p.generationRate == 0 {
		return 0
	}
	creationEpoch := p.timeProvider.EpochFromSlot(creationSlot)
	targetEpoch := p.timeProvider.EpochFromSlot(targetSlot)
	epochDiff := targetEpoch - creationEpoch
	amount256 := uint256.NewInt(uint64(amount))
	generationRate256 := uint256.NewInt(p.generationRate)
	decayFactorEpochsSum256 := uint256.NewInt(p.decayFactorEpochsSum)
	decayFactorEpochsSumExponent256 := p.decayFactorEpochsSumExponent

	//nolint:exhaustive // false-positive, we have a default case
	switch epochDiff {
	// case 0 means that the creationSlot and targetSlot belong to the same epoch. In that case, we generate mana according to the slotDiff, and no decay is applied
	case 0:
		slotDiff256 := uint256.NewInt(uint64(targetSlot - creationSlot))
		result := new(uint256.Int).Mul(generationRate256, amount256)
		result = new(uint256.Int).Mul(result, slotDiff256)
		result = new(uint256.Int).Rsh(result, uint(p.generationRateExponent))

		return iotago.Mana(result.Uint64())
	case 1:
		slotsBeforeNextEpoch256 := uint256.NewInt(uint64(p.timeProvider.SlotsBeforeNextEpoch(creationSlot)))
		slotsSinceEpochStart256 := uint256.NewInt(uint64(p.timeProvider.SlotsSinceEpochStart(targetSlot)))
		manaGeneratedFirstEpoch := new(uint256.Int).Mul(generationRate256, amount256)
		manaGeneratedFirstEpoch = new(uint256.Int).Mul(manaGeneratedFirstEpoch, slotsBeforeNextEpoch256)
		manaGeneratedFirstEpoch = new(uint256.Int).Rsh(manaGeneratedFirstEpoch, uint(p.generationRateExponent))

		manaDecayedFirstEpoch := p.decay256(manaGeneratedFirstEpoch.Uint64(), 1)

		manaGeneratedSecondEpoch := new(uint256.Int).Mul(generationRate256, amount256)
		manaGeneratedSecondEpoch = new(uint256.Int).Mul(manaGeneratedSecondEpoch, slotsSinceEpochStart256)
		manaGeneratedSecondEpoch = new(uint256.Int).Rsh(manaGeneratedSecondEpoch, uint(p.generationRateExponent))
		result, _ := safemath.SafeAdd(manaDecayedFirstEpoch, iotago.Mana(manaGeneratedSecondEpoch.Uint64()))

		return result
	default:
		slotsBeforeNextEpoch256 := uint256.NewInt(uint64(p.timeProvider.SlotsBeforeNextEpoch(creationSlot)))
		slotsSinceEpochStart256 := uint256.NewInt(uint64(p.timeProvider.SlotsSinceEpochStart(targetSlot)))

		c := new(uint256.Int).Mul(generationRate256, amount256)
		c = new(uint256.Int).Mul(decayFactorEpochsSum256, c)
		c = new(uint256.Int).Rsh(c, uint(decayFactorEpochsSumExponent256+p.generationRateExponent-uint64(p.timeProvider.SlotsPerEpochExponent())))

		manaGeneratedFirstEpoch := new(uint256.Int).Mul(generationRate256, amount256)
		manaGeneratedFirstEpoch = new(uint256.Int).Mul(manaGeneratedFirstEpoch, slotsBeforeNextEpoch256)
		manaGeneratedFirstEpoch = new(uint256.Int).Rsh(manaGeneratedFirstEpoch, uint(p.generationRateExponent))
		manaDecayedFirstEpoch := p.decay256(manaGeneratedFirstEpoch.Uint64(), epochDiff)

		manaDecayedIntermediateEpochs := p.decay256(c.Uint64(), epochDiff-1)

		manaGeneratedLastEpoch := new(uint256.Int).Mul(generationRate256, amount256)
		manaGeneratedLastEpoch = new(uint256.Int).Mul(manaGeneratedLastEpoch, slotsSinceEpochStart256)
		manaGeneratedLastEpoch = new(uint256.Int).Rsh(manaGeneratedLastEpoch, uint(p.generationRateExponent))

		result, _ := safemath.SafeAdd(iotago.Mana(c.Uint64()), iotago.Mana(manaGeneratedLastEpoch.Uint64()))
		result, _ = safemath.SafeSub(result, iotago.Mana(c.Uint64())>>p.decayFactorsExponent)
		result, _ = safemath.SafeSub(result, manaDecayedIntermediateEpochs)
		result, _ = safemath.SafeAdd(result, manaDecayedFirstEpoch)

		return result
	}
}

// ManaGenerationWithDecayFloat calculates mana generation with decay (for potential Mana) using floating point arithmetic.
func (p *TestReferenceManaDecayProvider) ManaGenerationWithDecayFloat(amount iotago.BaseToken, creationSlot iotago.SlotIndex, targetSlot iotago.SlotIndex) float64 {
	creationEpoch := p.timeProvider.EpochFromSlot(creationSlot)
	targetEpoch := p.timeProvider.EpochFromSlot(targetSlot)
	floatAmount := float64(amount)
	floatGenerationRate := float64(p.generationRate) * math.Pow(2, -float64(p.generationRateExponent))
	epochsPerYear := (365.0 * 24.0 * 60.0 * 60.0) / float64(p.timeProvider.EpochDurationSeconds())
	decayPerEpoch := math.Pow(float64(p.annualDecayFactorPercentage)/100.0, 1/epochsPerYear)
	epochDiff := targetEpoch - creationEpoch

	//nolint:exhaustive // false-positive, we have a default case
	switch epochDiff {
	case 0:
		floatSlotDiff := float64(targetSlot - creationSlot)

		return floatSlotDiff * floatAmount * floatGenerationRate

	case 1:
		slotsBeforeNextEpoch := p.timeProvider.SlotsBeforeNextEpoch(creationSlot)
		slotsSinceEpochStart := p.timeProvider.SlotsSinceEpochStart(targetSlot)
		manaDecayed := float64(slotsBeforeNextEpoch) * floatAmount * floatGenerationRate * decayPerEpoch
		manaGenerated := float64(slotsSinceEpochStart) * floatAmount * floatGenerationRate

		return manaDecayed + manaGenerated

	default:
		slotsBeforeNextEpoch := p.timeProvider.SlotsBeforeNextEpoch(creationSlot)
		slotsSinceEpochStart := p.timeProvider.SlotsSinceEpochStart(targetSlot)
		epochDiffFloat := float64(epochDiff)
		slotsPerEpochFloat := math.Pow(2, float64(p.timeProvider.SlotsPerEpochExponent()))
		constant := decayPerEpoch * (1 - math.Pow(decayPerEpoch, epochDiffFloat-1)) / (1 - decayPerEpoch)
		potentialMana_n := float64(slotsBeforeNextEpoch) * floatAmount * floatGenerationRate * math.Pow(decayPerEpoch, epochDiffFloat)
		potentialMana_n_1 := constant * floatAmount * floatGenerationRate * slotsPerEpochFloat
		potentialMana_0 := float64(slotsSinceEpochStart) * floatAmount * floatGenerationRate

		return potentialMana_n + potentialMana_n_1 + potentialMana_0
	}
}

// LowerBoundManaGenerationWithDecay calculates the lower bound for measuring the accuracy of the potential Mana fixed point calculations compared with the floating point result.
func (p *TestReferenceManaDecayProvider) LowerBoundManaGenerationWithDecay(amount iotago.BaseToken, creationSlot iotago.SlotIndex, targetSlot iotago.SlotIndex) float64 {
	epochsPerYear := (365.0 * 24.0 * 60.0 * 60.0) / float64(p.timeProvider.EpochDurationSeconds())
	decayPerEpoch := math.Pow(float64(p.annualDecayFactorPercentage)/100.0, 1/epochsPerYear)
	constant := decayPerEpoch / (1 - decayPerEpoch)

	return p.ManaGenerationWithDecayFloat(amount, creationSlot, targetSlot) - (4 + float64(amount)*float64(p.generationRate)*math.Pow(2, float64(p.timeProvider.SlotsPerEpochExponent()-uint8(p.generationRateExponent)))*(1+constant*math.Pow(2, -float64(p.decayFactorsExponent))))
}

// UpperBoundManaGenerationWithDecay calculates the upper bound for measuring the accuracy of the potential Mana fixed point calculations compared with the floating point result.
func (p *TestReferenceManaDecayProvider) UpperBoundManaGenerationWithDecay(amount iotago.BaseToken, creationSlot iotago.SlotIndex, targetSlot iotago.SlotIndex) float64 {
	return p.ManaGenerationWithDecayFloat(amount, creationSlot, targetSlot) + 2 - math.Pow(2, -float64(p.decayFactorsExponent-1))
}

// LowerBoundManaDecay calculates the lower bound for measuring the accuracy of the Mana decay fixed point calculations compared with the floating point result.
func (p *TestReferenceManaDecayProvider) LowerBoundManaDecay(mana iotago.Mana, creationSlot iotago.SlotIndex, targetSlot iotago.SlotIndex) float64 {
	return p.ManaDecayFloat(mana, creationSlot, targetSlot) - (float64(mana)*math.Pow(2, -float64(p.decayFactorsExponent)) + 1)
}

// UpperBoundManaDecay calculates the upper bound for measuring the accuracy of the Mana decay fixed point calculations compared with the floating point result.
func (p *TestReferenceManaDecayProvider) UpperBoundManaDecay(mana iotago.Mana, creationSlot iotago.SlotIndex, targetSlot iotago.SlotIndex) float64 {
	return p.ManaDecayFloat(mana, creationSlot, targetSlot)
}
