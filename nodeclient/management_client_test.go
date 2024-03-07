package nodeclient_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
	"github.com/iotaledger/iota.go/v4/nodeclient"
)

var sampleGossipInfo = &api.GossipInfo{
	Heartbeat: &api.GossipHeartbeat{
		SolidSlot:      234,
		PrunedSlot:     5872,
		LatestSlot:     1294,
		ConnectedPeers: 2392,
		SyncedPeers:    1234,
	},
	Metrics: &api.PeerGossipMetrics{
		NewBlocks:             40,
		KnownBlocks:           60,
		ReceivedBlocks:        100,
		ReceivedBlockRequests: 345,
		ReceivedSlotRequests:  194,
		ReceivedHeartbeats:    5,
		SentBlocks:            492,
		SentBlockRequests:     2396,
		SentSlotRequests:      9837,
		SentHeartbeats:        3,
		DroppedPackets:        10,
	},
}

func TestManagementClient_Enabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{api.ManagementPluginName},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)

	client := nodeClient(t)

	_, err := client.Management(context.TODO())
	require.NoError(t, err)
}

func TestManagementClient_Disabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{"someplugin/v1"},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)

	client := nodeClient(t)

	_, err := client.Management(context.TODO())
	require.Error(t, err, nodeclient.ErrManagementPluginNotAvailable)
}

func TestManagementClient_PeerByID(t *testing.T) {
	defer gock.Off()

	originRes := &api.PeerInfo{
		MultiAddresses: []iotago.PrefixedStringUint8{iotago.PrefixedStringUint8(fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID))},
		ID:             peerID,
		Connected:      true,
		Relation:       "autopeered",
		Gossip:         sampleGossipInfo,
	}

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{api.ManagementPluginName},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)
	mockGetJSON(api.EndpointWithNamedParameterValue(api.ManagementRoutePeer, api.ParameterPeerID, peerID), 200, originRes)

	client := nodeClient(t)

	management, err := client.Management(context.TODO())
	require.NoError(t, err)

	resp, err := management.PeerByID(context.Background(), peerID)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestManagementClient_RemovePeerByID(t *testing.T) {
	defer gock.Off()

	gock.New(nodeAPIUrl).
		Delete(api.EndpointWithNamedParameterValue(api.ManagementRoutePeer, api.ParameterPeerID, peerID)).
		Reply(200).
		Status(200)

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{api.ManagementPluginName},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)

	client := nodeClient(t)

	management, err := client.Management(context.TODO())
	require.NoError(t, err)

	err = management.RemovePeerByID(context.Background(), peerID)
	require.NoError(t, err)
}

func TestManagementClient_Peers(t *testing.T) {
	defer gock.Off()

	peerID2 := "12D3KooWFJ8Nq6gHLLvigTpPdddddsadsadscpJof8Y4y8yFAB32"

	originRes := &api.PeersResponse{
		Peers: []*api.PeerInfo{
			{
				ID:             peerID,
				MultiAddresses: []iotago.PrefixedStringUint8{iotago.PrefixedStringUint8(fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID))},
				Relation:       "autopeered",
				Gossip:         sampleGossipInfo,
				Connected:      true,
			},
			{
				ID:             peerID2,
				MultiAddresses: []iotago.PrefixedStringUint8{iotago.PrefixedStringUint8(fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID2))},
				Alias:          "Peer2",
				Relation:       "static",
				Gossip:         sampleGossipInfo,
				Connected:      true,
			},
		},
	}

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{api.ManagementPluginName},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)
	mockGetJSON(api.ManagementRoutePeers, 200, originRes)

	client := nodeClient(t)

	management, err := client.Management(context.TODO())
	require.NoError(t, err)

	resp, err := management.Peers(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}

func TestManagementClient_AddPeer(t *testing.T) {
	defer gock.Off()

	multiAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID)

	originRes := &api.PeerInfo{
		ID:             peerID,
		MultiAddresses: []iotago.PrefixedStringUint8{iotago.PrefixedStringUint8(multiAddr)},
		Relation:       "autopeered",
		Connected:      true,
		Gossip:         sampleGossipInfo,
	}

	req := &api.AddPeerRequest{MultiAddress: multiAddr}

	originRoutes := &api.RoutesResponse{
		Routes: []iotago.PrefixedStringUint8{api.ManagementPluginName},
	}

	mockGetJSON(api.RouteRoutes, 200, originRoutes)
	mockPostJSON(api.ManagementRoutePeers, 201, req, originRes)

	client := nodeClient(t)

	management, err := client.Management(context.TODO())
	require.NoError(t, err)

	resp, err := management.AddPeer(context.Background(), multiAddr)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}
