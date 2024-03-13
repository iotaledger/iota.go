package iotago

import (
	"cmp"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrStakingStartEpochInvalid gets returned when a new Staking Feature's start epoch
	// is not set to the epoch of the transaction.
	ErrStakingStartEpochInvalid = ierrors.New("staking start epoch must be the epoch of the transaction")
	// ErrStakingEndEpochTooEarly gets returned when a new Staking Feature's end epoch
	// is not at least set to the transaction epoch plus the unbonding period.
	ErrStakingEndEpochTooEarly = ierrors.New("staking end epoch must be set to the transaction epoch plus the unbonding period")
	// ErrStakingBlockIssuerFeatureMissing gets returned when an account contains a Staking Feature
	// but no Block Issuer Feature.
	ErrStakingBlockIssuerFeatureMissing = ierrors.New("block issuer feature missing for account with staking feature")
	// ErrStakingFeatureRemovedBeforeUnbonding gets returned when a staking feature is removed before the end of the unbonding period.
	ErrStakingFeatureRemovedBeforeUnbonding = ierrors.New("staking feature can only be removed after the unbonding period")
	// ErrStakingFeatureModifiedBeforeUnbonding gets returned when a staking feature's start epoch, fixed cost or
	// staked amount are modified before the unboding period.
	ErrStakingFeatureModifiedBeforeUnbonding = ierrors.New("staking start epoch, fixed cost and staked amount cannot be modified while bonded")
	// ErrStakingRewardInputMissing get returned when a staking feature is removed or reset without a reward input.
	ErrStakingRewardInputMissing = ierrors.New("staking feature removal or resetting requires a reward input")
	// ErrStakingRewardClaimingInvalid gets returned when mana rewards are claimed without removing or resetting the staking feature.
	ErrStakingRewardClaimingInvalid = ierrors.New("staking feature must be removed or reset in order to claim rewards")
	// ErrStakingCommitmentInputMissing gets returned when no commitment input was passed in a TX containing a staking feature.
	ErrStakingCommitmentInputMissing = ierrors.New("staking feature validation requires a commitment input")
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
