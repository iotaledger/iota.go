package iotago_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/iotaledger/iota.go/v4/tpkg/frameworks"
)

func TestUnlockConditionsDeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok - AddressUnlockCondition",
			Source: &iotago.AddressUnlockCondition{
				Address: tpkg.RandEd25519Address(),
			},
			Target: &iotago.AddressUnlockCondition{},
		},
		{
			Name: "ok - StorageDepositReturnUnlockCondition",
			Source: &iotago.StorageDepositReturnUnlockCondition{
				ReturnAddress: tpkg.RandEd25519Address(),
				Amount:        1337,
			},
			Target: &iotago.StorageDepositReturnUnlockCondition{},
		},
		{
			Name: "ok - TimelockUnlockCondition",
			Source: &iotago.TimelockUnlockCondition{
				Slot: 1000,
			},
			Target: &iotago.TimelockUnlockCondition{},
		},
		{
			Name: "ok - ExpirationUnlockCondition",
			Source: &iotago.ExpirationUnlockCondition{
				ReturnAddress: tpkg.RandEd25519Address(),
				Slot:          1000,
			},
			Target: &iotago.ExpirationUnlockCondition{},
		},
		{
			Name: "ok - StateControllerAddressUnlockCondition",
			Source: &iotago.StateControllerAddressUnlockCondition{
				Address: tpkg.RandEd25519Address(),
			},
			Target: &iotago.StateControllerAddressUnlockCondition{},
		},
		{
			Name: "ok - GovernorAddressUnlockCondition",
			Source: &iotago.GovernorAddressUnlockCondition{
				Address: tpkg.RandEd25519Address(),
			},
			Target: &iotago.GovernorAddressUnlockCondition{},
		},
		{
			Name: "fail - ImplicitAccountCreationAddress in GovernorAddressUnlockCondition",
			Source: &iotago.GovernorAddressUnlockCondition{
				Address: tpkg.RandImplicitAccountCreationAddress(),
			},
			Target:    &iotago.GovernorAddressUnlockCondition{},
			SeriErr:   iotago.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
			DeSeriErr: iotago.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
		},
		{
			Name: "fail - ImplicitAccountCreationAddress in StateControllerAddressUnlockCondition",
			Source: &iotago.StateControllerAddressUnlockCondition{
				Address: tpkg.RandImplicitAccountCreationAddress(),
			},
			Target:    &iotago.StateControllerAddressUnlockCondition{},
			SeriErr:   iotago.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
			DeSeriErr: iotago.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
		},
		{
			Name: "fail - ImplicitAccountCreationAddress in ExpirationUnlockCondition",
			Source: &iotago.ExpirationUnlockCondition{
				Slot:          3,
				ReturnAddress: tpkg.RandImplicitAccountCreationAddress(),
			},
			Target:    &iotago.ExpirationUnlockCondition{},
			SeriErr:   iotago.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
			DeSeriErr: iotago.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
		},
		{
			Name: "fail - ImplicitAccountCreationAddress in StorageDepositReturnUnlockCondition",
			Source: &iotago.StorageDepositReturnUnlockCondition{
				ReturnAddress: tpkg.RandImplicitAccountCreationAddress(),
			},
			Target:    &iotago.StorageDepositReturnUnlockCondition{},
			SeriErr:   iotago.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
			DeSeriErr: iotago.ErrImplicitAccountCreationAddressInInvalidUnlockCondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}
