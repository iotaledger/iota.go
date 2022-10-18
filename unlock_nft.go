package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// NFTUnlockSize defines the size of an NFTUnlock.
	NFTUnlockSize = serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
)

// NFTUnlock is an Unlock which references a previous unlock.
type NFTUnlock struct {
	// The other unlock this NFTUnlock references to.
	Reference uint16 `serix:"0,mapKey=reference"`
}

func (r *NFTUnlock) SourceAllowed(address Address) bool {
	_, ok := address.(*NFTAddress)
	return ok
}

func (r *NFTUnlock) Chainable() bool {
	return true
}

func (r *NFTUnlock) Ref() uint16 {
	return r.Reference
}

func (r *NFTUnlock) Type() UnlockType {
	return UnlockNFT
}

func (r *NFTUnlock) Size() int {
	return NFTUnlockSize
}
