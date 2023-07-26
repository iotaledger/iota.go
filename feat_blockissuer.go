package iotago

import (
	"bytes"

	"golang.org/x/exp/slices"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// MinBlockIssuerKeysCount is the minimum amount of block issuer keys allowed for a BlockIssuerFeature.
	MinBlockIssuerKeysCount = 1
	// MaxBlockIssuerKeysCount is the maximum amount of block issuer keys allowed for a BlockIssuerFeature.
	MaxBlockIssuerKeysCount = 128
)

// BlockIssuerKeys are the keys allowed to issue blocks from an account with a BlockIssuerFeature.
type BlockIssuerKeys []ed25519.PublicKey

func (s BlockIssuerKeys) Sort() {
	slices.SortFunc(s, func(a, b ed25519.PublicKey) bool {
		return bytes.Compare(a[:], b[:]) < 0
	})
}

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
	// TODO: add factor for block issuer keys (higher than regular keys factor).
	// VBFactorData: type prefix + expiry slot + keys length
	// VBFactorKey: numKeys * pubKeyLength
	return rentStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize+serializer.UInt64ByteSize+serializer.OneByte) +
		rentStruct.VBFactorKey.Multiply(VBytes(len(s.BlockIssuerKeys))*(ed25519.PublicKeySize))
}

func (s *BlockIssuerFeature) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	workScoreBytes, err := workScoreStructure.DataByte.Multiply(s.Size())
	if err != nil {
		return 0, err
	}

	// block issuer feature requires invocation of account and mana managers, so requires extra work.
	return workScoreBytes.Add(workScoreStructure.BlockIssuer)
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
		if !bytes.Equal(s.BlockIssuerKeys[i][:], otherFeat.BlockIssuerKeys[i][:]) {
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
	return serializer.SmallTypeDenotationByteSize + serializer.OneByte + len(s.BlockIssuerKeys)*ed25519.PublicKeySize + SlotIndexLength
}
