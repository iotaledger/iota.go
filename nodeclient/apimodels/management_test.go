package apimodels_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/iota.go/v4/nodeclient/apimodels"
)

func Test_PeersResponse(t *testing.T) {
	api := testAPI()
	{
		peerID := "12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32"
		peerID2 := "12D3KooWFJ8Nq6gHLLvigTpPdddddsadsadscpJof8Y4y8yFAB32"

		response := &apimodels.PeersResponse{
			Peers: []*apimodels.PeerInfo{
				{
					ID:             peerID,
					MultiAddresses: []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID)},
					Relation:       "autopeered",
					Gossip: &apimodels.GossipInfo{
						Heartbeat: &apimodels.GossipHeartbeat{
							SolidSlotIndex:  1,
							PrunedSlotIndex: 2,
							LatestSlotIndex: 3,
							ConnectedPeers:  4,
							SyncedPeers:     5,
						},
						Metrics: &apimodels.PeerGossipMetrics{
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
					Connected: true,
				},
				{
					ID:             peerID2,
					MultiAddresses: []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID2)},
					Alias:          "Peer2",
					Relation:       "static",
					Connected:      false,
				},
			},
		}

		jsonResponse, err := api.JSONEncode(response)
		require.NoError(t, err)

		expected := "{\"peers\":[{\"id\":\"12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32\",\"multiAddresses\":[\"/ip4/127.0.0.1/tcp/15600/p2p/12D3KooWFJ8Nq6gHLLvigTpPSbyMmLk35k1TcpJof8Y4y8yFAB32\"],\"relation\":\"autopeered\",\"connected\":true,\"gossip\":{\"heartbeat\":{\"solidSlotIndex\":\"1\",\"prunedSlotIndex\":\"2\",\"latestSlotIndex\":\"3\",\"connectedPeers\":4,\"syncedPeers\":5},\"metrics\":{\"newBlocks\":1,\"knownBlocks\":2,\"receivedBlocks\":3,\"receivedBlockRequests\":4,\"receivedSlotRequests\":5,\"receivedHeartbeats\":6,\"sentBlocks\":7,\"sentBlockRequests\":8,\"sentSlotRequests\":9,\"sentHeartbeats\":10,\"droppedPackets\":11}}},{\"id\":\"12D3KooWFJ8Nq6gHLLvigTpPdddddsadsadscpJof8Y4y8yFAB32\",\"multiAddresses\":[\"/ip4/127.0.0.1/tcp/15600/p2p/12D3KooWFJ8Nq6gHLLvigTpPdddddsadsadscpJof8Y4y8yFAB32\"],\"alias\":\"Peer2\",\"relation\":\"static\",\"connected\":false}]}"
		require.Equal(t, expected, string(jsonResponse))

		decoded := new(apimodels.PeersResponse)
		require.NoError(t, api.JSONDecode(jsonResponse, decoded))
		require.EqualValues(t, response, decoded)
	}
}
