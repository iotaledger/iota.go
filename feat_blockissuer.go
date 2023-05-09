package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/util"
)

// BlockIssuerFeature is a feature which indicates that this account can issue blocks.
// The feature includes a block issuer address as well as an expiry slot.
type BlockIssuerFeature struct {
	Address    Address `serix:"0,mapKey=address"`
	ExpiryTime uint32  `serix:"1,mapKey=expirytime"`
}

func (s *BlockIssuerFeature) Clone() Feature {
	return &BlockIssuerFeature{Address: s.Address.Clone(), ExpiryTime: s.ExpiryTime}
}

func (s *BlockIssuerFeature) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt32ByteSize) +
		s.Address.VBytes(rentStruct, nil)
}

func (s *BlockIssuerFeature) Equal(other Feature) bool {
	otherFeat, is := other.(*BlockIssuerFeature)
	if !is {
		return false
	}

	return s.Address.Equal(otherFeat.Address) && s.ExpiryTime == otherFeat.ExpiryTime
}

func (s *BlockIssuerFeature) Type() FeatureType {
	return FeatureBlockIssuer
}

func (s *BlockIssuerFeature) Size() int {
	return util.NumByteLen(byte(FeatureBlockIssuer)) + s.Address.Size() + serializer.UInt32ByteSize
}
