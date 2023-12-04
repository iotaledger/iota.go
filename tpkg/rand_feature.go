package tpkg

import (
	iotago "github.com/iotaledger/iota.go/v4"
)

// RandNativeTokenFeature returns a random NativeToken feature.
func RandNativeTokenFeature() *iotago.NativeTokenFeature {
	return &iotago.NativeTokenFeature{
		ID:     RandNativeTokenID(),
		Amount: RandUint256(),
	}
}
