package nodeclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/iotaledger/iota.go/v4/nodeclient/apimodels"
)

const (
	// ManagementEndpointPeer is the endpoint for getting peers by their peerID.
	// GET returns the peer.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	// DELETE deletes the peer.
	ManagementEndpointPeer = "/peers/%s"

	// ManagementEndpointPeers is the endpoint for getting all peers of the node.
	// GET returns a list of all peers.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	// POST adds a new peer.
	// "Content-Type" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	ManagementEndpointPeers = "/peers"

	// ManagementEndpointDatabasePrune is the endpoint to manually prune the database.
	// POST prunes the database.
	// "Content-Type" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	ManagementEndpointDatabasePrune = "/database/prune"

	// ManagementEndpointSnapshotsCreate is the endpoint to manually create a snapshot files.
	// POST creates a full snapshot.
	// "Content-Type" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	ManagementEndpointSnapshotsCreate = "/snapshots/create"
)

var (
	ManagementRoutePeer            = route(ManagementPluginName, ManagementEndpointPeer)
	ManagementRoutePeers           = route(ManagementPluginName, ManagementEndpointPeers)
	ManagementRouteDatabasePrune   = route(ManagementPluginName, ManagementEndpointDatabasePrune)
	ManagementRouteSnapshotsCreate = route(ManagementPluginName, ManagementEndpointSnapshotsCreate)
)

type (
	// ManagementClient is a client which queries the optional management functionality of a node.
	ManagementClient interface {
		// PeerByID gets a peer by its identifier.
		PeerByID(ctx context.Context, id string) (*apimodels.PeerInfo, error)
		// RemovePeerByID removes a peer by its identifier.
		RemovePeerByID(ctx context.Context, id string) error
		// Peers returns a list of all peers.
		Peers(ctx context.Context) (*apimodels.PeersResponse, error)
		// AddPeer adds a new peer by libp2p multi address with optional alias.
		AddPeer(ctx context.Context, multiAddress string, alias ...string) (*apimodels.PeerInfo, error)
	}

	managementClient struct {
		core *Client
	}
)

// Do executes a request against the endpoint.
// This function is only meant to be used for special routes not covered through the standard API.
func (client *managementClient) Do(ctx context.Context, method string, route string, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	return client.core.Do(ctx, method, route, reqObj, resObj)
}

// DoWithRequestHeaderHook executes a request against the endpoint.
// This function is only meant to be used for special routes not covered through the standard API.
func (client *managementClient) DoWithRequestHeaderHook(ctx context.Context, method string, route string, requestHeaderHook RequestHeaderHook, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	return client.core.DoWithRequestHeaderHook(ctx, method, route, requestHeaderHook, reqObj, resObj)
}

// PeerByID gets a peer by its identifier.
func (client *managementClient) PeerByID(ctx context.Context, id string) (*apimodels.PeerInfo, error) {
	query := fmt.Sprintf(ManagementRoutePeer, id)

	res := new(apimodels.PeerInfo)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// RemovePeerByID removes a peer by its identifier.
func (client *managementClient) RemovePeerByID(ctx context.Context, id string) error {
	query := fmt.Sprintf(ManagementRoutePeer, id)

	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodDelete, query, nil, nil); err != nil {
		return err
	}

	return nil
}

// Peers returns a list of all peers.
func (client *managementClient) Peers(ctx context.Context) (*apimodels.PeersResponse, error) {
	res := new(apimodels.PeersResponse)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, ManagementRoutePeers, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// AddPeer adds a new peer by libp2p multi address with optional alias.
func (client *managementClient) AddPeer(ctx context.Context, multiAddress string, alias ...string) (*apimodels.PeerInfo, error) {
	req := &apimodels.AddPeerRequest{
		MultiAddress: multiAddress,
	}

	if len(alias) > 0 {
		req.Alias = alias[0]
	}

	res := new(apimodels.PeerInfo)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodPost, ManagementRoutePeers, req, res); err != nil {
		return nil, err
	}

	return res, nil
}
