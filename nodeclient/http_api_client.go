package nodeclient

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/hexutil"
	"github.com/iotaledger/iota.go/v4/nodeclient/models"
)

const (
	// IndexerPluginName is the name for the indexer plugin.
	IndexerPluginName = "indexer/v2"

	// MQTTPluginName is the name for the MQTT plugin.
	MQTTPluginName = "mqtt/v2"
)

const (
	// RouteHealth is the route for querying a node's health status.
	RouteHealth = "/health"

	// RouteRoutes is the route for getting the routes the node supports.
	// GET returns the nodes routes.
	RouteRoutes = "/api/routes"

	// RouteInfo is the route for getting the node info.
	// GET returns the node info.
	RouteInfo = "/api/core/v3/info"

	// RouteBlockIssuance is the route for getting all needed information for block creation.
	// GET returns the data needed toa attach block.
	RouteBlockIssuance = "/api/core/v3/blocks/issuance"

	// RouteBlock is the route for getting a block by its ID.
	// GET returns the block based on the given type in the request "Accept" header.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes.
	RouteBlock = "/api/core/v3/blocks/%s"

	// RouteBlockMetadata is the route for getting block metadata by its ID.
	// GET returns block metadata (including info about "promotion/reattachment needed").
	RouteBlockMetadata = "/api/core/v3/blocks/%s/metadata"

	// RouteBlocks is the route for creating new blocks.
	// POST creates a single new block and returns the ID.
	// The block is parsed based on the given type in the request "Content-Type" header.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes.
	RouteBlocks = "/api/core/v3/blocks"

	// RouteTransactionsIncludedBlock is the route for getting the block that was included in the ledger for a given transaction ID.
	// GET returns the block based on the given type in the request "Accept" header.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes.
	RouteTransactionsIncludedBlock = "/api/core/v3/transactions/%s/included-block"

	// RouteTransactionsIncludedBlockMetadata is the route for getting the block metadata that was first confirmed in the ledger for a given transaction ID.
	// GET returns block metadata (including info about "promotion/reattachment needed").
	RouteTransactionsIncludedBlockMetadata = "/api/core/v3/transactions/%s/included-block/metadata"

	// RouteCommitmentByID is the route for getting a commitment by its ID.
	// GET returns the commitment.
	RouteCommitmentByID = "/api/core/v3/commitments/%s"

	// RouteCommitmentByIDUTXOChanges is the route for getting all UTXO changes of a milestone by its ID.
	// GET returns the output IDs of all UTXO changes.
	RouteCommitmentByIDUTXOChanges = "/api/core/v3/commitments/%s/utxo-changes"

	// RouteCommitmentByIndex is the route for getting a milestone by its milestoneIndex.
	// GET returns the milestone.
	RouteCommitmentByIndex = "/api/core/v3/commitments/by-index/%d"

	// RouteCommitmentByIndexUTXOChanges is the route for getting all UTXO changes of a milestone by its milestoneIndex.
	// GET returns the output IDs of all UTXO changes.
	RouteCommitmentByIndexUTXOChanges = "/api/core/v3/commitments/by-index/%d/utxo-changes"

	// RouteOutput is the route for getting an output by its outputID (transactionHash + outputIndex).
	// GET returns the output based on the given type in the request "Accept" header.
	// MIMEApplicationJSON => json
	// MIMEVendorIOTASerializer => bytes.
	RouteOutput = "/api/core/v3/outputs/%s"

	// RouteOutputMetadata is the route for getting output metadata by its outputID (transactionHash + outputIndex) without getting the data again.
	// GET returns the output metadata.
	RouteOutputMetadata = "/api/core/v3/outputs/%s/metadata"

	// RoutePeer is the route for getting peers by their peerID.
	// GET returns the peer
	// DELETE deletes the peer.
	RoutePeer = "/api/core/v3/peers/%s"

	// RoutePeers is the route for getting all peers of the node.
	// GET returns a list of all peers.
	// POST adds a new peer.
	RoutePeers = "/api/core/v3/peers"
)

