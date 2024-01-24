package api_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
	"github.com/iotaledger/iota.go/v4/tpkg/frameworks"
)

func Test_RoutesAPIDeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok - RoutesResponse",
			Source: &api.RoutesResponse{
				Routes: []iotago.PrefixedStringUint8{"route1", "route2"},
			},
			Target: &api.RoutesResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func Test_RoutesAPIJSONSerialization(t *testing.T) {
	tests := []*frameworks.JSONEncodeTest{
		{
			Name: "ok - RoutesResponse",
			Source: &api.RoutesResponse{
				Routes: []iotago.PrefixedStringUint8{"route1", "route2"},
			},
			Target: `{
	"routes": [
		"route1",
		"route2"
	]
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}
