package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// ReferenceUnlockSize defines the size of a ReferenceUnlock.
	ReferenceUnlockSize = serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
)

// ReferenceUnlock is an Unlock which references a previous unlock.
type ReferenceUnlock struct {
	// The other unlock this ReferenceUnlock references to.
	Reference uint16 `serix:"0,mapKey=reference"`
}

func (r *ReferenceUnlock) SourceAllowed(address Address) bool {
	_, ok := address.(ChainAddress)
	return !ok
}

func (r *ReferenceUnlock) Chainable() bool {
	return false
}

func (r *ReferenceUnlock) Ref() uint16 {
	return r.Reference
}

func (r *ReferenceUnlock) Type() UnlockType {
	return UnlockReference
}

func (r *ReferenceUnlock) Size() int {
	return ReferenceUnlockSize
}
