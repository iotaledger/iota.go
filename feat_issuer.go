//nolint:dupl
package iotago

import (
	"cmp"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// IssuerFeature is a feature which associates an output
// with an issuer address. Unlike the SenderFeature, the issuer address
// only has to be unlocked when the ChainOutput is first created,
// afterwards, the issuer feature must not change, meaning that subsequent outputs
// must always define the same issuer address (the address does not need to be unlocked anymore though).
type IssuerFeature struct {
	Address Address `serix:""`
}

func (s *IssuerFeature) Clone() Feature {
	return &IssuerFeature{Address: s.Address.Clone()}
}

func (s *IssuerFeature) StorageScore(storageScoreStruct *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return s.Address.StorageScore(storageScoreStruct, nil)
}

func (s *IssuerFeature) WorkScore(_ *WorkScoreParameters) (WorkScore, error) {
	// we do not need to charge for a signature check here as this is covered by the unlock that must be provided.
	return 0, nil
}

func (s *IssuerFeature) Compare(other Feature) int {
	return cmp.Compare(s.Type(), other.Type())
}

func (s *IssuerFeature) Equal(other Feature) bool {
	otherFeat, is := other.(*IssuerFeature)
	if !is {
		return false
	}

	return s.Address.Equal(otherFeat.Address)
}

func (s *IssuerFeature) Type() FeatureType {
	return FeatureIssuer
}

func (s *IssuerFeature) Size() int {
	// FeatureType + Address
	return serializer.SmallTypeDenotationByteSize + s.Address.Size()
}
