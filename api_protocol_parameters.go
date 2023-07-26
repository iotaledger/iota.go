package iotago

import (
	"github.com/iotaledger/hive.go/lo"
)

type basicProtocolParameters struct {
	// Version defines the version of the protocol this protocol parameters are for.
	Version Version `serix:"0,mapKey=version"`

	// NetworkName defines the human friendly name of the network.
	NetworkName string `serix:"1,lengthPrefixType=uint8,mapKey=networkName"`
	// Bech32HRP defines the HRP prefix used for Bech32 addresses in the network.
	Bech32HRP NetworkPrefix `serix:"2,lengthPrefixType=uint8,mapKey=bech32Hrp"`

	// RentStructure defines the rent structure used by given node/network.
	RentStructure RentStructure `serix:"3,mapKey=rentStructure"`
	// TokenSupply defines the current token supply on the network.
	TokenSupply BaseToken `serix:"4,mapKey=tokenSupply"`

	// GenesisUnixTimestamp defines the genesis timestamp at which the slots start to count.
	GenesisUnixTimestamp int64 `serix:"5,mapKey=genesisUnixTimestamp"`
	// SlotDurationInSeconds defines the duration of each slot in seconds.
	SlotDurationInSeconds uint8 `serix:"6,mapKey=slotDurationInSeconds"`
	// SlotsPerEpochExponent is the number of slots in an epoch expressed as an exponent of 2.
	// (2**SlotsPerEpochExponent) == slots in an epoch.
	SlotsPerEpochExponent uint8 `serix:"7,mapKey=slotsPerEpochExponent"`

	// ManaGenerationRate is the amount of potential Mana generated by 1 IOTA in 1 slot.
	ManaGenerationRate uint8 `serix:"8,mapKey=manaGenerationRate"`
	// ManaGenerationRateExponent is the scaling of ManaGenerationRate expressed as an exponent of 2.
	ManaGenerationRateExponent uint8 `serix:"9,mapKey=manaGenerationRateExponent"`
	// ManaDecayFactors is a lookup table of epoch index diff to mana decay factor (slice index 0 = 1 epoch).
	ManaDecayFactors []uint32 `serix:"10,lengthPrefixType=uint16,mapKey=manaDecayFactors"`
	// ManaDecayFactorsExponent is the scaling of ManaDecayFactors expressed as an exponent of 2.
	ManaDecayFactorsExponent uint8 `serix:"11,mapKey=manaDecayFactorsExponent"`
	// ManaDecayFactorEpochsSum is an integer approximation of the sum of decay over epochs.
	ManaDecayFactorEpochsSum uint32 `serix:"12,mapKey=manaDecayFactorEpochsSum"`
	// ManaDecayFactorEpochsSumExponent is the scaling of ManaDecayFactorEpochsSum expressed as an exponent of 2.
	ManaDecayFactorEpochsSumExponent uint8 `serix:"13,mapKey=manaDecayFactorEpochsSumExponent"`

	// StakingUnbondingPeriod defines the unbonding period in epochs before an account can stop staking.
	StakingUnbondingPeriod EpochIndex `serix:"14,mapKey=stakingUnbondingPeriod"`

	// EvictionAge defines the age in slots when you can evict blocks by committing them into a slot commitments and
	// when slots stop being a consumable accounts' state relative to the latest committed slot.
	EvictionAge SlotIndex `serix:"15,mapKey=evictionAge"`
	// LivenessThreshold is used by tip-selection to determine the if a block is eligible by evaluating issuingTimes
	// and commitments in its past-cone to ATT and lastCommittedSlot respectively.
	LivenessThreshold SlotIndex `serix:"16,mapKey=livenessThreshold"`
	// EpochNearingThreshold is used by the epoch orchestrator to detect the slot that should trigger a new committee
	// selection for the next and upcoming epoch.
	EpochNearingThreshold SlotIndex `serix:"17,mapKey=epochNearingThreshold"`

	VersionSignaling VersionSignaling `serix:"18,mapKey=versionSignaling"`
}

func (b basicProtocolParameters) Equals(other basicProtocolParameters) bool {
	return b.Version == other.Version &&
		b.NetworkName == other.NetworkName &&
		b.Bech32HRP == other.Bech32HRP &&
		b.RentStructure.Equals(other.RentStructure) &&
		b.TokenSupply == other.TokenSupply &&
		b.GenesisUnixTimestamp == other.GenesisUnixTimestamp &&
		b.SlotDurationInSeconds == other.SlotDurationInSeconds &&
		b.SlotsPerEpochExponent == other.SlotsPerEpochExponent &&
		b.ManaGenerationRate == other.ManaGenerationRate &&
		b.ManaGenerationRateExponent == other.ManaGenerationRateExponent &&
		lo.Equal(b.ManaDecayFactors, other.ManaDecayFactors) &&
		b.ManaDecayFactorsExponent == other.ManaDecayFactorsExponent &&
		b.ManaDecayFactorEpochsSum == other.ManaDecayFactorEpochsSum &&
		b.ManaDecayFactorEpochsSumExponent == other.ManaDecayFactorEpochsSumExponent &&
		b.StakingUnbondingPeriod == other.StakingUnbondingPeriod &&
		b.EvictionAge == other.EvictionAge &&
		b.LivenessThreshold == other.LivenessThreshold &&
		b.EpochNearingThreshold == other.EpochNearingThreshold &&
		b.VersionSignaling.Equals(other.VersionSignaling)
}
