package nodeclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
)

const (
	IndexerPluginName = "indexer/v1"
	MQTTPluginName    = "mqtt/v1"

	// NodeAPIRouteHealth is the route for querying a node's health status.
	NodeAPIRouteHealth = "/health"

	// NodeAPIRouteInfo is the route for getting the node info.
	// GET returns the node info.
	NodeAPIRouteInfo = "/api/v2/info"

	// NodeAPIRouteTips is the route for getting two tips.
	// GET returns the tips.
	NodeAPIRouteTips = "/api/v2/tips"

	// NodeAPIRouteMessageMetadata is the route for getting message metadata by its messageID.
	// GET returns message metadata (including info about "promotion/reattachment needed").
	NodeAPIRouteMessageMetadata = "/api/v2/messages/%s/metadata"

	// NodeAPIRouteMessageBytes is the route for getting message raw data by its messageID.
	// GET returns raw message data (bytes).
	NodeAPIRouteMessageBytes = "/api/v2/messages/%s/raw"

	// NodeAPIRouteMessageChildren is the route for getting message IDs of the children of a message, identified by its messageID.
	// GET returns the message IDs of all children.
	NodeAPIRouteMessageChildren = "/api/v2/messages/%s/children"

	// NodeAPIRouteMessages is the route for getting message IDs or creating new messages.
	// GET with query parameter (mandatory) returns all message IDs that fit these filter criteria (query parameters: "index").
	// POST creates a single new message and returns the new message ID.
	NodeAPIRouteMessages = "/api/v2/messages"

	// NodeAPIRouteMilestone is the route for getting a milestone by its milestoneIndex.
	// GET returns the milestone.
	NodeAPIRouteMilestone = "/api/v2/milestones/%d"

	// NodeAPIRouteMilestoneUTXOChanges is the route for getting all UTXO changes of a milestone by its milestoneIndex.
	// GET returns the output IDs of all UTXO changes.
	NodeAPIRouteMilestoneUTXOChanges = "/api/v2/milestones/%d/utxo-changes"

	// NodeAPIRouteOutput is the route for getting outputs by their outputID (transactionHash + outputIndex).
	// GET returns the output.
	NodeAPIRouteOutput = "/api/v2/outputs/%s"

	// NodeAPIRouteTreasury is the route for getting the current treasury.
	// GET returns the treasury.
	NodeAPIRouteTreasury = "/api/v2/treasury"

	// NodeAPIRouteTxIncludedMessage is the route for getting the included message of a transaction.
	// GET returns the message.
	NodeAPIRouteTxIncludedMessage = "/api/v2/transactions/%s/included-message"

	// NodeAPIRouteReceipts is the route for getting all persisted receipts on a node.
	// GET returns the receipts.
	NodeAPIRouteReceipts = "/api/v2/receipts"

	// NodeAPIRouteReceiptsByMigratedAtIndex is the route for getting all persisted receipts for a given migrated at index on a node.
	// GET returns the receipts for the given migrated at index.
	NodeAPIRouteReceiptsByMigratedAtIndex = "/api/v2/receipts/%d"

	// NodeAPIRoutePeer is the route for getting peers by their peerID.
	// GET returns the peer
	// DELETE deletes the peer.
	NodeAPIRoutePeer = "/api/v2/peers/%s"

	// NodeAPIRoutePeers is the route for getting all peers of the node.
	// GET returns a list of all peers.
	// POST adds a new peer.
	NodeAPIRoutePeers = "/api/v2/peers"
)

var (
	ErrIndexerPluginNotAvailable = errors.New("indexer plugin not available on the current node")
	ErrMQTTPluginNotAvailable    = errors.New("mqtt plugin not available on the current node")
)

// the default options applied to the Client.
var defaultNodeAPIOptions = []ClientOption{
	WithHTTPClient(http.DefaultClient),
	WithUserInfo(nil),
}

// ClientOptions define options for the Client.
type ClientOptions struct {
	// The HTTP client to use.
	httpClient *http.Client
	// The username and password information.
	userInfo *url.Userinfo
}

// applies the given ClientOption.
func (no *ClientOptions) apply(opts ...ClientOption) {
	for _, opt := range opts {
		opt(no)
	}
}

// WithHTTPClient sets the used HTTP Client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(opts *ClientOptions) {
		opts.httpClient = httpClient
	}
}

// WithUserInfo sets the Userinfo used to add basic auth "Authorization" headers to the requests.
func WithUserInfo(userInfo *url.Userinfo) ClientOption {
	return func(opts *ClientOptions) {
		opts.userInfo = userInfo
	}
}

// ClientOption is a function setting a Client option.
type ClientOption func(opts *ClientOptions)

