package iotago

import (
	"cmp"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrInvalidStakingStartEpoch gets returned when a new Staking Feature's start epoch
	// is not set to the epoch of the transaction.
	ErrInvalidStakingStartEpoch = ierrors.New("staking start epoch must be the epoch of the transaction")
	// ErrInvalidStakingEndEpochTooEarly gets returned when a new Staking Feature's end epoch
	// is not at least set to the transaction epoch plus the unbonding period.
	ErrInvalidStakingEndEpochTooEarly = ierrors.New("staking end epoch must be set to the transaction epoch plus the unbonding period")
	// ErrInvalidStakingBlockIssuerRequired gets returned when an account contains a Staking Feature
	// but no Block Issuer Feature.
	ErrInvalidStakingBlockIssuerRequired = ierrors.New("staking feature requires a block issuer feature")
	// ErrInvalidStakingBondedRemoval gets returned when a staking feature is removed before the end of the unbonding period.
	ErrInvalidStakingBondedRemoval = ierrors.New("staking feature can only be removed after the unbonding period")
	// ErrInvalidStakingBondedModified gets returned when a staking feature's start epoch, fixed cost or
	// staked amount are modified before the unboding period.
	ErrInvalidStakingBondedModified = ierrors.New("staking start epoch, fixed cost and staked amount cannot be modified while bonded")
	// ErrInvalidStakingRewardInputRequired get returned when a staking feature is removed or resetted without a reward input.
	ErrInvalidStakingRewardInputRequired = ierrors.New("staking feature removal or resetting requires a reward input")
	// ErrInvalidStakingRewardClaim gets returned when mana rewards are claimed without removing or resetting the staking feature.
	ErrInvalidStakingRewardClaim = ierrors.New("staking feature must be removed or reset in order to claim rewards")
	// ErrInvalidStakingCommitmentInput gets returned when no commitment input was passed in a TX containing a staking feature.
	ErrInvalidStakingCommitmentInput = ierrors.New("staking feature validation requires a commitment input")
)

// StakingFeature is a feature which indicates that this account wants to register as a validator.
// The feature includes a fixed cost that the staker can set and will receive as part of its rewards,
// as well as a range of epoch indices in which the feature is considered active and can claim rewards.
// Removing the feature can only be done by going through an unbonding period.
type StakingFeature struct {
	StakedAmount BaseToken  `serix:""`
	FixedCost    Mana       `serix:""`
	StartEpoch   EpochIndex `serix:""`
	EndEpoch     EpochIndex `serix:""`
}

func (s *StakingFeature) Clone() Feature {
	return &StakingFeature{StakedAmount: s.StakedAmount, FixedCost: s.FixedCost, StartEpoch: s.StartEpoch, EndEpoch: s.EndEpoch}
}

func (s *StakingFeature) StorageScore(storageScoreStruct *StorageScoreStructure, f StorageScoreFunc) StorageScore {
	if f != nil {
		return f(storageScoreStruct)
	}

	return storageScoreStruct.OffsetStakingFeature()
}

func (s *StakingFeature) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	// staking feature changes require invokation of staking managers so require extra work.
	return workScoreParameters.Staking, nil
}

func (s *StakingFeature) Compare(other Feature) int {
	return cmp.Compare(s.Type(), other.Type())
}

func (s *StakingFeature) Equal(other Feature) bool {
	otherFeat, is := other.(*StakingFeature)
	if !is {
		return false
	}

	return s.StakedAmount == otherFeat.StakedAmount &&
		s.FixedCost == otherFeat.FixedCost &&
		s.StartEpoch == otherFeat.StartEpoch &&
		s.EndEpoch == otherFeat.EndEpoch
}

func (s *StakingFeature) Type() FeatureType {
	return FeatureStaking
}

func (s *StakingFeature) Size() int {
	// FeatureType + StakedAmount + FixedCost + StartEpoch + EndEpoch
	return serializer.SmallTypeDenotationByteSize + BaseTokenSize + ManaSize + EpochIndexLength + EpochIndexLength
}
