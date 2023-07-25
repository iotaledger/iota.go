package iotago

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/runtime/options"
)

// V3ProtocolParameters defines the parameters of the protocol.
type V3ProtocolParameters struct {
	defaultProtocolParameters `serix:"0"`

	// Derived fields
	livenessThresholdDurationOnce sync.Once
	livenessThresholdDuration     time.Duration
}

func NewV3ProtocolParameters(opts ...options.Option[V3ProtocolParameters]) *V3ProtocolParameters {
	return options.Apply(
		new(V3ProtocolParameters),
		append([]options.Option[V3ProtocolParameters]{
			WithVersion(apiV3Version),
			WithNetworkOptions("testnet", PrefixTestnet),
			WithSupplyOptions(1813620509061365, 100, 1, 10),
			WithWorkScoreOptions(1, 100, 500, 20, 20, 20, 20, 100, 100, 100, 200, 4),
			WithTimeProviderOptions(time.Now().Unix(), 10, 13),
			// TODO: add sane default values
			WithManaOptions(1,
				0,
				[]uint32{
					10,
					20,
				},
				0,
				0,
				0,
			),
			WithLivenessOptions(10, 3, 4),
			WithStakingOptions(10),
			WithVersionSignalingOptions(7, 5, 7),
		},
			opts...,
		),
	)
}

var _ ProtocolParameters = &V3ProtocolParameters{}

func (p *V3ProtocolParameters) Version() Version {
	return p.defaultProtocolParameters.Version
}

func (p *V3ProtocolParameters) Bech32HRP() NetworkPrefix {
	return p.defaultProtocolParameters.Bech32HRP
}

func (p *V3ProtocolParameters) NetworkName() string {
	return p.defaultProtocolParameters.NetworkName
}

func (p *V3ProtocolParameters) RentStructure() *RentStructure {
	return &p.defaultProtocolParameters.RentStructure
}

func (p *V3ProtocolParameters) WorkScoreStructure() *WorkScoreStructure {
	return &p.defaultProtocolParameters.WorkScoreStructure
}

func (p *V3ProtocolParameters) TokenSupply() BaseToken {
	return p.defaultProtocolParameters.TokenSupply
}

func (p *V3ProtocolParameters) NetworkID() NetworkID {
	return NetworkIDFromString(p.defaultProtocolParameters.NetworkName)
}

func (p *V3ProtocolParameters) TimeProvider() *TimeProvider {
	return NewTimeProvider(p.defaultProtocolParameters.GenesisUnixTimestamp, int64(p.defaultProtocolParameters.SlotDurationInSeconds), p.defaultProtocolParameters.SlotsPerEpochExponent)
}

// ParamEpochDurationInSlots defines the amount of slots in an epoch.
func (p *V3ProtocolParameters) ParamEpochDurationInSlots() SlotIndex {
	return 1 << p.defaultProtocolParameters.SlotsPerEpochExponent
}

func (p *V3ProtocolParameters) StakingUnbondingPeriod() EpochIndex {
	return p.defaultProtocolParameters.StakingUnbondingPeriod
}

func (p *V3ProtocolParameters) LivenessThreshold() SlotIndex {
	return p.defaultProtocolParameters.LivenessThreshold
}

func (p *V3ProtocolParameters) LivenessThresholdDuration() time.Duration {
	p.livenessThresholdDurationOnce.Do(func() {
		p.livenessThresholdDuration = time.Duration(uint64(p.defaultProtocolParameters.LivenessThreshold)*uint64(p.defaultProtocolParameters.SlotDurationInSeconds)) * time.Second
	})

	return p.livenessThresholdDuration
}

func (p *V3ProtocolParameters) EvictionAge() SlotIndex {
	return p.defaultProtocolParameters.EvictionAge
}

func (p *V3ProtocolParameters) EpochNearingThreshold() SlotIndex {
	return p.defaultProtocolParameters.EpochNearingThreshold
}

func (p *V3ProtocolParameters) VersionSignaling() *VersionSignaling {
	return &p.defaultProtocolParameters.VersionSignaling
}

func (p *V3ProtocolParameters) Bytes() ([]byte, error) {
	return commonSerixAPI().Encode(context.TODO(), p)
}

func (p *V3ProtocolParameters) Hash() (Identifier, error) {
	bytes, err := p.Bytes()
	if err != nil {
		return Identifier{}, err
	}

	return IdentifierFromData(bytes), nil
}

func (p *V3ProtocolParameters) String() string {
	return fmt.Sprintf("ProtocolParameters: {\n\tVersion: %d\n\tNetwork Name: %s\n\tBech32 HRP Prefix: %s\n\tRent Structure: %v\n\tWorkScore Structure: %v\n\tToken Supply: %d\n\tGenesis Unix Timestamp: %d\n\tSlot Duration in Seconds: %d\n\tSlots per Epoch Exponent: %d\n\tMana Generation Rate: %d\n\tMana Generation Rate Exponent: %d\t\nMana Decay Factors: %v\n\tMana Decay Factors Exponent: %d\n\tMana Decay Factor Epochs Sum: %d\n\tMana Decay Factor Epochs Sum Exponent: %d\n\tStaking Unbonding Period: %d\n\tEviction Age: %d\n\tLiveness Threshold: %d\n}",
		p.defaultProtocolParameters.Version, p.defaultProtocolParameters.NetworkName, p.defaultProtocolParameters.Bech32HRP, p.defaultProtocolParameters.RentStructure, p.defaultProtocolParameters.WorkScoreStructure, p.defaultProtocolParameters.TokenSupply, p.defaultProtocolParameters.GenesisUnixTimestamp, p.defaultProtocolParameters.SlotDurationInSeconds, p.defaultProtocolParameters.SlotsPerEpochExponent, p.defaultProtocolParameters.ManaGenerationRate, p.defaultProtocolParameters.ManaGenerationRateExponent, p.defaultProtocolParameters.ManaDecayFactors, p.defaultProtocolParameters.ManaDecayFactorsExponent, p.defaultProtocolParameters.ManaDecayFactorEpochsSum, p.defaultProtocolParameters.ManaDecayFactorEpochsSumExponent, p.defaultProtocolParameters.StakingUnbondingPeriod, p.defaultProtocolParameters.EvictionAge, p.defaultProtocolParameters.LivenessThreshold)
}

