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
				Gossip: &api.GossipInfo{
					Heartbeat: &api.GossipHeartbeat{
						SolidSlot:      1,
						PrunedSlot:     2,
						LatestSlot:     3,
						ConnectedPeers: 4,
						SyncedPeers:    5,
					},
					Metrics: &api.PeerGossipMetrics{
						NewBlocks:             1,
						KnownBlocks:           2,
						ReceivedBlocks:        3,
						ReceivedBlockRequests: 4,
						ReceivedSlotRequests:  5,
						ReceivedHeartbeats:    6,
						SentBlocks:            7,
						SentBlockRequests:     8,
						SentSlotRequests:      9,
						SentHeartbeats:        10,
						DroppedPackets:        11,
					},
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
						Gossip: &api.GossipInfo{
							Heartbeat: &api.GossipHeartbeat{
								SolidSlot:      1,
								PrunedSlot:     2,
								LatestSlot:     3,
								ConnectedPeers: 4,
								SyncedPeers:    5,
							},
							Metrics: &api.PeerGossipMetrics{
								NewBlocks:             1,
								KnownBlocks:           2,
								ReceivedBlocks:        3,
								ReceivedBlockRequests: 4,
								ReceivedSlotRequests:  5,
								ReceivedHeartbeats:    6,
								SentBlocks:            7,
								SentBlockRequests:     8,
								SentSlotRequests:      9,
								SentHeartbeats:        10,
								DroppedPackets:        11,
							},
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
		{
			Name: "ok - CreateSnapshotsRequest",
			Source: &api.CreateSnapshotsRequest{
				Slot: 1,
			},
			Target: &api.CreateSnapshotsRequest{},
		},
		{
			Name: "ok - CreateSnapshotsResponse",
			Source: &api.CreateSnapshotsResponse{
				Slot:     1,
				FilePath: "filePath",
			},
			Target: &api.CreateSnapshotsResponse{},
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
				Gossip: &api.GossipInfo{
					Heartbeat: &api.GossipHeartbeat{
						SolidSlot:      1,
						PrunedSlot:     2,
						LatestSlot:     3,
						ConnectedPeers: 4,
						SyncedPeers:    5,
					},
					Metrics: &api.PeerGossipMetrics{
						NewBlocks:             1,
						KnownBlocks:           2,
						ReceivedBlocks:        3,
						ReceivedBlockRequests: 4,
						ReceivedSlotRequests:  5,
						ReceivedHeartbeats:    6,
						SentBlocks:            7,
						SentBlockRequests:     8,
						SentSlotRequests:      9,
						SentHeartbeats:        10,
						DroppedPackets:        11,
					},
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
	"gossip": {
		"heartbeat": {
			"solidSlot": 1,
			"prunedSlot": 2,
			"latestSlot": 3,
			"connectedPeers": 4,
			"syncedPeers": 5
		},
		"metrics": {
			"newBlocks": 1,
			"knownBlocks": 2,
			"receivedBlocks": 3,
			"receivedBlockRequests": 4,
			"receivedSlotRequests": 5,
			"receivedHeartbeats": 6,
			"sentBlocks": 7,
			"sentBlockRequests": 8,
			"sentSlotRequests": 9,
			"sentHeartbeats": 10,
			"droppedPackets": 11
		}
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
						Gossip: &api.GossipInfo{
							Heartbeat: &api.GossipHeartbeat{
								SolidSlot:      1,
								PrunedSlot:     2,
								LatestSlot:     3,
								ConnectedPeers: 4,
								SyncedPeers:    5,
							},
							Metrics: &api.PeerGossipMetrics{
								NewBlocks:             1,
								KnownBlocks:           2,
								ReceivedBlocks:        3,
								ReceivedBlockRequests: 4,
								ReceivedSlotRequests:  5,
								ReceivedHeartbeats:    6,
								SentBlocks:            7,
								SentBlockRequests:     8,
								SentSlotRequests:      9,
								SentHeartbeats:        10,
								DroppedPackets:        11,
							},
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
			"gossip": {
				"heartbeat": {
					"solidSlot": 1,
					"prunedSlot": 2,
					"latestSlot": 3,
					"connectedPeers": 4,
					"syncedPeers": 5
				},
				"metrics": {
					"newBlocks": 1,
					"knownBlocks": 2,
					"receivedBlocks": 3,
					"receivedBlockRequests": 4,
					"receivedSlotRequests": 5,
					"receivedHeartbeats": 6,
					"sentBlocks": 7,
					"sentBlockRequests": 8,
					"sentSlotRequests": 9,
					"sentHeartbeats": 10,
					"droppedPackets": 11
				}
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
		{
			Name: "ok - CreateSnapshotsRequest",
			Source: &api.CreateSnapshotsRequest{
				Slot: 1,
			},
			Target: `{
	"slot": 1
}`,
		},
		{
			Name: "ok - CreateSnapshotsResponse",
			Source: &api.CreateSnapshotsResponse{
				Slot:     1,
				FilePath: "filePath",
			},
			Target: `{
	"slot": 1,
	"filePath": "filePath"
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, tt.Run)
	}
}
