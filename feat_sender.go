package iotago

import (
	"cmp"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// SenderFeature is a feature which associates an output
// with a sender address. The sender address needs to be unlocked in the transaction
// for the SenderFeature to be valid.
type SenderFeature struct {
	Address Address `serix:""`
}

func (s *SenderFeature) Clone() Feature {
	return &SenderFeature{Address: s.Address.Clone()}
}

func (s *SenderFeature) StorageScore(storageScoreStruct *StorageScoreStructure, f StorageScoreFunc) StorageScore {
	if f != nil {
		return f(storageScoreStruct)
	}

	return s.Address.StorageScore(storageScoreStruct, nil)
}

func (s *SenderFeature) WorkScore(_ *WorkScoreParameters) (WorkScore, error) {
	// we do not need to charge for a signature check here as this is covered by the unlock that must be provided.
	return 0, nil
}

func (s *SenderFeature) Compare(other Feature) int {
	return cmp.Compare(s.Type(), other.Type())
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
