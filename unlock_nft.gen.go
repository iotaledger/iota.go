package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

// NFTUnlock is an Unlock which references a previous unlock.
type NFTUnlock struct {
	// The other unlock this NFTUnlock references to.
	Reference uint16 `serix:""`
}

func (r *NFTUnlock) Clone() Unlock {
	return &NFTUnlock{
		Reference: r.Reference,
	}
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
	// UnlockType + Reference
	return serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
}

func (r *NFTUnlock) WorkScore(_ *WorkScoreParameters) (WorkScore, error) {
	return 0, nil
}
