package iotago

import (
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// TaggedData is a payload which holds a tag and associated data.
type TaggedData struct {
	// The tag to use to categorize the data.
	Tag []byte `serix:"0,lengthPrefixType=uint8,mapKey=tag,omitempty,maxLen=64"`
	// The data within the payload.
	Data []byte `serix:"1,lengthPrefixType=uint32,mapKey=data,maxLen=8192"`
}

func (u *TaggedData) Clone() Payload {
	return &TaggedData{
		Tag:  lo.CopySlice(u.Tag),
		Data: lo.CopySlice(u.Data),
	}
}

func (u *TaggedData) PayloadType() PayloadType {
	return PayloadTaggedData
}

func (u *TaggedData) Size() int {
	// PayloadType
	return serializer.TypeDenotationByteSize +
		serializer.OneByte + len(u.Tag) +
		serializer.UInt32ByteSize + len(u.Data)
}

func (u *TaggedData) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	// we account for the network traffic only on "Payload" level
	return workScoreParameters.DataByte.Multiply(u.Size())
}
