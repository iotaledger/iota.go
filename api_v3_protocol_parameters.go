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
	basicProtocolParameters `serix:""`

	hashIdentifier Identifier
	hashOnce       sync.Once

	bytes      []byte
	bytesMutex sync.Mutex

	networkID     NetworkID
	networkIDOnce sync.Once
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

func (p *V3ProtocolParameters) StorageScoreParameters() *StorageScoreParameters {
	return &p.basicProtocolParameters.StorageScoreParameters
}

func (p *V3ProtocolParameters) WorkScoreParameters() *WorkScoreParameters {
	return &p.basicProtocolParameters.WorkScoreParameters
}

func (p *V3ProtocolParameters) ManaParameters() *ManaParameters {
	return &p.basicProtocolParameters.ManaParameters
}

func (p *V3ProtocolParameters) TokenSupply() BaseToken {
	return p.basicProtocolParameters.TokenSupply
}

func (p *V3ProtocolParameters) NetworkID() NetworkID {
	p.networkIDOnce.Do(func() {
		p.networkID = NetworkIDFromString(p.basicProtocolParameters.NetworkName)
	})

	return p.networkID
}

// GenesisBlockID defines the block ID of the genesis block.
func (p *V3ProtocolParameters) GenesisBlockID() BlockID {
	return NewBlockID(p.basicProtocolParameters.GenesisSlot, EmptyIdentifier)
}

// GenesisSlot defines the genesis slot.
func (p *V3ProtocolParameters) GenesisSlot() SlotIndex {
	return p.basicProtocolParameters.GenesisSlot
}

// GenesisUnixTimestamp defines the genesis timestamp at which the slots start to count.
func (p *V3ProtocolParameters) GenesisUnixTimestamp() int64 {
	return p.basicProtocolParameters.GenesisUnixTimestamp
}

// SlotDurationInSeconds defines the duration of each slot in seconds.
func (p *V3ProtocolParameters) SlotDurationInSeconds() uint8 {
	return p.basicProtocolParameters.SlotDurationInSeconds
}

// SlotsPerEpochExponent is the number of slots in an epoch expressed as an exponent of 2.
func (p *V3ProtocolParameters) SlotsPerEpochExponent() uint8 {
	return p.basicProtocolParameters.SlotsPerEpochExponent
}

// ParamEpochDurationInSlots defines the amount of slots in an epoch.
func (p *V3ProtocolParameters) ParamEpochDurationInSlots() SlotIndex {
	return 1 << p.basicProtocolParameters.SlotsPerEpochExponent
}

func (p *V3ProtocolParameters) StakingUnbondingPeriod() EpochIndex {
	return p.basicProtocolParameters.StakingUnbondingPeriod
}

func (p *V3ProtocolParameters) ValidationBlocksPerSlot() uint8 {
	return p.basicProtocolParameters.ValidationBlocksPerSlot
}

func (p *V3ProtocolParameters) PunishmentEpochs() EpochIndex {
	return p.basicProtocolParameters.PunishmentEpochs
}

func (p *V3ProtocolParameters) LivenessThresholdLowerBound() time.Duration {
	return time.Duration(p.basicProtocolParameters.LivenessThresholdLowerBoundInSeconds) * time.Second
}

