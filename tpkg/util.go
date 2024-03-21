package tpkg

import (
	iotago "github.com/iotaledger/iota.go/v4"
)

// Must panics if the given error is not nil.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// ReferenceUnlock returns a reference unlock with the given index.
func ReferenceUnlock(index uint16) *iotago.ReferenceUnlock {
	return &iotago.ReferenceUnlock{Reference: index}
}
