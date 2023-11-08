package iotago

type basicProtocolParameters struct {
	// Version defines the version of the protocol this protocol parameters are for.
	Version Version `serix:"0,mapKey=version"`

	// NetworkName defines the human friendly name of the network.
	NetworkName string `serix:"1,lengthPrefixType=uint8,mapKey=networkName"`
	// Bech32HRP defines the HRP prefix used for Bech32 addresses in the network.
	Bech32HRP NetworkPrefix `serix:"2,lengthPrefixType=uint8,mapKey=bech32Hrp"`

	// StorageScoreParameters defines the Storage Score Parameters used by the given network.
	StorageScoreParameters StorageScoreParameters `serix:"3,mapKey=storageScoreParameters"`
	// WorkScoreParameters defines the work score parameters used by by the given network.
	WorkScoreParameters WorkScoreParameters `serix:"4,mapKey=workScoreParameters"`
	// ManaParameters defines the mana parameters used by the given network.
	ManaParameters ManaParameters `serix:"5,mapKey=manaParameters"`

	// TokenSupply defines the current token supply on the network.
	TokenSupply BaseToken `serix:"6,mapKey=tokenSupply"`

	// GenesisUnixTimestamp defines the genesis timestamp at which the slots start to count.
	GenesisUnixTimestamp int64 `serix:"7,mapKey=genesisUnixTimestamp"`
	// SlotDurationInSeconds defines the duration of each slot in seconds.
	SlotDurationInSeconds uint8 `serix:"8,mapKey=slotDurationInSeconds"`
	// SlotsPerEpochExponent is the number of slots in an epoch expressed as an exponent of 2.
	// (2**SlotsPerEpochExponent) == slots in an epoch.
	SlotsPerEpochExponent uint8 `serix:"9,mapKey=slotsPerEpochExponent"`

	// StakingUnbondingPeriod defines the unbonding period in epochs before an account can stop staking.
	StakingUnbondingPeriod EpochIndex `serix:"10,mapKey=stakingUnbondingPeriod"`
	// ValidationBlocksPerSlot is the number of validation blocks that each validator should issue each slot.
	ValidationBlocksPerSlot uint8 `serix:"11,mapKey=validationBlocksPerSlot"`
	// PunishmentEpochs is the number of epochs worth of Mana that a node is punished with for each additional validation block it issues.
	PunishmentEpochs EpochIndex `serix:"12,mapKey=punishmentEpochs"`

	// LivenessThresholdLowerBound is used by tip-selection to determine if a block is eligible by evaluating issuingTimes.
	// and commitments in its past-cone to ATT and lastCommittedSlot respectively.
	LivenessThresholdLowerBoundInSeconds uint16 `serix:"13,mapKey=livenessThresholdLowerBound"`
	// LivenessThresholdUpperBound is used by tip-selection to determine if a block is eligible by evaluating issuingTimes
	// and commitments in its past-cone to ATT and lastCommittedSlot respectively.
	LivenessThresholdUpperBoundInSeconds uint16 `serix:"14,mapKey=livenessThresholdUpperBound"`

	// MinCommittableAge is the minimum age relative to the accepted tangle time slot index that a slot can be committed.
	// For example, if the last accepted slot is in slot 100, and minCommittableAge=10, then the latest committed slot can be at most 100-10=90.
	MinCommittableAge SlotIndex `serix:"15,mapKey=minCommittableAge"`
	// MaxCommittableAge is the maximum age for a slot commitment to be included in a block relative to the slot index of the block issuing time.
	// For example, if the last accepted slot is in slot 100, and maxCommittableAge=20, then the oldest referencable commitment is 100-20=80.
	MaxCommittableAge SlotIndex `serix:"16,mapKey=maxCommittableAge"`
	// EpochNearingThreshold is used by the epoch orchestrator to detect the slot that should trigger a new committee
	// selection for the next and upcoming epoch.
	EpochNearingThreshold SlotIndex `serix:"17,mapKey=epochNearingThreshold"`
	// RMCParameters defines the parameters used by to calculate the Reference Mana Cost (RMC).
	CongestionControlParameters CongestionControlParameters `serix:"18,mapKey=congestionControlParameters"`
	// VersionSignaling defines the parameters used for version upgrades.
	VersionSignaling VersionSignaling `serix:"19,mapKey=versionSignaling"`
	// RewardsParameters defines the parameters used for reward calculation.
	RewardsParameters RewardsParameters `serix:"20,mapKey=rewardsParameters"`

	// TargetCommitteeSize defines the target size of the committee. If there's fewer candidates the actual committee size could be smaller in a given epoch.
	TargetCommitteeSize uint8 `serix:"21,mapKey=targetCommitteeSize"`
}

func (b basicProtocolParameters) Equals(other basicProtocolParameters) bool {
	return b.Version == other.Version &&
		b.NetworkName == other.NetworkName &&
		b.Bech32HRP == other.Bech32HRP &&
		b.StorageScoreParameters.Equals(other.StorageScoreParameters) &&
		b.WorkScoreParameters.Equals(other.WorkScoreParameters) &&
		b.ManaParameters.Equals(other.ManaParameters) &&
		b.TokenSupply == other.TokenSupply &&
		b.GenesisUnixTimestamp == other.GenesisUnixTimestamp &&
		b.SlotDurationInSeconds == other.SlotDurationInSeconds &&
		b.SlotsPerEpochExponent == other.SlotsPerEpochExponent &&
		b.StakingUnbondingPeriod == other.StakingUnbondingPeriod &&
		b.ValidationBlocksPerSlot == other.ValidationBlocksPerSlot &&
		b.PunishmentEpochs == other.PunishmentEpochs &&
		b.LivenessThresholdLowerBoundInSeconds == other.LivenessThresholdLowerBoundInSeconds &&
		b.LivenessThresholdUpperBoundInSeconds == other.LivenessThresholdUpperBoundInSeconds &&
		b.MinCommittableAge == other.MinCommittableAge &&
		b.MaxCommittableAge == other.MaxCommittableAge &&
		b.EpochNearingThreshold == other.EpochNearingThreshold &&
		b.CongestionControlParameters.Equals(other.CongestionControlParameters) &&
		b.VersionSignaling.Equals(other.VersionSignaling) &&
		b.RewardsParameters.Equals(other.RewardsParameters) &&
		b.TargetCommitteeSize == other.TargetCommitteeSize
}
