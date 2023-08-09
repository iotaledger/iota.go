package iotago

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/runtime/options"
)

// V3ProtocolParameters defines the parameters of the protocol.
type V3ProtocolParameters struct {
	basicProtocolParameters `serix:"0"`
}

func NewV3ProtocolParameters(opts ...options.Option[V3ProtocolParameters]) *V3ProtocolParameters {
	return options.Apply(
		new(V3ProtocolParameters),
		append([]options.Option[V3ProtocolParameters]{
			WithVersion(apiV3Version),
			WithNetworkOptions("testnet", PrefixTestnet),
			WithSupplyOptions(1813620509061365, 100, 1, 10, 100, 100),
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
			WithLivenessOptions(3, 10, 20, 24),
			WithStakingOptions(10),
			WithVersionSignalingOptions(7, 5, 7),
		},
			opts...,
		),
	)
}

var _ ProtocolParameters = &V3ProtocolParameters{}

func (p *V3ProtocolParameters) Version() Version {
	return p.basicProtocolParameters.Version
}

func (p *V3ProtocolParameters) Bech32HRP() NetworkPrefix {
	return p.basicProtocolParameters.Bech32HRP
}

func (p *V3ProtocolParameters) NetworkName() string {
	return p.basicProtocolParameters.NetworkName
}

func (p *V3ProtocolParameters) RentStructure() *RentStructure {
	return &p.basicProtocolParameters.RentStructure
}

func (p *V3ProtocolParameters) WorkScoreStructure() *WorkScoreStructure {
	return &p.basicProtocolParameters.WorkScoreStructure
}

func (p *V3ProtocolParameters) TokenSupply() BaseToken {
	return p.basicProtocolParameters.TokenSupply
}

func (p *V3ProtocolParameters) NetworkID() NetworkID {
	return NetworkIDFromString(p.basicProtocolParameters.NetworkName)
}

func (p *V3ProtocolParameters) TimeProvider() *TimeProvider {
	return NewTimeProvider(p.basicProtocolParameters.GenesisUnixTimestamp, int64(p.basicProtocolParameters.SlotDurationInSeconds), p.basicProtocolParameters.SlotsPerEpochExponent)
}

// ParamEpochDurationInSlots defines the amount of slots in an epoch.
func (p *V3ProtocolParameters) ParamEpochDurationInSlots() SlotIndex {
	return 1 << p.basicProtocolParameters.SlotsPerEpochExponent
}

func (p *V3ProtocolParameters) StakingUnbondingPeriod() EpochIndex {
	return p.basicProtocolParameters.StakingUnbondingPeriod
}

func (p *V3ProtocolParameters) LivenessThreshold() SlotIndex {
	return p.basicProtocolParameters.LivenessThreshold
}

func (p *V3ProtocolParameters) MinCommittableAge() SlotIndex {
	return p.basicProtocolParameters.MinCommittableAge
}

func (p *V3ProtocolParameters) MaxCommittableAge() SlotIndex {
	return p.basicProtocolParameters.MaxCommittableAge
}

func (p *V3ProtocolParameters) EpochNearingThreshold() SlotIndex {
	return p.basicProtocolParameters.EpochNearingThreshold
}

func (p *V3ProtocolParameters) VersionSignaling() *VersionSignaling {
	return &p.basicProtocolParameters.VersionSignaling
}

func (p *V3ProtocolParameters) Bytes() ([]byte, error) {
	return CommonSerixAPI().Encode(context.TODO(), p)
}

func (p *V3ProtocolParameters) Hash() (Identifier, error) {
	bytes, err := p.Bytes()
	if err != nil {
		return Identifier{}, err
	}

	return IdentifierFromData(bytes), nil
}

func (p *V3ProtocolParameters) String() string {
	return fmt.Sprintf("ProtocolParameters: {\n\tVersion: %d\n\tNetwork Name: %s\n\tBech32 HRP Prefix: %s\n\tRent Structure: %v\n\tWorkScore Structure: %v\n\tToken Supply: %d\n\tGenesis Unix Timestamp: %d\n\tSlot Duration in Seconds: %d\n\tSlots per Epoch Exponent: %d\n\tMana Generation Rate: %d\n\tMana Generation Rate Exponent: %d\t\nMana Decay Factors: %v\n\tMana Decay Factors Exponent: %d\n\tMana Decay Factor Epochs Sum: %d\n\tMana Decay Factor Epochs Sum Exponent: %d\n\tStaking Unbonding Period: %d\n\tLiveness Threshold: %d\n\tMin Committable Age: %d\n\tMax Committable Age: %d\n}",
		p.basicProtocolParameters.Version, p.basicProtocolParameters.NetworkName, p.basicProtocolParameters.Bech32HRP, p.basicProtocolParameters.RentStructure, p.basicProtocolParameters.WorkScoreStructure, p.basicProtocolParameters.TokenSupply, p.basicProtocolParameters.GenesisUnixTimestamp, p.basicProtocolParameters.SlotDurationInSeconds, p.basicProtocolParameters.SlotsPerEpochExponent, p.basicProtocolParameters.ManaGenerationRate, p.basicProtocolParameters.ManaGenerationRateExponent, p.basicProtocolParameters.ManaDecayFactors, p.basicProtocolParameters.ManaDecayFactorsExponent, p.basicProtocolParameters.ManaDecayFactorEpochsSum, p.basicProtocolParameters.ManaDecayFactorEpochsSumExponent, p.basicProtocolParameters.StakingUnbondingPeriod, p.basicProtocolParameters.LivenessThreshold, p.basicProtocolParameters.MinCommittableAge, p.basicProtocolParameters.MaxCommittableAge)
}