// New returns a new Client using the given base URL.
func New(baseURL string, opts ...ClientOption) *Client {

	options := &ClientOptions{}
	options.apply(defaultNodeAPIOptions...)
	options.apply(opts...)

	client := &Client{
		BaseURL: baseURL,
		opts:    options,
	}

	return client
}

// Client is a client for node HTTP REST API endpoints.
type Client struct {
	// The base URL for all API calls.
	BaseURL string
	// holds the Client options.
	opts *ClientOptions
}

// HTTPErrorResponseEnvelope defines the error response schema for node API responses.
type HTTPErrorResponseEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// RawDataEnvelope is used internally to encapsulate binary data.
type RawDataEnvelope struct {
	// The encapsulated binary data.
	Data []byte
}

// Do executes a request against the endpoint.
// This function is only meant to be used for special routes not covered through the standard API.
func (client *Client) Do(ctx context.Context, method string, route string, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	return do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, method, route, reqObj, resObj)
}

// Indexer returns the IndexerClient.
// Returns ErrIndexerPluginNotAvailable if the current node does not support the plugin.
func (client *Client) Indexer(ctx context.Context) (IndexerClient, error) {
	hasPlugin, err := client.NodeSupportPlugin(ctx, IndexerPluginName)
	if err != nil {
		return nil, err
	}
	if !hasPlugin {
		return nil, ErrIndexerPluginNotAvailable
	}
	return &indexerClient{core: client}, nil
}

// EventAPI returns the EventAPIClient if supported by the node.
// Returns ErrMQTTPluginNotAvailable if the current node does not support the plugin.
func (client *Client) EventAPI(ctx context.Context) (*EventAPIClient, error) {
	hasPlugin, err := client.NodeSupportPlugin(ctx, MQTTPluginName)
	if err != nil {
		return nil, err
	}
	if !hasPlugin {
		return nil, ErrMQTTPluginNotAvailable
	}
	return newEventAPIClient(client), nil
}

// Health returns whether the given node is healthy.
func (client *Client) Health(ctx context.Context) (bool, error) {
	res, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodGet, NodeAPIRouteHealth, nil, nil)
	if err != nil {
		return false, err
	}
	if res.StatusCode == http.StatusServiceUnavailable {
		return false, nil
	}
	return true, nil
}

