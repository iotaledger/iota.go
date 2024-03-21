package api_test

import (
	"testing"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
	"github.com/iotaledger/iota.go/v4/tpkg/frameworks"
)

func Test_ManagementAPIDeSerialize(t *testing.T) {
	tests := []*frameworks.DeSerializeTest{
		{
			Name: "ok - AddPeerRequest",
			Source: &api.AddPeerRequest{
				MultiAddress: "multiAddress",
				Alias:        "alias",
			},
			Target: &api.AddPeerRequest{},
		},
		{
			Name: "ok - PeerInfo",
			Source: &api.PeerInfo{
				ID:             "id",
				MultiAddresses: []iotago.PrefixedStringUint8{"multiAddress"},
				Alias:          "alias",
				Relation:       "relation",
				Connected:      true,
				GossipMetrics: &api.PeerGossipMetrics{
					PacketsReceived: 1,
					PacketsSent:     2,
				},
			},
			Target: &api.PeerInfo{},
		},
		{
			Name: "ok - PeersResponse",
			Source: &api.PeersResponse{
				Peers: []*api.PeerInfo{
					{
						ID:             "id",
						MultiAddresses: []iotago.PrefixedStringUint8{"multiAddress"},
						Alias:          "alias",
						Relation:       "relation",
						Connected:      true,
						GossipMetrics: &api.PeerGossipMetrics{
							PacketsReceived: 1,
							PacketsSent:     2,
						},
					},
				},
			},
			Target: &api.PeersResponse{},
		},
		{
			Name: "ok - PruneDatabaseRequest",
			Source: &api.PruneDatabaseRequest{
				Epoch:              1,
				Depth:              2,
				TargetDatabaseSize: "targetDatabaseSize",
			},
			Target: &api.PruneDatabaseRequest{},
		},
		{
			Name: "ok - PruneDatabaseResponse",
			Source: &api.PruneDatabaseResponse{
				Epoch: 1,
			},
			Target: &api.PruneDatabaseResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}

func Test_ManagementAPIJSONSerialization(t *testing.T) {
	tests := []*frameworks.JSONEncodeTest{
		{
			Name: "ok - AddPeerRequest",
			Source: &api.AddPeerRequest{
				MultiAddress: "multiAddress",
				Alias:        "alias",
			},
			Target: `{
	"multiAddress": "multiAddress",
	"alias": "alias"
}`,
		},
		{
			Name: "ok - PeerInfo",
			Source: &api.PeerInfo{
				ID:             "id",
				MultiAddresses: []iotago.PrefixedStringUint8{"multiAddress"},
				Alias:          "alias",
				Relation:       "relation",
				Connected:      true,
				GossipMetrics: &api.PeerGossipMetrics{
					PacketsReceived: 1,
					PacketsSent:     2,
				},
			},
			Target: `{
	"id": "id",
	"multiAddresses": [
		"multiAddress"
	],
	"alias": "alias",
	"relation": "relation",
	"connected": true,
	"gossipMetrics": {
		"packetsReceived": 1,
		"packetsSent": 2
	}
}`,
		},
		{
			Name: "ok - PeersResponse",
			Source: &api.PeersResponse{
				Peers: []*api.PeerInfo{
					{
						ID:             "id",
						MultiAddresses: []iotago.PrefixedStringUint8{"multiAddress"},
						Alias:          "alias",
						Relation:       "relation",
						Connected:      true,
						GossipMetrics: &api.PeerGossipMetrics{
							PacketsReceived: 1,
							PacketsSent:     2,
						},
					},
				},
			},
			Target: `{
	"peers": [
		{
			"id": "id",
			"multiAddresses": [
				"multiAddress"
			],
			"alias": "alias",
			"relation": "relation",
			"connected": true,
			"gossipMetrics": {
				"packetsReceived": 1,
				"packetsSent": 2
			}
		}
	]
}`,
		},
		{
			Name: "ok - PruneDatabaseRequest",
			Source: &api.PruneDatabaseRequest{
				Epoch:              1,
				Depth:              2,
				TargetDatabaseSize: "targetDatabaseSize",
			},
			Target: `{
	"epoch": 1,
	"depth": 2,
	"targetDatabaseSize": "targetDatabaseSize"
}`,
		},
		{
			Name: "ok - PruneDatabaseResponse",
			Source: &api.PruneDatabaseResponse{
				Epoch: 1,
			},
			Target: `{
	"epoch": 1
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}
