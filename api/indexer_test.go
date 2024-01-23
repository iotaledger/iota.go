package api_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
	"github.com/iotaledger/iota.go/v4/tpkg"
	"github.com/iotaledger/iota.go/v4/tpkg/frameworks"
)

func Test_IndexerAPIDeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok",
			Source: &api.IndexerResponse{
				CommittedSlot: tpkg.RandSlot(),
				PageSize:      1000,
				Items:         iotago.HexOutputIDsFromOutputIDs(tpkg.RandOutputIDs(2)...),
				Cursor:        "cursor-value",
			},
			Target:    &api.IndexerResponse{},
			SeriErr:   nil,
			DeSeriErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func Test_IndexerAPIJSONSerialization(t *testing.T) {
	tests := []*frameworks.JSONEncodeTest{
		{
			Name: "ok - IndexerResponse",
			Source: &api.IndexerResponse{
				CommittedSlot: 281,
				PageSize:      1000,
				Items: iotago.HexOutputIDsFromOutputIDs(
					iotago.OutputID{0xff},
					iotago.OutputID{0xfa},
				),
				Cursor: "cursor-value",
			},
			Target: `{
	"committedSlot": 281,
	"pageSize": 1000,
	"items": [
		"0xff00000000000000000000000000000000000000000000000000000000000000000000000000",
		"0xfa00000000000000000000000000000000000000000000000000000000000000000000000000"
	],
	"cursor": "cursor-value"
}`,
		},
		{
			Name: "ok - IndexerResponse - omitempty",
			Source: &api.IndexerResponse{
				CommittedSlot: 281,
				PageSize:      1000,
				Items: iotago.HexOutputIDsFromOutputIDs(
					iotago.OutputID{0xff},
					iotago.OutputID{0xfa},
				),
			},
			Target: `{
	"committedSlot": 281,
	"pageSize": 1000,
	"items": [
		"0xff00000000000000000000000000000000000000000000000000000000000000000000000000",
		"0xfa00000000000000000000000000000000000000000000000000000000000000000000000000"
	]
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}
