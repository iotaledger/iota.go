package api_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/iotaledger/iota.go/v4/tpkg/frameworks"
)

func Test_BlockIssuerAPIDeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok",
			Source: &api.BlockIssuerInfo{
				BlockIssuerAddress:     tpkg.RandAccountAddress().Bech32(iotago.PrefixTestnet),
				PowTargetTrailingZeros: 10,
			},
			Target:    &api.BlockIssuerInfo{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func Test_BlockIssuerAPIJSONSerialization(t *testing.T) {
	tests := []*frameworks.JSONEncodeTest{
		{
			Name: "ok - BlockIssuerInfo",
			Source: &api.BlockIssuerInfo{
				BlockIssuerAddress:     iotago.AccountID{0xFF}.ToAddress().Bech32(iotago.PrefixTestnet),
				PowTargetTrailingZeros: 10,
			},
			Target: `{
	"blockIssuerAddress": "rms1prlsqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqcyz9fx",
	"powTargetTrailingZeros": 10
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}