func (p *V3ProtocolParameters) ManaDecayProvider() *ManaDecayProvider {
	return NewManaDecayProvider(p.TimeProvider(), p.basicProtocolParameters.SlotsPerEpochExponent, p.basicProtocolParameters.ManaGenerationRate, p.basicProtocolParameters.ManaGenerationRateExponent, p.basicProtocolParameters.ManaDecayFactors, p.basicProtocolParameters.ManaDecayFactorsExponent, p.basicProtocolParameters.ManaDecayFactorEpochsSum, p.basicProtocolParameters.ManaDecayFactorEpochsSumExponent)
}

func (p *V3ProtocolParameters) Equals(other ProtocolParameters) bool {
	otherV3Params, matches := other.(*V3ProtocolParameters)
	if !matches {
		return false
	}

	return p.basicProtocolParameters.Equals(otherV3Params.basicProtocolParameters)
}

func WithVersion(version Version) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.Version = version
	}
}

func WithNetworkOptions(networkName string, bech32HRP NetworkPrefix) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.NetworkName = networkName
		p.basicProtocolParameters.Bech32HRP = bech32HRP
	}
}

func WithSupplyOptions(totalSupply BaseToken, vByteCost uint32, vBFactorData, vBFactorKey, vBFactorIssuerKeys, vBFactorStakingFeature VByteCostFactor) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.TokenSupply = totalSupply
		p.basicProtocolParameters.RentStructure = RentStructure{
			VByteCost:              vByteCost,
			VBFactorData:           vBFactorData,
			VBFactorKey:            vBFactorKey,
			VBFactorIssuerKeys:     vBFactorIssuerKeys,
			VBFactorStakingFeature: vBFactorStakingFeature,
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
		p.basicProtocolParameters.WorkScoreStructure = WorkScoreStructure{
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
		p.basicProtocolParameters.GenesisUnixTimestamp = genesisTimestamp
		p.basicProtocolParameters.SlotDurationInSeconds = slotDuration
		p.basicProtocolParameters.SlotsPerEpochExponent = slotsPerEpochExponent
	}
}

func WithManaOptions(manaGenerationRate uint8, manaGenerationRateExponent uint8, manaDecayFactors []uint32, manaDecayFactorsExponent uint8, manaDecayFactorEpochsSum uint32, manaDecayFactorEpochsSumExponent uint8) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.ManaGenerationRate = manaGenerationRate
		p.basicProtocolParameters.ManaGenerationRateExponent = manaGenerationRateExponent
		p.basicProtocolParameters.ManaDecayFactors = manaDecayFactors
		p.basicProtocolParameters.ManaDecayFactorsExponent = manaDecayFactorsExponent
		p.basicProtocolParameters.ManaDecayFactorEpochsSum = manaDecayFactorEpochsSum
		p.basicProtocolParameters.ManaDecayFactorEpochsSumExponent = manaDecayFactorEpochsSumExponent
	}
}

func WithLivenessOptions(livenessThreshold SlotIndex, minCommittableAge SlotIndex, maxCommittableAge SlotIndex, epochNearingThreshold SlotIndex) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.LivenessThreshold = livenessThreshold
		p.basicProtocolParameters.MinCommittableAge = minCommittableAge
		p.basicProtocolParameters.MaxCommittableAge = maxCommittableAge
		p.basicProtocolParameters.EpochNearingThreshold = epochNearingThreshold
	}
}

func WithStakingOptions(unbondingPeriod EpochIndex) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.StakingUnbondingPeriod = unbondingPeriod
	}
}

func WithVersionSignalingOptions(windowSize uint8, windowTargetRatio uint8, activationOffset uint8) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.VersionSignaling = VersionSignaling{
			WindowSize:        windowSize,
			WindowTargetRatio: windowTargetRatio,
			ActivationOffset:  activationOffset,
		}
	}
}
