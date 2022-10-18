package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

// IssuerFeature is a feature which associates an output
// with an issuer identity. Unlike the SenderFeature, the issuer identity
// only has to be unlocked when the ChainOutput is first created,
// afterwards, the issuer feature must not change, meaning that subsequent outputs
// must always define the same issuer identity (the identity does not need to be unlocked anymore though).
type IssuerFeature struct {
	Address Address `serix:"0,mapKey=address"`
}

func (s *IssuerFeature) Clone() Feature {
	return &IssuerFeature{Address: s.Address.Clone()}
}

func (s *IssuerFeature) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize) +
		s.Address.VBytes(rentStruct, nil)
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
	return util.NumByteLen(byte(FeatureIssuer)) + s.Address.Size()
}
