package iotago_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

func TestFeatureBlocksDeSerialize(t *testing.T) {
	tests := []deSerializeTest{
		{
			name:   "ok - SenderFeatureBlock",
			source: &iotago.SenderFeatureBlock{Address: tpkg.RandEd25519Address()},
			target: &iotago.SenderFeatureBlock{},
		},
		{
			name:   "ok - IssuerFeatureBlock",
			source: &iotago.IssuerFeatureBlock{Address: tpkg.RandEd25519Address()},
			target: &iotago.IssuerFeatureBlock{},
		},
		{
			name: "ok - DustDepositReturnFeatureBlock",
			source: &iotago.DustDepositReturnFeatureBlock{
				Amount: 1337,
			},
			target: &iotago.DustDepositReturnFeatureBlock{},
		},
		{
			name: "ok - TimelockMilestoneIndexFeatureBlock",
			source: &iotago.TimelockMilestoneIndexFeatureBlock{
				MilestoneIndex: 100,
			},
			target: &iotago.TimelockMilestoneIndexFeatureBlock{},
		},
		{
			name: "ok - TimelockUnixFeatureBlock",
			source: &iotago.TimelockUnixFeatureBlock{
				UnixTime: 1000,
			},
			target: &iotago.TimelockUnixFeatureBlock{},
		},
		{
			name: "ok - ExpirationMilestoneIndexFeatureBlock",
			source: &iotago.ExpirationMilestoneIndexFeatureBlock{
				MilestoneIndex: 100,
			},
			target: &iotago.ExpirationMilestoneIndexFeatureBlock{},
		},
		{
			name: "ok - ExpirationUnixFeatureBlock",
			source: &iotago.ExpirationUnixFeatureBlock{
				UnixTime: 1000,
			},
			target: &iotago.ExpirationUnixFeatureBlock{},
		},
		{
			name: "ok - MetadataFeatureBlock",
			source: &iotago.MetadataFeatureBlock{
				Data: []byte("hello world"),
			},
			target: &iotago.MetadataFeatureBlock{},
		},
		{
			name: "ok - IndexationFeatureBlock",
			source: &iotago.IndexationFeatureBlock{
				Tag: []byte("hello world"),
			},
			target: &iotago.IndexationFeatureBlock{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.deSerialize)
	}
}
