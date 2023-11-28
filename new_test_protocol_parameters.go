package iotago

import (
	"fmt"
	"math"
	"time"

	"github.com/iotaledger/hive.go/core/safemath"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/runtime/options"
)

// NewV3TestProtocolParameters creates a new V3ProtocolParameters instance with the given options.
// NB: this function should only be used to derive new protocol params for tests because it uses floating point arithmetic
// to derive mana decay factors and decay factor epochs sum. This will result in different parameters on different machines.
func NewV3TestProtocolParameters(opts ...options.Option[V3ProtocolParameters]) *V3ProtocolParameters {
	newProtocolParams := options.Apply(
		new(V3ProtocolParameters),
		append([]options.Option[V3ProtocolParameters]{
			WithVersion(apiV3Version),
			WithStorageOptions(100, 1, 10, 100, 100, 100),
			WithWorkScoreOptions(0, 1, 0, 0, 0, 0, 0, 0, 0, 0),
			WithTimeOptions(0, time.Now().Unix(), 10, 13, 15, 30, 10, 20, 60),
			WithSupplyOptions(1813620509061365, 63, 1, 17, 32, 21, 70),
			WithCongestionControlOptions(1, 0, 0, 800_000, 500_000, 100_000, 1000, 100),
			WithStakingOptions(10, 10, 10),
			WithVersionSignalingOptions(7, 5, 7),
			WithRewardsOptions(8, 8, 31, 1080, 2, 1),
			WithTargetCommitteeSize(32),
		},
			opts...,
		),
	)

	// fix the network options to be testnet.
	// Do not use this function outside of tests.
	newProtocolParams.basicProtocolParameters.NetworkName = "testnet"
	newProtocolParams.basicProtocolParameters.Bech32HRP = PrefixTestnet

	// Compute derived parameters
	newProtocolParams.basicProtocolParameters.ManaParameters.DecayFactors = deriveManaDecayFactors(
		newProtocolParams.ManaParameters().AnnualDecayFactorPercentage,
		newProtocolParams.SlotsPerEpochExponent(),
		newProtocolParams.SlotDurationInSeconds(),
		newProtocolParams.ManaParameters().DecayFactorsExponent,
	)
	newProtocolParams.basicProtocolParameters.ManaParameters.DecayFactorEpochsSum = deriveManaDecayFactorEpochsSum(
		newProtocolParams.ManaParameters().AnnualDecayFactorPercentage,
		newProtocolParams.SlotsPerEpochExponent(),
		newProtocolParams.SlotDurationInSeconds(),
		newProtocolParams.ManaParameters().DecayFactorEpochsSumExponent,
	)

	// Sanity checks
	manaSupplySanityCheck(newProtocolParams)
	timeSanityCheck(newProtocolParams)
	congestionControlSanityCheck(newProtocolParams)

	return newProtocolParams
}

// deriveManaDecayFactors computes a lookup table of mana decay factors using floating point arithmetic.
func deriveManaDecayFactors(annualDecayFactorPercentage uint8, slotsPerEpochExponent uint8, slotDurationSeconds uint8, decayFactorsExponent uint8) []uint32 {
	epochsPerYear := ((365.0 * 24.0 * 60.0 * 60.0) / float64(slotDurationSeconds)) / math.Pow(2, float64(slotsPerEpochExponent))
	decayFactors := make([]uint32, int(epochsPerYear))

	decayFactorPerEpoch := math.Pow(float64(annualDecayFactorPercentage)/100.0, 1.0/epochsPerYear)

	for epoch := 1; epoch <= int(epochsPerYear); epoch++ {
		decayFactor := math.Pow(decayFactorPerEpoch, float64(epoch)) * (math.Pow(2, float64(decayFactorsExponent)))
		decayFactors[epoch-1] = uint32(decayFactor)
	}

	return decayFactors
}

