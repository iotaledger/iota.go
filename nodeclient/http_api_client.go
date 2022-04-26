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
	// IndexerPluginName is the name for the indexer plugin.
	IndexerPluginName = "indexer/v1"

	// MQTTPluginName is the name for the MQTT plugin.
	MQTTPluginName = "mqtt/v1"
)

const (
	// RouteHealth is the route for querying a node's health status.
	RouteHealth = "/health"

	// RouteInfo is the route for getting the node info.
	// GET returns the node info.
	RouteInfo = "/api/v2/info"

	// RouteTips is the route for getting tips.
	// GET returns the tips.
	RouteTips = "/api/v2/tips"

	// RouteMessage is the route for getting a message by its messageID.
	// GET returns the message based on the given type in the request "Accept" header.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes
	RouteMessage = "/api/v2/messages/%s"

	// RouteMessageMetadata is the route for getting message metadata by its messageID.
	// GET returns message metadata (including info about "promotion/reattachment needed").
	RouteMessageMetadata = "/api/v2/messages/%s/metadata"

	// RouteMessageChildren is the route for getting message IDs of the children of a message, identified by its messageID.
	// GET returns the message IDs of all children.
	RouteMessageChildren = "/api/v2/messages/%s/children"

	// RouteMessages is the route for creating new messages.
	// POST creates a single new message and returns the new message ID.
	// The message is parsed based on the given type in the request "Content-Type" header.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes
	RouteMessages = "/api/v2/messages"

	// RouteTransactionsIncludedMessage is the route for getting the message that was included in the ledger for a given transaction ID.
	// GET returns the message based on the given type in the request "Accept" header.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes
	RouteTransactionsIncludedMessage = "/api/v2/transactions/%s/included-message"

	// RouteMilestoneByIndex is the route for getting a milestone by its milestoneIndex.
	// GET returns the milestone.
	RouteMilestoneByIndex = "/api/v2/milestones/index/%d"

	// RouteMilestone is the route for getting a milestone by its ID.
	// GET returns the milestone.
	RouteMilestone = "/api/v2/milestones/milestoneId/%s"

	// RouteMilestoneUTXOChanges is the route for getting all UTXO changes of a milestone by its milestoneIndex.
	// GET returns the output IDs of all UTXO changes.
	RouteMilestoneUTXOChanges = "/api/v2/milestones/%d/utxo-changes"

	// RouteOutput is the route for getting an output by its outputID (transactionHash + outputIndex).
	// GET returns the output based on the given type in the request "Accept" header.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes
	RouteOutput = "/api/v2/outputs/%s"

	// RouteOutputMetadata is the route for getting output metadata by its outputID (transactionHash + outputIndex) without getting the data again.
	// GET returns the output metadata.
	RouteOutputMetadata = "/api/v2/outputs/%s/metadata"

	// RouteTreasury is the route for getting the current treasury output.
	// GET returns the treasury.
	RouteTreasury = "/api/v2/treasury"

	// RouteReceipts is the route for getting all persisted receipts on a node.
	// GET returns the receipts.
	RouteReceipts = "/api/v2/receipts"

	// RouteReceiptsMigratedAtIndex is the route for getting all persisted receipts for a given migrated at index on a node.
	// GET returns the receipts for the given migrated at index.
	RouteReceiptsMigratedAtIndex = "/api/v2/receipts/%d"

	// RouteComputeWhiteFlagMutations is the route to compute the white flag mutations for the cone of the given parents.
	// POST computes the white flag mutations.
	RouteComputeWhiteFlagMutations = "/api/v2/whiteflag"

	// RoutePeer is the route for getting peers by their peerID.
	// GET returns the peer
	// DELETE deletes the peer.
	RoutePeer = "/api/v2/peers/%s"

	// RoutePeers is the route for getting all peers of the node.
	// GET returns a list of all peers.
	// POST adds a new peer.
	RoutePeers = "/api/v2/peers"
)

var (
	// ErrIndexerPluginNotAvailable is returned when the indexer plugin is not available on the node.
	ErrIndexerPluginNotAvailable = errors.New("indexer plugin not available on the current node")
	// ErrMQTTPluginNotAvailable is returned when the MQTT plugin is not available on the node.
	ErrMQTTPluginNotAvailable = errors.New("mqtt plugin not available on the current node")
)