func (p *V3ProtocolParameters) LivenessThresholdUpperBound() time.Duration {
	return time.Duration(p.basicProtocolParameters.LivenessThresholdUpperBoundInSeconds) * time.Second
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

func (p *V3ProtocolParameters) CongestionControlParameters() *CongestionControlParameters {
	return &p.basicProtocolParameters.CongestionControlParameters
}

func (p *V3ProtocolParameters) VersionSignalingParameters() *VersionSignalingParameters {
	return &p.basicProtocolParameters.VersionSignalingParameters
}

func (p *V3ProtocolParameters) RewardsParameters() *RewardsParameters {
	return &p.basicProtocolParameters.RewardsParameters
}

func (p *V3ProtocolParameters) TargetCommitteeSize() uint8 {
	return p.basicProtocolParameters.TargetCommitteeSize
}

func (p *V3ProtocolParameters) Bytes() ([]byte, error) {
	if len(p.bytes) > 0 {
		return p.bytes, nil
	}

	p.bytesMutex.Lock()
	defer p.bytesMutex.Unlock()

	// Check if some other goroutine cached the bytes while waiting for the lock.
	if len(p.bytes) > 0 {
		return p.bytes, nil
	}

	bytes, err := CommonSerixAPI().Encode(context.TODO(), p)
	if err != nil {
		return nil, err
	}

	p.bytes = bytes

	return p.bytes, nil
}

func (p *V3ProtocolParameters) Hash() (Identifier, error) {
	bytes, err := p.Bytes()
	if err != nil {
		return Identifier{}, err
	}

	p.hashOnce.Do(func() {
		p.hashIdentifier = IdentifierFromData(bytes)
	})

	return p.hashIdentifier, nil
}

func (p *V3ProtocolParameters) String() string {
	return fmt.Sprintf("ProtocolParameters: {\n\tVersion: %d\n\tNetwork Name: %s\n\tBech32 HRP Prefix: %s\n\tStorageScore Structure: %v\n\tWorkScore Structure: %v\n\tMana Structure: %v\n\tToken Supply: %d\n\tGenesis Slot: %d\n\tGenesis Unix Timestamp: %d\n\tSlot Duration in Seconds: %d\n\tSlots per Epoch Exponent: %d\n\tStaking Unbonding Period: %d\n\tValidation Blocks per Slot: %d\n\tPunishment Epochs: %d\n\tLiveness Threshold Lower Bound: %d\n\tLiveness Threshold Upper Bound: %d\n\tMin Committable Age: %d\n\tMax Committable Age: %d\n\tEpoch Nearing Threshold: %d\n\tCongestion Control parameters: %v\n\tVersion Signaling: %v\n\tRewardsParameters: %v\n",
		p.basicProtocolParameters.Version,
		p.basicProtocolParameters.NetworkName,
		p.basicProtocolParameters.Bech32HRP,
		p.basicProtocolParameters.StorageScoreParameters,
		p.basicProtocolParameters.WorkScoreParameters,
		p.basicProtocolParameters.ManaParameters,
		p.basicProtocolParameters.TokenSupply,
		p.basicProtocolParameters.GenesisSlot,
		p.basicProtocolParameters.GenesisUnixTimestamp,
		p.basicProtocolParameters.SlotDurationInSeconds,
		p.basicProtocolParameters.SlotsPerEpochExponent,
		p.basicProtocolParameters.StakingUnbondingPeriod,
		p.basicProtocolParameters.ValidationBlocksPerSlot,
		p.basicProtocolParameters.PunishmentEpochs,
		p.basicProtocolParameters.LivenessThresholdLowerBoundInSeconds,
		p.basicProtocolParameters.LivenessThresholdUpperBoundInSeconds,
		p.basicProtocolParameters.MinCommittableAge,
		p.basicProtocolParameters.MaxCommittableAge,
		p.basicProtocolParameters.EpochNearingThreshold,
		p.basicProtocolParameters.CongestionControlParameters,
		p.basicProtocolParameters.VersionSignalingParameters,
		p.basicProtocolParameters.RewardsParameters,
	)
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

func WithStorageOptions(storageCost BaseToken, factorData StorageScoreFactor, offsetOutputOverhead, offsetEd25519BlockIssuerKey, offsetStakingFeature, offsetDelegation StorageScore) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.StorageScoreParameters = StorageScoreParameters{
			StorageCost:                 storageCost,
			FactorData:                  factorData,
			OffsetOutputOverhead:        offsetOutputOverhead,
			OffsetEd25519BlockIssuerKey: offsetEd25519BlockIssuerKey,
			OffsetStakingFeature:        offsetStakingFeature,
			OffsetDelegation:            offsetDelegation,
		}
	}
}

func WithWorkScoreOptions(
	dataByte WorkScore,
	block WorkScore,
	input WorkScore,
	contextInput WorkScore,
	output WorkScore,
	nativeToken WorkScore,
	staking WorkScore,
	blockIssuer WorkScore,
	allotment WorkScore,
	signatureEd25519 WorkScore,
) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.WorkScoreParameters = WorkScoreParameters{
			DataByte:         dataByte,
			Block:            block,
			Input:            input,
			ContextInput:     contextInput,
			Output:           output,
			NativeToken:      nativeToken,
			Staking:          staking,
			BlockIssuer:      blockIssuer,
			Allotment:        allotment,
			SignatureEd25519: signatureEd25519,
		}
	}
}