// deriveManaDecayFactorEpochsSum computes mana decay factor epochs sum parameter using floating point arithmetic.
func deriveManaDecayFactorEpochsSum(annualDecayFactorPercentage uint8, slotsPerEpochExponent uint8, slotDurationSeconds uint8, decayFactorEpochsSumExponent uint8) uint32 {
	delta := math.Pow(2, float64(slotsPerEpochExponent)) * (1.0 / (365.0 * 24.0 * 60.0 * 60.0)) * float64(slotDurationSeconds)
	annualDecayFactor := float64(annualDecayFactorPercentage) / 100.0

	return uint32((math.Pow(annualDecayFactor, delta) / (1 - math.Pow(annualDecayFactor, delta)) * (math.Pow(2, float64(decayFactorEpochsSumExponent)))))
}

func manaSupplySanityCheck(protocolParams *V3ProtocolParameters) {
	beta := -math.Log(float64(protocolParams.ManaParameters().AnnualDecayFactorPercentage))
	epochDurationInYears := float64(protocolParams.SlotDurationInSeconds()) * math.Pow(2.0, float64(protocolParams.SlotsPerEpochExponent())) / (365 * 24 * 60 * 60)
	maxManaSupply := 21.0 * float64(protocolParams.TokenSupply()) * float64(protocolParams.ManaParameters().GenerationRate) * math.Pow(2.0, float64(protocolParams.SlotsPerEpochExponent())-float64(protocolParams.ManaParameters().GenerationRateExponent)) / (beta * epochDurationInYears)
	if maxManaSupply >= math.Pow(2.0, float64(protocolParams.ManaParameters().BitsCount)) {
		panic("the combination of parameters might lead to overflowing of the Mana supply")
	}
	// this check is specific to the way decay is calculated to prevent overflow
	if _, err := safemath.SafeMul(protocolParams.ManaParameters().DecayFactorEpochsSum, uint32(protocolParams.ManaParameters().GenerationRate)); err != nil {
		panic("decayFactorEpochsSum * generationRate must not require more than 32 bits")
	}
}

func timeSanityCheck(protocolParams *V3ProtocolParameters) {
	if protocolParams.LivenessThresholdLowerBoundInSeconds > protocolParams.LivenessThresholdUpperBoundInSeconds {
		panic("LivenessThresholdLowerBoundInSeconds must be less than or equal to LivenessThresholdUpperBoundInSeconds")
	}
	if SlotIndex(protocolParams.LivenessThresholdUpperBoundInSeconds) >= protocolParams.MinCommittableAge()*SlotIndex(protocolParams.SlotDurationInSeconds()) {
		panic("LivenessThresholdUpperBoundInSeconds * SlotDurationInSeconds must be less than MinCommittableAge")
	}
	if protocolParams.MinCommittableAge() >= protocolParams.MaxCommittableAge() {
		panic("MinCommittableAge must be strictly less than MaxCommittableAge")
	}
	if lo.PanicOnErr(safemath.SafeMul(2, protocolParams.MaxCommittableAge())) > protocolParams.EpochNearingThreshold() {
		panic("EpochNearingThreshold must be at least 2 times MaxCommittableAge")
	}
	if (1 << protocolParams.SlotsPerEpochExponent()) < 2*protocolParams.EpochNearingThreshold() {
		panic("Epoch duration in slots must be at least 2 times EpochNearingThreshold")
	}
}

func congestionControlSanityCheck(protocolParams *V3ProtocolParameters) {
	if protocolParams.CongestionControlParameters().IncreaseThreshold > WorkScore(protocolParams.SlotDurationInSeconds())*protocolParams.CongestionControlParameters().SchedulerRate {
		fmt.Printf("IncreaseThreshold: %d, SchedulerRate: %d, SlotDurationSeconds: %d\n", protocolParams.CongestionControlParameters().IncreaseThreshold, protocolParams.CongestionControlParameters().SchedulerRate, protocolParams.SlotDurationInSeconds())
		panic("IncreaseThreshold must be less than or equal to SchedulerRate*SlotDurationInSeconds")
	}
	if protocolParams.CongestionControlParameters().DecreaseThreshold > WorkScore(protocolParams.SlotDurationInSeconds())*protocolParams.CongestionControlParameters().SchedulerRate {
		panic("DecreaseThreshold must be less than or equal to SchedulerRate*SlotDurationInSeconds")
	}
	if protocolParams.CongestionControlParameters().DecreaseThreshold > protocolParams.CongestionControlParameters().IncreaseThreshold {
		panic("DecreaseThreshold must be less than or equal to IncreaseThreshold")
	}
}
