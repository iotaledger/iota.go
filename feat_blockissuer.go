package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// MinBlockIssuerKeysCount is the minimum amount of block issuer keys allowed for a BlockIssuerFeature.
	MinBlockIssuerKeysCount = 1
	// MaxBlockIssuerKeysCount is the maximum amount of block issuer keys allowed for a BlockIssuerFeature.
	MaxBlockIssuerKeysCount = 128
)

// BlockIssuerFeature is a feature which indicates that this account can issue blocks.
// The feature includes a block issuer address as well as an expiry slot.
type BlockIssuerFeature struct {
	BlockIssuerKeys BlockIssuerKeys `serix:"0,mapKey=blockIssuerKeys,lengthPrefixType=uint8"`
	ExpirySlot      SlotIndex       `serix:"1,mapKey=expirySlot"`
}

func (s *BlockIssuerFeature) Clone() Feature {
	return &BlockIssuerFeature{BlockIssuerKeys: s.BlockIssuerKeys, ExpirySlot: s.ExpirySlot}
}

func (s *BlockIssuerFeature) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	// VBFactorData: type prefix + expiry slot
	// + block issuer keys vbytes
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+SlotIndexLength) +
		s.BlockIssuerKeys.VBytes(rentStruct, nil)
}

func (s *BlockIssuerFeature) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	// block issuer feature requires invocation of account and mana managers, so requires extra work.
	return workScoreStructure.BlockIssuer, nil
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
		if !s.BlockIssuerKeys[i].Equal(otherFeat.BlockIssuerKeys[i]) {
			return false
		}
	}

	return s.ExpirySlot == otherFeat.ExpirySlot
}

func (s *BlockIssuerFeature) Type() FeatureType {
	return FeatureBlockIssuer
}

func (s *BlockIssuerFeature) Size() int {
	// FeatureType + BlockIssuerKeys + ExpirySlot
	return serializer.SmallTypeDenotationByteSize + s.BlockIssuerKeys.Size() + SlotIndexLength
}