var (
	// ErrIndexerPluginNotAvailable is returned when the indexer plugin is not available on the node.
	ErrIndexerPluginNotAvailable = ierrors.New("indexer plugin not available on the current node")
	// ErrMQTTPluginNotAvailable is returned when the MQTT plugin is not available on the node.
	ErrMQTTPluginNotAvailable = ierrors.New("mqtt plugin not available on the current node")
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
	// the iotago API instance to use.
	iotagoAPI iotago.API
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

// WithIOTAGoAPI is used to de/serialize objects.
func WithIOTAGoAPI(api iotago.API) ClientOption {
	return func(opts *ClientOptions) {
		opts.iotagoAPI = api
	}
}

// ClientOption is a function setting a Client option.
type ClientOption func(opts *ClientOptions)

const initInfoEndpointCallTimeout = 5 * time.Second

// New returns a new Client using the given base URL.
// This constructor will automatically call Client.Info() in order to initialize the Client
// with the appropriate protocol parameters and latest iotago.API version (use WithIOTAGoAPI() to override this behavior).
func New(baseURL string, opts ...ClientOption) (*Client, error) {

	options := &ClientOptions{}
	options.apply(defaultNodeAPIOptions...)
	options.apply(opts...)

	client := &Client{
		BaseURL: baseURL,
		opts:    options,
	}

	if client.opts.iotagoAPI == nil {
		ctx, cancelFunc := context.WithTimeout(context.Background(), initInfoEndpointCallTimeout)
		defer cancelFunc()
		info, err := client.Info(ctx)
		if err != nil {
			return nil, ierrors.Errorf("unable to call info endpoint for protocol parameter init: %w", err)
		}
		protoParams, err := info.DecodeProtocolParameters()
		if err != nil {
			return nil, ierrors.Errorf("unable to parse protocol parameters from info response: %w", err)
		}
		client.opts.iotagoAPI = iotago.LatestAPI(protoParams)
	}

	return client, nil
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

// HTTPClient returns the underlying HTTP client.
func (client *Client) HTTPClient() *http.Client {
	return client.opts.httpClient
}

// Do executes a request against the endpoint.
// This function is only meant to be used for special routes not covered through the standard API.
func (client *Client) Do(ctx context.Context, method string, route string, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	return do(ctx, client.opts.httpClient, client.BaseURL, client.opts.userInfo, method, route, client.opts.requestURLHook, nil, reqObj, resObj)
}

// DoWithRequestHeaderHook executes a request against the endpoint.
// This function is only meant to be used for special routes not covered through the standard API.
func (client *Client) DoWithRequestHeaderHook(ctx context.Context, method string, route string, requestHeaderHook RequestHeaderHook, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	return do(ctx, client.opts.httpClient, client.BaseURL, client.opts.userInfo, method, route, client.opts.requestURLHook, requestHeaderHook, reqObj, resObj)
}

// Indexer returns the IndexerClient.
// Returns ErrIndexerPluginNotAvailable if the current node does not support the plugin.
func (client *Client) Indexer(ctx context.Context) (IndexerClient, error) {
	hasPlugin, err := client.NodeSupportsRoute(ctx, IndexerPluginName)
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
	hasPlugin, err := client.NodeSupportsRoute(ctx, MQTTPluginName)
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
		if ierrors.Is(err, ErrHTTPServiceUnavailable) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// Routes gets the routes the node supports.
func (client *Client) Routes(ctx context.Context) (*models.RoutesResponse, error) {
	res := &models.RoutesResponse{}
	if _, err := client.Do(ctx, http.MethodGet, RouteRoutes, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// Info gets the info of the node.
func (client *Client) Info(ctx context.Context) (*models.InfoResponse, error) {
	res := &models.InfoResponse{}
	if _, err := client.Do(ctx, http.MethodGet, RouteInfo, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// BlockIssuance gets the info to issue a block.
func (client *Client) BlockIssuance(ctx context.Context) (*models.IssuanceBlockHeaderResponse, error) {
	res := &models.IssuanceBlockHeaderResponse{}
	if _, err := client.Do(ctx, http.MethodGet, RouteBlockIssuance, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// NodeSupportsRoute gets the routes of the node and checks if the given route is enabled.
func (client *Client) NodeSupportsRoute(ctx context.Context, route string) (bool, error) {
	routes, err := client.Routes(ctx)
	if err != nil {
		return false, err
	}
	for _, p := range routes.Routes {
		if p == route {
			return true, nil
		}
	}
	return false, nil
}

// SubmitBlock submits the given Block to the node.
// The node will take care of filling missing information.
// This function returns the blockID of the finalized block.
// To get the finalized block you need to call "BlockByBlockID".
func (client *Client) SubmitBlock(ctx context.Context, m *iotago.ProtocolBlock) (iotago.BlockID, error) {
	// do not check the block because the validation would fail if
	// no parents were given. The node will first add this missing information and
	// validate the block afterwards.
	data, err := client.opts.iotagoAPI.Encode(m)
	if err != nil {
		return iotago.EmptyBlockID(), err
	}

	req := &RawDataEnvelope{Data: data}
	res, err := client.Do(ctx, http.MethodPost, RouteBlocks, req, nil)
	if err != nil {
		return iotago.EmptyBlockID(), err
	}

	blockID, err := iotago.SlotIdentifierFromHexString(res.Header.Get(locationHeader))
	if err != nil {
		return iotago.EmptyBlockID(), err
	}

	return blockID, nil
}

// BlockMetadataByBlockID gets the metadata of a block by its ID from the node.
func (client *Client) BlockMetadataByBlockID(ctx context.Context, blockID iotago.BlockID) (*models.BlockMetadataResponse, error) {
	query := fmt.Sprintf(RouteBlockMetadata, hexutil.EncodeHex(blockID[:]))

	res := &models.BlockMetadataResponse{}
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// BlockByBlockID get a block by its block ID from the node.
func (client *Client) BlockByBlockID(ctx context.Context, blockID iotago.BlockID) (*iotago.ProtocolBlock, error) {
	query := fmt.Sprintf(RouteBlock, hexutil.EncodeHex(blockID[:]))

	res := &RawDataEnvelope{}
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptIOTASerializerV1, nil, res); err != nil {
		return nil, err
	}

	block := &iotago.ProtocolBlock{}
	if _, err := client.opts.iotagoAPI.Decode(res.Data, block, serix.WithValidation()); err != nil {
		return nil, err
	}

	return block, nil
}

// TransactionIncludedBlock get a block that included the given transaction ID in the ledger.
func (client *Client) TransactionIncludedBlock(ctx context.Context, txID iotago.TransactionID) (*iotago.ProtocolBlock, error) {
	query := fmt.Sprintf(RouteTransactionsIncludedBlock, hexutil.EncodeHex(txID[:]))

	res := &RawDataEnvelope{}
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptIOTASerializerV1, nil, res); err != nil {
		return nil, err
	}

	block := &iotago.ProtocolBlock{}
	if _, err := client.opts.iotagoAPI.Decode(res.Data, block, serix.WithValidation()); err != nil {
		return nil, err
	}

	return block, nil
}

// BlockMetadataByBlockID gets the metadata of a block by its ID from the node.
func (client *Client) TransactionIncludedBlockMetadata(ctx context.Context, txID iotago.TransactionID) (*models.BlockMetadataResponse, error) {
	query := fmt.Sprintf(RouteTransactionsIncludedBlockMetadata, hexutil.EncodeHex(txID[:]))

	res := &models.BlockMetadataResponse{}
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// OutputByID gets an output by its ID from the node.
func (client *Client) OutputByID(ctx context.Context, outputID iotago.OutputID) (iotago.Output, error) {
	query := fmt.Sprintf(RouteOutput, outputID.ToHex())

	res := &RawDataEnvelope{}
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptIOTASerializerV1, nil, res); err != nil {
		return nil, err
	}

	var output iotago.TxEssenceOutput
	if _, err := client.opts.iotagoAPI.Decode(res.Data, &output, serix.WithValidation()); err != nil {
		return nil, err
	}

	return output, nil
}

// OutputMetadataByID gets an output's metadata by its ID from the node without getting the output data again.
func (client *Client) OutputMetadataByID(ctx context.Context, outputID iotago.OutputID) (*models.OutputMetadataResponse, error) {
	query := fmt.Sprintf(RouteOutputMetadata, outputID.ToHex())

	res := &models.OutputMetadataResponse{}
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentByID gets a commitment details by its ID.
func (client *Client) CommitmentByID(ctx context.Context, id iotago.CommitmentID) (*iotago.Commitment, error) {
	query := fmt.Sprintf(RouteCommitmentByID, id.ToHex())

	res := &iotago.Commitment{}
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentUTXOChangesByID returns all UTXO changes of a commitment by its ID.
func (client *Client) CommitmentUTXOChangesByID(ctx context.Context, id iotago.CommitmentID) (*models.UTXOChangesResponse, error) {
	query := fmt.Sprintf(RouteCommitmentByIDUTXOChanges, id.ToHex())

	res := &models.UTXOChangesResponse{}
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentByIndex gets a commitment details by its index.
func (client *Client) CommitmentByIndex(ctx context.Context, index iotago.SlotIndex) (*iotago.Commitment, error) {
	query := fmt.Sprintf(RouteCommitmentByIndex, index)

	res := &iotago.Commitment{}
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentUTXOChangesByIndex returns all UTXO changes of a commitment by its index.
func (client *Client) CommitmentUTXOChangesByIndex(ctx context.Context, index iotago.SlotIndex) (*models.UTXOChangesResponse, error) {
	query := fmt.Sprintf(RouteCommitmentByIndexUTXOChanges, index)

	res := &models.UTXOChangesResponse{}
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// PeerByID gets a peer by its identifier.
func (client *Client) PeerByID(ctx context.Context, id string) (*models.PeerResponse, error) {
	query := fmt.Sprintf(RoutePeer, id)

	res := &models.PeerResponse{}
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
func (client *Client) Peers(ctx context.Context) ([]*models.PeerResponse, error) {
	res := []*models.PeerResponse{}
	if _, err := client.Do(ctx, http.MethodGet, RoutePeers, nil, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// AddPeer adds a new peer by libp2p multi address with optional alias.
func (client *Client) AddPeer(ctx context.Context, multiAddress string, alias ...string) (*models.PeerResponse, error) {
	req := &models.AddPeerRequest{
		MultiAddress: multiAddress,
	}

	if len(alias) > 0 {
		req.Alias = &alias[0]
	}

	res := &models.PeerResponse{}
	if _, err := client.Do(ctx, http.MethodPost, RoutePeers, req, res); err != nil {
		return nil, err
	}

	return res, nil
}
