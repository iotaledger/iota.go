package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3/util"
)

// TaggedData is a payload which holds a tag and associated data.
type TaggedData struct {
	// The tag to use to categorize the data.
	Tag []byte `serix:"0,lengthPrefixType=uint8,mapKey=tag,omitempty,maxLen=64"`
	// The data within the payload.
	Data []byte `serix:"1,lengthPrefixType=uint32,mapKey=data,maxLen=8192"`
}

func (u *TaggedData) PayloadType() PayloadType {
	return PayloadTaggedData
}

func (u *TaggedData) Size() int {
	return util.NumByteLen(uint32(PayloadTaggedData)) +
		serializer.OneByte + len(u.Tag) +
		serializer.UInt32ByteSize + len(u.Data)
}
