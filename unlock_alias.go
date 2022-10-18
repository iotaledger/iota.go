package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// AliasUnlockSize defines the size of an AliasUnlock.
	AliasUnlockSize = serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
)

// AliasUnlock is an Unlock which references a previous unlock.
type AliasUnlock struct {
	// The other unlock this AliasUnlock references to.
	Reference uint16 `serix:"0,mapKey=reference"`
}

func (r *AliasUnlock) SourceAllowed(address Address) bool {
	_, ok := address.(*AliasAddress)
	return ok
}

func (r *AliasUnlock) Chainable() bool {
	return true
}

func (r *AliasUnlock) Ref() uint16 {
	return r.Reference
}

func (r *AliasUnlock) Type() UnlockType {
	return UnlockAlias
}

func (r *AliasUnlock) Size() int {
	return AliasUnlockSize
}
