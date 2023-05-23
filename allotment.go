package iotago

import "github.com/iotaledger/hive.go/serializer/v2"

type Allotments map[AccountID]uint64

func (a Allotments) Size() int {
	return len(a) * serializer.UInt64ByteSize
}
