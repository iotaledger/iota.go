package iotago_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
)

func TestUnlockConditionsDeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name: "ok - AddressUnlockCondition",
			source: &iotago.AddressUnlockCondition{
				Address: tpkg.RandEd25519Address(),
			},
			target: &iotago.AddressUnlockCondition{},
		},
		{
			name: "ok - StorageDepositReturnUnlockCondition",
			source: &iotago.StorageDepositReturnUnlockCondition{
				ReturnAddress: tpkg.RandEd25519Address(),
				Amount:        1337,
			},
			target: &iotago.StorageDepositReturnUnlockCondition{},
		},
		{
			name: "ok - TimelockUnlockCondition",
			source: &iotago.TimelockUnlockCondition{
				SlotIndex: 1000,
			},
			target: &iotago.TimelockUnlockCondition{},
		},
		{
			name: "ok - ExpirationUnlockCondition",
			source: &iotago.ExpirationUnlockCondition{
				ReturnAddress: tpkg.RandEd25519Address(),
				SlotIndex:     1000,
			},
			target: &iotago.ExpirationUnlockCondition{},
		},
		{
			name: "ok - StateControllerAddressUnlockCondition",
			source: &iotago.StateControllerAddressUnlockCondition{
				Address: tpkg.RandEd25519Address(),
			},
			target: &iotago.StateControllerAddressUnlockCondition{},
		},
		{
			name: "ok - GovernorAddressUnlockCondition",
			source: &iotago.GovernorAddressUnlockCondition{
				Address: tpkg.RandEd25519Address(),
			},
			target: &iotago.GovernorAddressUnlockCondition{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
