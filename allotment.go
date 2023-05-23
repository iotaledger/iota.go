package iotago

import "github.com/iotaledger/hive.go/serializer/v2"

// Allotments is a slice of Allotment.
type Allotments []Allotment

// Allotment is a struct that represents a list of account IDs and an allotted value.
type Allotment struct {
	AccountID AccountID `serix:"0"`
	Amount    uint64    `serix:"1"`
}

func (a Allotments) Size() int {
	return len(a) * (AccountIDLength + serializer.UInt64ByteSize)
}
