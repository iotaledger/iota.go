package iotago

import (
	"github.com/iotaledger/hive.go/serializer/v2"
)

// AnchorUnlock is an Unlock which references a previous input/unlock.
type AnchorUnlock struct {
	// The other input/unlock this AnchorUnlock references to.
	Reference uint16 `serix:""`
}

func (r *AnchorUnlock) Clone() Unlock {
	return &AnchorUnlock{
		Reference: r.Reference,
	}
}

func (r *AnchorUnlock) SourceAllowed(address Address) bool {
	_, ok := address.(*AnchorAddress)

	return ok
}

func (r *AnchorUnlock) Chainable() bool {
	return true
}

func (r *AnchorUnlock) ReferencedInputIndex() uint16 {
	return r.Reference
}

func (r *AnchorUnlock) Type() UnlockType {
	return UnlockAnchor
}

func (r *AnchorUnlock) Size() int {
	// UnlockType + Reference
	return serializer.SmallTypeDenotationByteSize + serializer.UInt16ByteSize
}

func (r *AnchorUnlock) WorkScore(_ *WorkScoreParameters) (WorkScore, error) {
	return 0, nil
}
