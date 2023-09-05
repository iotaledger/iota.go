package iotago

type basicProtocolParameters struct {
	// Version defines the version of the protocol this protocol parameters are for.
	Version Version `serix:"0,mapKey=version"`

	// NetworkName defines the human friendly name of the network.
	NetworkName string `serix:"1,lengthPrefixType=uint8,mapKey=networkName"`
	// Bech32HRP defines the HRP prefix used for Bech32 addresses in the network.
	Bech32HRP NetworkPrefix `serix:"2,lengthPrefixType=uint8,mapKey=bech32Hrp"`

	// RentStructure defines the rent structure used by given node/network.
	RentStructure RentStructure `serix:"3,mapKey=rentStructure"`
	// WorkScoreStructure defines the work score structure used by given node/network.
	WorkScoreStructure WorkScoreStructure `serix:"4,mapKey=workScoreStructure"`
	// TokenSupply defines the current token supply on the network.
	TokenSupply BaseToken `serix:"5,mapKey=tokenSupply"`

	// GenesisUnixTimestamp defines the genesis timestamp at which the slots start to count.
	GenesisUnixTimestamp int64 `serix:"6,mapKey=genesisUnixTimestamp"`
	// SlotDurationInSeconds defines the duration of each slot in seconds.
	SlotDurationInSeconds uint8 `serix:"7,mapKey=slotDurationInSeconds"`
	// SlotsPerEpochExponent is the number of slots in an epoch expressed as an exponent of 2.
	// (2**SlotsPerEpochExponent) == slots in an epoch.
	SlotsPerEpochExponent uint8 `serix:"8,mapKey=slotsPerEpochExponent"`

	// ManaStructure defines the mana parameters used by mana calculation.
	ManaStructure ManaStructure `serix:"9,mapKey=manaStructure"`

	// StakingUnbondingPeriod defines the unbonding period in epochs before an account can stop staking.
	StakingUnbondingPeriod EpochIndex `serix:"10,mapKey=stakingUnbondingPeriod"`

	// LivenessThreshold is used by tip-selection to determine the if a block is eligible by evaluating issuingTimes
	// and commitments in its past-cone to ATT and lastCommittedSlot respectively.
	LivenessThreshold SlotIndex `serix:"11,mapKey=livenessThreshold"`
	// MinCommittableAge is the minimum age relative to the accepted tangle time slot index that a slot can be committed.
	// For example, if the last accepted slot is in slot 100, and minCommittableAge=10, then the latest committed slot can be at most 100-10=90.
	MinCommittableAge SlotIndex `serix:"12,mapKey=minCommittableAge"`
	// MaxCommittableAge is the maximum age for a slot commitment to be included in a block relative to the slot index of the block issuing time.
	// For example, if the last accepted slot is in slot 100, and maxCommittableAge=20, then the oldest referencable commitment is 100-20=80.
	MaxCommittableAge SlotIndex `serix:"13,mapKey=maxCommittableAge"`
	// EpochNearingThreshold is used by the epoch orchestrator to detect the slot that should trigger a new committee
	// selection for the next and upcoming epoch.
	EpochNearingThreshold SlotIndex `serix:"14,mapKey=epochNearingThreshold"`
	// RMCParameters defines the parameters used by to calculate the Reference Mana Cost (RMC).
	CongestionControlParameters CongestionControlParameters `serix:"15,mapKey=congestionControlParameters"`

	VersionSignaling VersionSignaling `serix:"16,mapKey=versionSignaling"`
}

func (b basicProtocolParameters) Equals(other basicProtocolParameters) bool {
	return b.Version == other.Version &&
		b.NetworkName == other.NetworkName &&
		b.Bech32HRP == other.Bech32HRP &&
		b.RentStructure.Equals(other.RentStructure) &&
		b.WorkScoreStructure.Equals(other.WorkScoreStructure) &&
		b.TokenSupply == other.TokenSupply &&
		b.GenesisUnixTimestamp == other.GenesisUnixTimestamp &&
		b.SlotDurationInSeconds == other.SlotDurationInSeconds &&
		b.SlotsPerEpochExponent == other.SlotsPerEpochExponent &&
		b.ManaStructure.Equals(other.ManaStructure) &&
		b.StakingUnbondingPeriod == other.StakingUnbondingPeriod &&
		b.LivenessThreshold == other.LivenessThreshold &&
		b.MinCommittableAge == other.MinCommittableAge &&
		b.MaxCommittableAge == other.MaxCommittableAge &&
		b.EpochNearingThreshold == other.EpochNearingThreshold &&
		b.CongestionControlParameters.Equals(other.CongestionControlParameters) &&
		b.VersionSignaling.Equals(other.VersionSignaling)
}
