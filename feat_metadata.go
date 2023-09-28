package iotago

import (
	"bytes"

	"github.com/iotaledger/hive.go/serializer/v2"
)

// MetadataFeature is a feature which simply holds binary data to be freely
// interpreted by higher layer applications.
type MetadataFeature struct {
	Data []byte `serix:"0,lengthPrefixType=uint16,mapKey=data,minLen=1,maxLen=8192"`
}

func (s *MetadataFeature) Clone() Feature {
	return &MetadataFeature{Data: append([]byte(nil), s.Data...)}
}

func (s *MetadataFeature) VBytes(rentStruct *RentStructure, _ VBytesFunc) VBytes {
	return 0
}

func (s *MetadataFeature) WorkScore(_ *WorkScoreStructure) (WorkScore, error) {
	return 0, nil
}

func (s *MetadataFeature) Equal(other Feature) bool {
	otherFeat, is := other.(*MetadataFeature)
	if !is {
		return false
	}

	return bytes.Equal(s.Data, otherFeat.Data)
}

func (s *MetadataFeature) Type() FeatureType {
	return FeatureMetadata
}

func (s *MetadataFeature) Size() int {
	// FeatureType + Data
	return serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize + len(s.Data)
}