// Info gets the info of the node.
func (client *Client) Info(ctx context.Context) (*InfoResponse, error) {
	res := &InfoResponse{}
	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodGet, NodeAPIRouteInfo, nil, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// NodeSupportPlugin gets the info of the node and checks if the given plugin is enabled.
func (client *Client) NodeSupportPlugin(ctx context.Context, pluginName string) (bool, error) {
	info, err := client.Info(ctx)
	if err != nil {
		return false, err
	}
	for _, p := range info.Plugins {
		if p == pluginName {
			return true, nil
		}
	}
	return false, nil
}

// NodeTipsResponse defines the response of a GET tips REST API call.
type NodeTipsResponse struct {
	// The hex encoded message IDs of the tips.
	TipsHex []string `json:"tipMessageIds"`
}

// Tips returns the hex encoded tips as MessageIDs.
func (ntr *NodeTipsResponse) Tips() (iotago.MessageIDs, error) {
	msgIDs := make(iotago.MessageIDs, len(ntr.TipsHex))
	for i, tip := range ntr.TipsHex {
		msgID, err := iotago.DecodeHex(tip)
		if err != nil {
			return nil, err
		}
		copy(msgIDs[i][:], msgID)
	}
	return msgIDs, nil
}

// Tips gets the two tips from the node.
func (client *Client) Tips(ctx context.Context) (*NodeTipsResponse, error) {
	res := &NodeTipsResponse{}
	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodGet, NodeAPIRouteTips, nil, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// SubmitMessage submits the given Message to the node.
// The node will take care of filling missing information.
// This function returns the finalized message created by the node.
func (client *Client) SubmitMessage(ctx context.Context, m *iotago.Message, deSeriParas *iotago.DeSerializationParameters) (*iotago.Message, error) {
	// do not check the message because the validation would fail if
	// no parents were given. The node will first add this missing information and
	// validate the message afterwards.
	data, err := m.Serialize(serializer.DeSeriModeNoValidation, deSeriParas)
	if err != nil {
		return nil, err
	}

	req := &RawDataEnvelope{Data: data}
	res, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodPost, NodeAPIRouteMessages, req, nil)
	if err != nil {
		return nil, err
	}

	messageID, err := iotago.MessageIDFromHexString(res.Header.Get(locationHeader))
	if err != nil {
		return nil, err
	}

	msg, err := client.MessageByMessageID(ctx, messageID, deSeriParas)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// MessageMetadataByMessageID gets the metadata of a message by its message ID from the node.
func (client *Client) MessageMetadataByMessageID(ctx context.Context, msgID iotago.MessageID) (*MessageMetadataResponse, error) {
	query := fmt.Sprintf(NodeAPIRouteMessageMetadata, iotago.EncodeHex(msgID[:]))

	res := &MessageMetadataResponse{}
	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// MessageByMessageID get a message by its message ID from the node.
func (client *Client) MessageByMessageID(ctx context.Context, msgID iotago.MessageID, deSeriParas *iotago.DeSerializationParameters) (*iotago.Message, error) {
	query := fmt.Sprintf(NodeAPIRouteMessageBytes, iotago.EncodeHex(msgID[:]))

	res := &RawDataEnvelope{}
	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	msg := &iotago.Message{}
	if _, err = msg.Deserialize(res.Data, serializer.DeSeriModePerformValidation, deSeriParas); err != nil {
		return nil, err
	}
	return msg, nil
}

// ChildrenByMessageID gets the MessageIDs of the child messages of a given message.
func (client *Client) ChildrenByMessageID(ctx context.Context, parentMsgID iotago.MessageID) (*ChildrenResponse, error) {
	query := fmt.Sprintf(NodeAPIRouteMessageChildren, iotago.EncodeHex(parentMsgID[:]))

	res := &ChildrenResponse{}
	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// TransactionIncludedMessage get a message that included the given transaction ID in the ledger.
func (client *Client) TransactionIncludedMessage(ctx context.Context, txID iotago.TransactionID, deSeriParas *iotago.DeSerializationParameters) (*iotago.Message, error) {
	query := fmt.Sprintf(NodeAPIRouteTxIncludedMessage, iotago.EncodeHex(txID[:]))

	res := &RawDataEnvelope{}
	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	msg := &iotago.Message{}
	if _, err = msg.Deserialize(res.Data, serializer.DeSeriModePerformValidation, deSeriParas); err != nil {
		return nil, err
	}
	return msg, nil
}

// OutputByID gets an outputs by its ID from the node.
func (client *Client) OutputByID(ctx context.Context, outputID iotago.OutputID) (*OutputResponse, error) {
	query := fmt.Sprintf(NodeAPIRouteOutput, outputID.ToHex())

	res := &OutputResponse{}
	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Treasury gets the current treasury.
func (client *Client) Treasury(ctx context.Context) (*TreasuryResponse, error) {
	res := &TreasuryResponse{}
	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodGet, NodeAPIRouteTreasury, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Receipts gets all receipts persisted on the node.
func (client *Client) Receipts(ctx context.Context) ([]*ReceiptTuple, error) {
	res := &ReceiptsResponse{}
	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodGet, NodeAPIRouteReceipts, nil, res)
	if err != nil {
		return nil, err
	}

	return res.Receipts, nil
}

// ReceiptsByMigratedAtIndex gets all receipts for the given migrated at index persisted on the node.
func (client *Client) ReceiptsByMigratedAtIndex(ctx context.Context, index uint32) ([]*ReceiptTuple, error) {
	query := fmt.Sprintf(NodeAPIRouteReceiptsByMigratedAtIndex, index)

	res := &ReceiptsResponse{}
	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res.Receipts, nil
}

// MilestoneByIndex gets a milestone by its index.
func (client *Client) MilestoneByIndex(ctx context.Context, index uint32) (*MilestoneResponse, error) {
	query := fmt.Sprintf(NodeAPIRouteMilestone, index)

	res := &MilestoneResponse{}
	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// MilestoneUTXOChangesByIndex returns all UTXO changes of a milestone by its milestoneIndex.
func (client *Client) MilestoneUTXOChangesByIndex(ctx context.Context, index uint32) (*MilestoneUTXOChangesResponse, error) {
	query := fmt.Sprintf(NodeAPIRouteMilestoneUTXOChanges, index)

	res := &MilestoneUTXOChangesResponse{}
	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// PeerByID gets a peer by its identifier.
func (client *Client) PeerByID(ctx context.Context, id string) (*PeerResponse, error) {
	query := fmt.Sprintf(NodeAPIRoutePeer, id)

	res := &PeerResponse{}
	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// RemovePeerByID removes a peer by its identifier.
func (client *Client) RemovePeerByID(ctx context.Context, id string) error {
	query := fmt.Sprintf(NodeAPIRoutePeer, id)

	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodDelete, query, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// Peers returns a list of all peers.
func (client *Client) Peers(ctx context.Context) ([]*PeerResponse, error) {
	res := []*PeerResponse{}
	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodGet, NodeAPIRoutePeers, nil, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// AddPeer adds a new peer by libp2p multi address with optional alias.
func (client *Client) AddPeer(ctx context.Context, multiAddress string, alias ...string) (*PeerResponse, error) {
	req := &AddPeerRequest{
		MultiAddress: multiAddress,
	}

	if len(alias) > 0 {
		req.Alias = &alias[0]
	}

	res := &PeerResponse{}
	_, err := do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, http.MethodPost, NodeAPIRoutePeers, req, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
