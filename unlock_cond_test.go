package iotago_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
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
			name: "ok - DustDepositReturnUnlockCondition",
			source: &iotago.DustDepositReturnUnlockCondition{
				ReturnAddress: tpkg.RandEd25519Address(),
				Amount:        1337,
			},
			target: &iotago.DustDepositReturnUnlockCondition{},
		},
		{
			name: "ok - TimelockUnlockCondition",
			source: &iotago.TimelockUnlockCondition{
				MilestoneIndex: 100,
			},
			target: &iotago.TimelockUnlockCondition{},
		},
		{
			name: "ok - ExpirationUnlockCondition",
			source: &iotago.ExpirationUnlockCondition{
				ReturnAddress: tpkg.RandEd25519Address(),
				UnixTime:      1000,
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