func (p *V3ProtocolParameters) ManaDecayProvider() *ManaDecayProvider {
	return NewManaDecayProvider(p.TimeProvider(), p.defaultProtocolParameters.SlotsPerEpochExponent, p.defaultProtocolParameters.ManaGenerationRate, p.defaultProtocolParameters.ManaGenerationRateExponent, p.defaultProtocolParameters.ManaDecayFactors, p.defaultProtocolParameters.ManaDecayFactorsExponent, p.defaultProtocolParameters.ManaDecayFactorEpochsSum, p.defaultProtocolParameters.ManaDecayFactorEpochsSumExponent)
}

func (p *V3ProtocolParameters) Equals(other *V3ProtocolParameters) bool {
	return p.defaultProtocolParameters.Equals(other.defaultProtocolParameters) &&
		p.LivenessThresholdDuration() == other.LivenessThresholdDuration()
}

func WithVersion(version Version) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.defaultProtocolParameters.Version = version
	}
}

func WithNetworkOptions(networkName string, bech32HRP NetworkPrefix) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.defaultProtocolParameters.NetworkName = networkName
		p.defaultProtocolParameters.Bech32HRP = bech32HRP
	}
}

func WithSupplyOptions(totalSupply BaseToken, vByteCost uint32, vBFactorData VByteCostFactor, vBFactorKey VByteCostFactor) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.defaultProtocolParameters.TokenSupply = totalSupply
		p.defaultProtocolParameters.RentStructure = RentStructure{
			VByteCost:    vByteCost,
			VBFactorData: vBFactorData,
			VBFactorKey:  vBFactorKey,
		}
	}
}

func WithWorkScoreOptions(
	dataByte WorkScore,
	block WorkScore,
	missingParent WorkScore,
	input WorkScore,
	contextInput WorkScore,
	output WorkScore,
	nativeToken WorkScore,
	staking WorkScore,
	blockIssuer WorkScore,
	allotment WorkScore,
	signatureEd25519 WorkScore,
	minStrongParentsThreshold byte) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.defaultProtocolParameters.WorkScoreStructure = WorkScoreStructure{
			DataByte:                  dataByte,
			Block:                     block,
			MissingParent:             missingParent,
			Input:                     input,
			ContextInput:              contextInput,
			Output:                    output,
			NativeToken:               nativeToken,
			Staking:                   staking,
			BlockIssuer:               blockIssuer,
			Allotment:                 allotment,
			SignatureEd25519:          signatureEd25519,
			MinStrongParentsThreshold: minStrongParentsThreshold,
		}
	}
}

func WithTimeProviderOptions(genesisTimestamp int64, slotDuration uint8, slotsPerEpochExponent uint8) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.defaultProtocolParameters.GenesisUnixTimestamp = genesisTimestamp
		p.defaultProtocolParameters.SlotDurationInSeconds = slotDuration
		p.defaultProtocolParameters.SlotsPerEpochExponent = slotsPerEpochExponent
	}
}

func WithManaOptions(manaGenerationRate uint8, manaGenerationRateExponent uint8, manaDecayFactors []uint32, manaDecayFactorsExponent uint8, manaDecayFactorEpochsSum uint32, manaDecayFactorEpochsSumExponent uint8) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.defaultProtocolParameters.ManaGenerationRate = manaGenerationRate
		p.defaultProtocolParameters.ManaGenerationRateExponent = manaGenerationRateExponent
		p.defaultProtocolParameters.ManaDecayFactors = manaDecayFactors
		p.defaultProtocolParameters.ManaDecayFactorsExponent = manaDecayFactorsExponent
		p.defaultProtocolParameters.ManaDecayFactorEpochsSum = manaDecayFactorEpochsSum
		p.defaultProtocolParameters.ManaDecayFactorEpochsSumExponent = manaDecayFactorEpochsSumExponent
	}
}

func WithLivenessOptions(evictionAge SlotIndex, livenessThreshold SlotIndex, epochNearingThreshold SlotIndex) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.defaultProtocolParameters.EvictionAge = evictionAge
		p.defaultProtocolParameters.LivenessThreshold = livenessThreshold
		p.defaultProtocolParameters.EpochNearingThreshold = epochNearingThreshold
	}
}

func WithStakingOptions(unboundPeriod EpochIndex) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.defaultProtocolParameters.StakingUnbondingPeriod = unboundPeriod
	}
}

func WithVersionSignalingOptions(windowSize uint8, windowTargetRatio uint8, activationOffset uint8) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.defaultProtocolParameters.VersionSignaling = VersionSignaling{
			WindowSize:        windowSize,
			WindowTargetRatio: windowTargetRatio,
			ActivationOffset:  activationOffset,
		}
	}
}