// RequestURLHook is a function to modify the URL before sending a request.
type RequestURLHook func(url string) string

// RequestHeaderHook is a function to modify the request header before sending a request.
type RequestHeaderHook func(header http.Header)

var (
	// RequestHeaderHookAcceptJSON is used to set the request "Accept" header to MIMEApplicationJSON.
	RequestHeaderHookAcceptJSON = func(header http.Header) { header.Set("Accept", MIMEApplicationJSON) }
	// RequestHeaderHookAcceptIOTASerializerV1 is used to set the request "Accept" header to MIMEApplicationVendorIOTASerializerV1.
	RequestHeaderHookAcceptIOTASerializerV1 = func(header http.Header) { header.Set("Accept", MIMEApplicationVendorIOTASerializerV1) }
)

// the default options applied to the Client.
var defaultNodeAPIOptions = []ClientOption{
	WithHTTPClient(http.DefaultClient),
	WithUserInfo(nil),
	WithRequestURLHook(nil),
}

// ClientOptions define options for the Client.
type ClientOptions struct {
	// The HTTP client to use.
	httpClient *http.Client
	// The username and password information.
	userInfo *url.Userinfo
	// The hook to modify the URL before sending a request.
	requestURLHook RequestURLHook
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

// WithRequestURLHook is used to modify the URL before sending a request.
func WithRequestURLHook(requestURLHook RequestURLHook) ClientOption {
	return func(opts *ClientOptions) {
		opts.requestURLHook = requestURLHook
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
	return do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, method, route, client.opts.requestURLHook, nil, reqObj, resObj)
}

// DoWithRequestHeaderHook executes a request against the endpoint.
// This function is only meant to be used for special routes not covered through the standard API.
func (client *Client) DoWithRequestHeaderHook(ctx context.Context, method string, route string, requestHeaderHook RequestHeaderHook, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	return do(client.opts.httpClient, client.BaseURL, ctx, client.opts.userInfo, method, route, client.opts.requestURLHook, requestHeaderHook, reqObj, resObj)
}

// Indexer returns the IndexerClient.
// Returns ErrIndexerPluginNotAvailable if the current node does not support the plugin.
func (client *Client) Indexer(ctx context.Context) (IndexerClient, error) {
	hasPlugin, err := client.NodeSupportsPlugin(ctx, IndexerPluginName)
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
	hasPlugin, err := client.NodeSupportsPlugin(ctx, MQTTPluginName)
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
	if _, err := client.Do(ctx, http.MethodGet, RouteHealth, nil, nil); err != nil {
		if errors.Is(err, ErrHTTPServiceUnavailable) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// Info gets the info of the node.
func (client *Client) Info(ctx context.Context) (*InfoResponse, error) {
	res := &InfoResponse{}
	if _, err := client.Do(ctx, http.MethodGet, RouteInfo, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// NodeSupportsPlugin gets the info of the node and checks if the given plugin is enabled.
func (client *Client) NodeSupportsPlugin(ctx context.Context, pluginName string) (bool, error) {
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

// Tips returns the hex encoded tips as MessageIDs.
func (ntr *TipsResponse) Tips() (iotago.MessageIDs, error) {
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
func (client *Client) Tips(ctx context.Context) (*TipsResponse, error) {
	res := &TipsResponse{}
	if _, err := client.Do(ctx, http.MethodGet, RouteTips, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// SubmitMessage submits the given Message to the node.
// The node will take care of filling missing information.
// This function returns the finalized message created by the node.
func (client *Client) SubmitMessage(ctx context.Context, m *iotago.Message, protoParas *iotago.ProtocolParameters) (*iotago.Message, error) {
	// do not check the message because the validation would fail if
	// no parents were given. The node will first add this missing information and
	// validate the message afterwards.
	data, err := m.Serialize(serializer.DeSeriModeNoValidation, protoParas)
	if err != nil {
		return nil, err
	}

	req := &RawDataEnvelope{Data: data}
	res, err := client.Do(ctx, http.MethodPost, RouteMessages, req, nil)
	if err != nil {
		return nil, err
	}

	messageID, err := iotago.MessageIDFromHexString(res.Header.Get(locationHeader))
	if err != nil {
		return nil, err
	}

	msg, err := client.MessageByMessageID(ctx, messageID, protoParas)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// MessageMetadataByMessageID gets the metadata of a message by its message ID from the node.
func (client *Client) MessageMetadataByMessageID(ctx context.Context, msgID iotago.MessageID) (*MessageMetadataResponse, error) {
	query := fmt.Sprintf(RouteMessageMetadata, iotago.EncodeHex(msgID[:]))

	res := &MessageMetadataResponse{}
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// MessageByMessageID get a message by its message ID from the node.
func (client *Client) MessageByMessageID(ctx context.Context, msgID iotago.MessageID, protoParas *iotago.ProtocolParameters) (*iotago.Message, error) {
	query := fmt.Sprintf(RouteMessage, iotago.EncodeHex(msgID[:]))

	res := &RawDataEnvelope{}
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptIOTASerializerV1, nil, res); err != nil {
		return nil, err
	}

	msg := &iotago.Message{}
	if _, err := msg.Deserialize(res.Data, serializer.DeSeriModePerformValidation, protoParas); err != nil {
		return nil, err
	}

	return msg, nil
}

// ChildrenByMessageID gets the MessageIDs of the child messages of a given message.
func (client *Client) ChildrenByMessageID(ctx context.Context, parentMsgID iotago.MessageID) (*ChildrenResponse, error) {
	query := fmt.Sprintf(RouteMessageChildren, iotago.EncodeHex(parentMsgID[:]))

	res := &ChildrenResponse{}
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// TransactionIncludedMessage get a message that included the given transaction ID in the ledger.
func (client *Client) TransactionIncludedMessage(ctx context.Context, txID iotago.TransactionID, protoParas *iotago.ProtocolParameters) (*iotago.Message, error) {
	query := fmt.Sprintf(RouteTransactionsIncludedMessage, iotago.EncodeHex(txID[:]))

	res := &RawDataEnvelope{}
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptIOTASerializerV1, nil, res); err != nil {
		return nil, err
	}

	msg := &iotago.Message{}
	if _, err := msg.Deserialize(res.Data, serializer.DeSeriModePerformValidation, protoParas); err != nil {
		return nil, err
	}

	return msg, nil
}

// OutputByID gets an output by its ID from the node.
func (client *Client) OutputByID(ctx context.Context, outputID iotago.OutputID) (*OutputResponse, error) {
	query := fmt.Sprintf(RouteOutput, outputID.ToHex())

	res := &OutputResponse{}
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// OutputMetadataByID gets an output's metadata by its ID from the node without getting the output data again.
func (client *Client) OutputMetadataByID(ctx context.Context, outputID iotago.OutputID) (*OutputResponse, error) {
	query := fmt.Sprintf(RouteOutputMetadata, outputID.ToHex())

	res := &OutputResponse{}
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// Treasury gets the current treasury.
func (client *Client) Treasury(ctx context.Context) (*TreasuryResponse, error) {
	res := &TreasuryResponse{}
	if _, err := client.Do(ctx, http.MethodGet, RouteTreasury, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// Receipts gets all receipts persisted on the node.
func (client *Client) Receipts(ctx context.Context) ([]*ReceiptTuple, error) {
	res := &ReceiptsResponse{}
	if _, err := client.Do(ctx, http.MethodGet, RouteReceipts, nil, res); err != nil {
		return nil, err
	}

	return res.Receipts, nil
}

// ReceiptsByMigratedAtIndex gets all receipts for the given migrated at index persisted on the node.
func (client *Client) ReceiptsByMigratedAtIndex(ctx context.Context, index uint32) ([]*ReceiptTuple, error) {
	query := fmt.Sprintf(RouteReceiptsMigratedAtIndex, index)

	res := &ReceiptsResponse{}
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res.Receipts, nil
}

// MilestoneByIndex gets a milestone by its index.
func (client *Client) MilestoneByIndex(ctx context.Context, index uint32) (*iotago.Milestone, error) {
	query := fmt.Sprintf(RouteMilestoneByIndex, index)

	res := &iotago.Milestone{}
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// MilestoneByID gets a milestone by its ID.
func (client *Client) MilestoneByID(ctx context.Context, id iotago.MilestoneID) (*iotago.Milestone, error) {
	query := fmt.Sprintf(RouteMilestone, iotago.EncodeHex(id[:]))

	res := &iotago.Milestone{}
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// MilestoneUTXOChangesByIndex returns all UTXO changes of a milestone by its milestoneIndex.
func (client *Client) MilestoneUTXOChangesByIndex(ctx context.Context, index uint32) (*MilestoneUTXOChangesResponse, error) {
	query := fmt.Sprintf(RouteMilestoneUTXOChanges, index)

	res := &MilestoneUTXOChangesResponse{}
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// ComputeWhiteFlagMutations is the route to compute the white flag mutations for the cone of the given parents.
// This function returns the merkle tree roots calculated by the node.
func (client *Client) ComputeWhiteFlagMutations(ctx context.Context, index uint32, timestamp uint32, parents iotago.MessageIDs, previousMilestoneID iotago.MilestoneID) (*ComputeWhiteFlagMutationsResponse, error) {

	parentsHex := make([]string, len(parents))
	for i, parent := range parents {
		parentsHex[i] = iotago.EncodeHex(parent[:])
	}

	req := &ComputeWhiteFlagMutationsRequest{
		Index:               index,
		Timestamp:           timestamp,
		Parents:             parentsHex,
		PreviousMilestoneID: iotago.EncodeHex(previousMilestoneID[:]),
	}

	res := &ComputeWhiteFlagMutationsResponseInternal{}
	if _, err := client.Do(ctx, http.MethodPost, RouteComputeWhiteFlagMutations, req, res); err != nil {
		return nil, err
	}

	confirmedMerkleRootBytes, err := iotago.DecodeHex(res.ConfirmedMerkleRoot)
	if err != nil {
		return nil, err
	}

	if len(confirmedMerkleRootBytes) != iotago.MilestoneMerkleProofLength {
		return nil, fmt.Errorf("unknown confirmed merkle tree hash length (%d)", len(confirmedMerkleRootBytes))
	}

	appliedMerkleRootBytes, err := iotago.DecodeHex(res.AppliedMerkleRoot)
	if err != nil {
		return nil, err
	}

	if len(appliedMerkleRootBytes) != iotago.MilestoneMerkleProofLength {
		return nil, fmt.Errorf("unknown applied merkle tree hash length (%d)", len(appliedMerkleRootBytes))
	}

	result := &ComputeWhiteFlagMutationsResponse{
		ConfirmedMerkleRoot: iotago.MilestoneMerkleProof{},
		AppliedMerkleRoot:   iotago.MilestoneMerkleProof{},
	}
	copy(result.ConfirmedMerkleRoot[:], confirmedMerkleRootBytes)
	copy(result.AppliedMerkleRoot[:], appliedMerkleRootBytes)

	return result, nil
}

// PeerByID gets a peer by its identifier.
func (client *Client) PeerByID(ctx context.Context, id string) (*PeerResponse, error) {
	query := fmt.Sprintf(RoutePeer, id)

	res := &PeerResponse{}
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// RemovePeerByID removes a peer by its identifier.
func (client *Client) RemovePeerByID(ctx context.Context, id string) error {
	query := fmt.Sprintf(RoutePeer, id)

	if _, err := client.Do(ctx, http.MethodDelete, query, nil, nil); err != nil {
		return err
	}

	return nil
}

// Peers returns a list of all peers.
func (client *Client) Peers(ctx context.Context) ([]*PeerResponse, error) {
	res := []*PeerResponse{}
	if _, err := client.Do(ctx, http.MethodGet, RoutePeers, nil, &res); err != nil {
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
	if _, err := client.Do(ctx, http.MethodPost, RoutePeers, req, res); err != nil {
		return nil, err
	}

	return res, nil
}