func WithSupplyOptions(baseTokenSupply BaseToken, bitsCount uint8, generationRate uint8, generationRateExponent uint8, decayFactorsExponent uint8, decayFactorEpochsSumExponent uint8, annualDecayFactorPercentage uint8) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.TokenSupply = baseTokenSupply
		p.basicProtocolParameters.ManaParameters.BitsCount = bitsCount
		p.basicProtocolParameters.ManaParameters.GenerationRate = generationRate
		p.basicProtocolParameters.ManaParameters.GenerationRateExponent = generationRateExponent
		p.basicProtocolParameters.ManaParameters.DecayFactorsExponent = decayFactorsExponent
		p.basicProtocolParameters.ManaParameters.DecayFactorEpochsSumExponent = decayFactorEpochsSumExponent
		p.basicProtocolParameters.ManaParameters.AnnualDecayFactorPercentage = annualDecayFactorPercentage
	}
}

func WithTimeOptions(genesisSlot SlotIndex, genesisTimestamp int64, slotDurationInSeconds uint8, slotsPerEpochExponent uint8, livenessThresholdLowerBoundInSeconds uint16, livenessThresholdUpperBoundInSeconds uint16, minCommittableAge SlotIndex, maxCommittableAge SlotIndex, epochNearingThreshold SlotIndex) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.GenesisSlot = genesisSlot
		p.basicProtocolParameters.GenesisUnixTimestamp = genesisTimestamp
		p.basicProtocolParameters.SlotDurationInSeconds = slotDurationInSeconds
		p.basicProtocolParameters.SlotsPerEpochExponent = slotsPerEpochExponent
		p.basicProtocolParameters.LivenessThresholdLowerBoundInSeconds = livenessThresholdLowerBoundInSeconds
		p.basicProtocolParameters.LivenessThresholdUpperBoundInSeconds = livenessThresholdUpperBoundInSeconds
		p.basicProtocolParameters.MinCommittableAge = minCommittableAge
		p.basicProtocolParameters.MaxCommittableAge = maxCommittableAge
		p.basicProtocolParameters.EpochNearingThreshold = epochNearingThreshold
	}
}

func WithCongestionControlOptions(minReferenceManaCost Mana, rmcIncrease Mana, rmcDecrease Mana, rmcIncreaseThreshold WorkScore, rmcDecreaseThreshold WorkScore, schedulerRate WorkScore, maxBufferSize uint32, maxValBufferSize uint32) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.CongestionControlParameters.MinReferenceManaCost = minReferenceManaCost
		p.basicProtocolParameters.CongestionControlParameters.Increase = rmcIncrease
		p.basicProtocolParameters.CongestionControlParameters.Decrease = rmcDecrease
		p.basicProtocolParameters.CongestionControlParameters.IncreaseThreshold = rmcIncreaseThreshold
		p.basicProtocolParameters.CongestionControlParameters.DecreaseThreshold = rmcDecreaseThreshold
		p.basicProtocolParameters.CongestionControlParameters.SchedulerRate = schedulerRate
		p.basicProtocolParameters.CongestionControlParameters.MaxBufferSize = maxBufferSize
		p.basicProtocolParameters.CongestionControlParameters.MaxValidationBufferSize = maxValBufferSize
	}
}

func WithStakingOptions(unbondingPeriod EpochIndex, validationBlocksPerSlot uint8, punishmentEpochs EpochIndex) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.StakingUnbondingPeriod = unbondingPeriod
		p.basicProtocolParameters.ValidationBlocksPerSlot = validationBlocksPerSlot
		p.basicProtocolParameters.PunishmentEpochs = punishmentEpochs
	}
}

func WithVersionSignalingOptions(windowSize uint8, windowTargetRatio uint8, activationOffset uint8) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.VersionSignalingParameters = VersionSignalingParameters{
			WindowSize:        windowSize,
			WindowTargetRatio: windowTargetRatio,
			ActivationOffset:  activationOffset,
		}
	}
}

func WithRewardsOptions(profitMarginExponent, decayBalancingConstantExponent, poolCoefficientExponent uint8, manaShareCoefficient, decayBalancingConstant uint64) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.RewardsParameters.ProfitMarginExponent = profitMarginExponent
		p.basicProtocolParameters.RewardsParameters.ManaShareCoefficient = manaShareCoefficient
		p.basicProtocolParameters.RewardsParameters.DecayBalancingConstantExponent = decayBalancingConstantExponent
		p.basicProtocolParameters.RewardsParameters.DecayBalancingConstant = decayBalancingConstant
		p.basicProtocolParameters.RewardsParameters.PoolCoefficientExponent = poolCoefficientExponent
	}
}

func WithTargetCommitteeSize(targetCommitteeSize uint8) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.TargetCommitteeSize = targetCommitteeSize
	}
}
