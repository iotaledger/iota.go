package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

// SenderFeature is a feature which associates an output
// with a sender identity. The sender identity needs to be unlocked in the transaction
// for the SenderFeature to be valid.
type SenderFeature struct {
	Address Address `serix:"0,mapKey=address"`
}

func (s *SenderFeature) Clone() Feature {
	return &SenderFeature{Address: s.Address.Clone()}
}

func (s *SenderFeature) StorageScore(rentStruct *RentStructure, f StorageScoreFunc) StorageScore {
	if f != nil {
		return f(rentStruct)
	}

	return s.Address.StorageScore(rentStruct, nil)
}

func (s *SenderFeature) WorkScore(_ *WorkScoreStructure) (WorkScore, error) {
	// we do not need to charge for a signature check here as this is covered by the unlock that must be provided.
	return 0, nil
}

func (s *SenderFeature) Equal(other Feature) bool {
	otherFeat, is := other.(*SenderFeature)
	if !is {
		return false
	}

	return s.Address.Equal(otherFeat.Address)
}

func (s *SenderFeature) Type() FeatureType {
	return FeatureSender
}

func (s *SenderFeature) Size() int {
	// FeatureType + Address
	return serializer.SmallTypeDenotationByteSize + s.Address.Size()
}
