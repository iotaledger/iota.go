package iotago

import (
	"bytes"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/util"
)

// TagFeature is a feature which allows to additionally tag an output by a user defined value.
type TagFeature struct {
	Tag []byte `serix:"0,lengthPrefixType=uint8,mapKey=tag,minLen=1,maxLen=64"`
}

func (s *TagFeature) Clone() Feature {
	return &TagFeature{Tag: append([]byte(nil), s.Tag...)}
}

func (s *TagFeature) VBytes(rentStruct *RentStructure, f VBytesFunc) VBytes {
	if f != nil {
		return f(rentStruct)
	}
	return rentStruct.VBFactorData.Multiply(VBytes(serializer.SmallTypeDenotationByteSize + serializer.OneByte + len(s.Tag)))
}

func (s *TagFeature) WorkScore(workScoreStructure *WorkScoreStructure) WorkScore {
	return workScoreStructure.Factors.Data.Multiply(s.Size())
}

func (s *TagFeature) Equal(other Feature) bool {
	otherFeat, is := other.(*TagFeature)
	if !is {
		return false
	}

	return bytes.Equal(s.Tag, otherFeat.Tag)
}

func (s *TagFeature) Type() FeatureType {
	return FeatureTag
}

func (s *TagFeature) Size() int {
	// tag length prefix = 1 byte
	return util.NumByteLen(byte(FeatureSender)) + serializer.OneByte + len(s.Tag)
}
