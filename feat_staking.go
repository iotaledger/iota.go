package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/util"
)

// StakingFeature is a feature which indicates that this account wants to register as a validator.
// The feature includes a fixed cost that the staker can set and will receive as part of its rewards,
// as well as a range of epoch indices in which the feature is considered active and can claim rewards.
// Removing the feature can only be done by going through an unbonding period.
type StakingFeature struct {
	StakedAmount uint64 `serix:"0,mapKey=stakedAmount"`
	FixedCost    uint64 `serix:"1,mapKey=fixedCost"`
	StartEpoch   uint64 `serix:"2,mapKey=startEpoch"`
	EndEpoch     uint64 `serix:"3,mapKey=endEpoch"`
}

func (s *StakingFeature) Clone() Feature {
	return &StakingFeature{StakedAmount: s.StakedAmount, FixedCost: s.FixedCost, StartEpoch: s.StartEpoch, EndEpoch: s.EndEpoch}
}

func (s *StakingFeature) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	vbytes := serializer.SmallTypeDenotationByteSize + (serializer.UInt64ByteSize * 4)
	// TODO: Introduce another vbyte factor for the staking feature.
	return rentStruct.VBFactorData.Multiply(VBytes(vbytes))
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
	return util.NumByteLen(byte(FeatureStaking)) + serializer.UInt64ByteSize*4
}
