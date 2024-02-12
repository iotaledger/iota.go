package iotago

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/core/safemath"
	"github.com/iotaledger/hive.go/lo"
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

func (p *V3ProtocolParameters) ChainSwitchingThreshold() uint8 {
	return p.basicProtocolParameters.ChainSwitchingThreshold
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

// NewV3SnapshotProtocolParameters creates a new V3ProtocolParameters instance with the given options.
// IMPORTANT: this function should only be used to derive new protocol params for genesis snapshots or tests because it uses
// floating point arithmetic to derive mana decay factors and decay factor epochs sum.
// This might result in different parameters on different machines.
func NewV3SnapshotProtocolParameters(opts ...options.Option[V3ProtocolParameters]) *V3ProtocolParameters {
	newProtocolParams := options.Apply(
		new(V3ProtocolParameters),
		append([]options.Option[V3ProtocolParameters]{
			WithVersion(apiV3Version),
			WithNetworkOptions("testnet", PrefixTestnet),
			WithStorageOptions(100, 1, 10, 100, 100, 100),
			WithWorkScoreOptions(500, 110_000, 7_500, 40_000, 90_000, 50_000, 40_000, 70_000, 5_000, 15_000),
			WithTimeProviderOptions(0, time.Now().Unix(), 10, 13),
			WithLivenessOptions(15, 30, 10, 20, 60),
			WithSupplyOptions(1813620509061365, 63, 1, 17, 32, 21, 70),
			WithCongestionControlOptions(1, 10, 10, 400_000_000, 250_000_000, 50_000_000, 1000, 100),
			WithStakingOptions(10, 10, 10),
			WithVersionSignalingOptions(7, 5, 7),
			WithRewardsOptions(8, 8, 11, 2, 1, 384),
			WithTargetCommitteeSize(32),
			WithChainSwitchingThreshold(3),
		},
			opts...,
		),
	)

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
	newProtocolParams.basicProtocolParameters.RewardsParameters.BootstrappingDuration = deriveBootstrappingDuration(
		newProtocolParams.ManaParameters().AnnualDecayFactorPercentage,
		newProtocolParams.SlotsPerEpochExponent(),
		newProtocolParams.SlotDurationInSeconds(),
	)

	// Sanity checks
	manaSupplySanityCheck(newProtocolParams)
	timeSanityCheck(newProtocolParams)
	congestionControlSanityCheck(newProtocolParams)
	stakingSanityCheck(newProtocolParams)

	return newProtocolParams
}

// deriveManaDecayFactors computes a lookup table of mana decay factors using floating point arithmetic.
func deriveManaDecayFactors(annualDecayFactorPercentage uint8, slotsPerEpochExponent uint8, slotDurationSeconds uint8, decayFactorsExponent uint8) []uint32 {
	epochsPerYear := ((365.0 * 24.0 * 60.0 * 60.0) / float64(slotDurationSeconds)) / math.Pow(2, float64(slotsPerEpochExponent))
	epochsInTable := lo.Min(65535, int(epochsPerYear))
	decayFactors := make([]uint32, epochsInTable)

	decayFactorPerEpoch := math.Pow(float64(annualDecayFactorPercentage)/100.0, 1.0/epochsPerYear)

	for epoch := 1; epoch <= epochsInTable; epoch++ {
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

// deriveBootstrappingDuration computes the bootstrapping duration using floating point arithmetic.
func deriveBootstrappingDuration(annualDecayFactorPercentage uint8, slotsPerEpochExponent uint8, slotDurationSeconds uint8) EpochIndex {
	epochsPerYear := (365.0 * 24.0 * 60.0 * 60.0) / (math.Pow(2, float64(slotsPerEpochExponent)) * float64(slotDurationSeconds))
	annualDecayFactor := float64(annualDecayFactorPercentage) / 100.0
	beta := -math.Log(annualDecayFactor)

	return EpochIndex(epochsPerYear / beta)
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
		panic("LivenessThresholdUpperBoundInSeconds must be strictly less than MinCommittableAge * SlotDurationInSeconds")
	}
	if protocolParams.MinCommittableAge() >= protocolParams.MaxCommittableAge() {
		panic("MinCommittableAge must be strictly less than MaxCommittableAge")
	}
	if protocolParams.MaxCommittableAge() >= protocolParams.EpochNearingThreshold() {
		panic("MaxCommittableAge must be strictly less than EpochNearingThreshold")
	}
	if (1 << protocolParams.SlotsPerEpochExponent()) <= protocolParams.EpochNearingThreshold() {
		panic("Epoch duration in slots must be strictly greater than EpochNearingThreshold")
	}
	// TODO: add warning level log for EpochNearingThreshold > 2 * MaxCommittableAge and EpochsPerSlot > 2 * EpochNearingThreshold
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

func stakingSanityCheck(protocolParams *V3ProtocolParameters) {
	tokenSupplyBitsCount := uint8(math.Log2(float64(protocolParams.TokenSupply()))) + 1
	poolCoefficientExponent := protocolParams.RewardsParameters().PoolCoefficientExponent
	if poolCoefficientExponent > 64 || tokenSupplyBitsCount+poolCoefficientExponent > 64 {
		message := fmt.Sprintf("Token supply bits count (%d) + PoolCoefficientExponent (%d) must be less than or equal to 64\n", tokenSupplyBitsCount, protocolParams.RewardsParameters().PoolCoefficientExponent)
		panic(message)
	}

	if protocolParams.ValidationBlocksPerSlot() > 32 {
		panic("ValidationBlocksPerSlot must be less than or equal to 32")
	}
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

func WithTimeProviderOptions(genesisSlot SlotIndex, genesisTimestamp int64, slotDurationInSeconds uint8, slotsPerEpochExponent uint8) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.GenesisSlot = genesisSlot
		p.basicProtocolParameters.GenesisUnixTimestamp = genesisTimestamp
		p.basicProtocolParameters.SlotDurationInSeconds = slotDurationInSeconds
		p.basicProtocolParameters.SlotsPerEpochExponent = slotsPerEpochExponent
	}
}

func WithLivenessOptions(livenessThresholdLowerBoundInSeconds uint16, livenessThresholdUpperBoundInSeconds uint16, minCommittableAge SlotIndex, maxCommittableAge SlotIndex, epochNearingThreshold SlotIndex) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
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

func WithRewardsOptions(profitMarginExponent, decayBalancingConstantExponent, poolCoefficientExponent uint8, manaShareCoefficient, decayBalancingConstant uint64, retentionPeriod uint16) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.RewardsParameters.ProfitMarginExponent = profitMarginExponent
		p.basicProtocolParameters.RewardsParameters.ManaShareCoefficient = manaShareCoefficient
		p.basicProtocolParameters.RewardsParameters.DecayBalancingConstantExponent = decayBalancingConstantExponent
		p.basicProtocolParameters.RewardsParameters.DecayBalancingConstant = decayBalancingConstant
		p.basicProtocolParameters.RewardsParameters.PoolCoefficientExponent = poolCoefficientExponent
		p.basicProtocolParameters.RewardsParameters.RetentionPeriod = retentionPeriod
	}
}

func WithTargetCommitteeSize(targetCommitteeSize uint8) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.TargetCommitteeSize = targetCommitteeSize
	}
}

func WithNetworkOptions(networkName string, hrp NetworkPrefix) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.NetworkName = networkName
		p.basicProtocolParameters.Bech32HRP = hrp
	}
}

func WithChainSwitchingThreshold(chainSwitchingThreshold uint8) options.Option[V3ProtocolParameters] {
	return func(p *V3ProtocolParameters) {
		p.basicProtocolParameters.ChainSwitchingThreshold = chainSwitchingThreshold
	}
}
