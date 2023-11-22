package nodeclient_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	"github.com/iotaledger/iota.go/v4/nodeclient"
	"github.com/iotaledger/iota.go/v4/nodeclient/apimodels"
)

func TestManagementClient_Enabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &apimodels.RoutesResponse{
		Routes: []string{nodeclient.ManagementPluginName},
	}

	mockGetJSON(nodeclient.RouteRoutes, 200, originRoutes)

	client := nodeClient(t)

	_, err := client.Management(context.TODO())
	require.NoError(t, err)
}

func TestManagementClient_Disabled(t *testing.T) {
	defer gock.Off()

	originRoutes := &apimodels.RoutesResponse{
		Routes: []string{"someplugin/v1"},
	}

	mockGetJSON(nodeclient.RouteRoutes, 200, originRoutes)

	client := nodeClient(t)

	_, err := client.Management(context.TODO())
	require.Error(t, err, nodeclient.ErrManagementPluginNotAvailable)
}

func TestManagementClient_PeerByID(t *testing.T) {
	defer gock.Off()

	originRes := &apimodels.PeerInfo{
		MultiAddresses: []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID)},
		ID:             peerID,
		Connected:      true,
		Relation:       "autopeered",
		Gossip:         sampleGossipInfo,
	}

	originRoutes := &apimodels.RoutesResponse{
		Routes: []string{nodeclient.ManagementPluginName},
	}

	mockGetJSON(nodeclient.RouteRoutes, 200, originRoutes)
	mockGetJSON(fmt.Sprintf(nodeclient.ManagementRoutePeer, peerID), 200, originRes)

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
		Delete(fmt.Sprintf(nodeclient.ManagementRoutePeer, peerID)).
		Reply(200).
		Status(200)

	originRoutes := &apimodels.RoutesResponse{
		Routes: []string{nodeclient.ManagementPluginName},
	}

	mockGetJSON(nodeclient.RouteRoutes, 200, originRoutes)

	client := nodeClient(t)

	management, err := client.Management(context.TODO())
	require.NoError(t, err)

	err = management.RemovePeerByID(context.Background(), peerID)
	require.NoError(t, err)
}

func TestManagementClient_Peers(t *testing.T) {
	defer gock.Off()

	peerID2 := "12D3KooWFJ8Nq6gHLLvigTpPdddddsadsadscpJof8Y4y8yFAB32"

	originRes := &apimodels.PeersResponse{
		Peers: []*apimodels.PeerInfo{
			{
				ID:             peerID,
				MultiAddresses: []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID)},
				Relation:       "autopeered",
				Gossip:         sampleGossipInfo,
				Connected:      true,
			},
			{
				ID:             peerID2,
				MultiAddresses: []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/15600/p2p/%s", peerID2)},
				Alias:          "Peer2",
				Relation:       "static",
				Gossip:         sampleGossipInfo,
				Connected:      true,
			},
		},
	}

	originRoutes := &apimodels.RoutesResponse{
		Routes: []string{nodeclient.ManagementPluginName},
	}

	mockGetJSON(nodeclient.RouteRoutes, 200, originRoutes)
	mockGetJSON(nodeclient.ManagementRoutePeers, 200, originRes)

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

	originRes := &apimodels.PeerInfo{
		ID:             peerID,
		MultiAddresses: []string{multiAddr},
		Relation:       "autopeered",
		Connected:      true,
		Gossip:         sampleGossipInfo,
	}

	req := &apimodels.AddPeerRequest{MultiAddress: multiAddr}

	originRoutes := &apimodels.RoutesResponse{
		Routes: []string{nodeclient.ManagementPluginName},
	}

	mockGetJSON(nodeclient.RouteRoutes, 200, originRoutes)
	mockPostJSON(nodeclient.ManagementRoutePeers, 201, req, originRes)

	client := nodeClient(t)

	management, err := client.Management(context.TODO())
	require.NoError(t, err)

	resp, err := management.AddPeer(context.Background(), multiAddr)
	require.NoError(t, err)
	require.EqualValues(t, originRes, resp)
}
