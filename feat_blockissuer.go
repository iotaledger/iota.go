package iotago

import (
	"crypto/ed25519"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/util"
)

// BlockIssuerFeature is a feature which indicates that this account can issue blocks.
// The feature includes a block issuer address as well as an expiry slot.
type BlockIssuerFeature struct {
	BlockIssuerKeys []ed25519.PublicKey `serix:"0,mapKey=blockIssuerKeys"`
	ExpirySlot      SlotIndex           `serix:"1,mapKey=expirySlot"`
}

func (s *BlockIssuerFeature) Clone() Feature {
	return &BlockIssuerFeature{BlockIssuerKeys: s.BlockIssuerKeys, ExpirySlot: s.ExpirySlot}
}

// TODO: add factor for block issuer keys (higher than regular keys factor)
func (s *BlockIssuerFeature) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt32ByteSize) +
		rentStruct.VBFactorKey.Multiply(uint64(len(s.BlockIssuerKeys))*ed25519.PublicKeySize)
}

func (s *BlockIssuerFeature) Equal(other Feature) bool {
	otherFeat, is := other.(*BlockIssuerFeature)
	if !is {
		return false
	}
	if len(s.BlockIssuerKeys) != len(otherFeat.BlockIssuerKeys) {
		return false
	}
	for i := range s.BlockIssuerKeys {
		if s.BlockIssuerKeys[i].Equal(otherFeat.BlockIssuerKeys[i]) {
			return false
		}
	}
	return s.ExpirySlot == otherFeat.ExpirySlot
}

func (s *BlockIssuerFeature) Type() FeatureType {
	return FeatureBlockIssuer
}

func (s *BlockIssuerFeature) Size() int {
	return util.NumByteLen(byte(FeatureBlockIssuer)) + len(s.BlockIssuerKeys)*ed25519.PublicKeySize + serializer.UInt32ByteSize
}
