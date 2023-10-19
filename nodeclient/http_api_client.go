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
	"github.com/iotaledger/iota.go/v4/api"
	"github.com/iotaledger/iota.go/v4/hexutil"
	"github.com/iotaledger/iota.go/v4/nodeclient/apimodels"
)

const (
	// IndexerPluginName is the name for the indexer plugin.
	IndexerPluginName = "indexer/v2"

	// MQTTPluginName is the name for the MQTT plugin.
	MQTTPluginName = "mqtt/v2"

	// BlockIssuerPluginName is the name for the blockissuer plugin.
	BlockIssuerPluginName = "blockissuer/v1"
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

	// RouteCongestion is the route for getting congestion details for the account.
	// GET returns the congestion details for the account.
	// MIMEApplicationJSON => json.
	// MIMEApplicationVendorIOTASerializerV2 => bytes.
	RouteCongestion = "/api/core/v3/accounts/%s/congestion"

	// RouteRewards is the route for getting the rewards for staking or delegation based on the provided output.
	// Rewards are decayed up to returned epochEnd index.
	// GET returns the rewards for the output.
	// MIMEApplicationJSON => json.
	// MIMEApplicationVendorIOTASerializerV2 => bytes.
	RouteRewards = "/api/core/v3/rewards/%s"

	// RouteValidators is the route for getting the information about current registered validators.
	// GET returns the paginated information about about registered validators.
	// MIMEApplicationJSON => json.
	// MIMEApplicationVendorIOTASerializerV2 => bytes.
	RouteValidators = "/api/core/v3/validators"

	// RouteValidatorsAccount is the route for getting validator by its accountID.
	// GET returns the account details.
	// MIMEApplicationJSON => json.
	// MIMEApplicationVendorIOTASerializerV2 => bytes.
	RouteValidatorsAccount = "/api/core/v3/validators/%s"

	// RouteCommittee is the route for getting the information about the current committee.
	// GET returns the information about the current committee.
	// MIMEApplicationJSON => json.
	// MIMEApplicationVendorIOTASerializerV2 => bytes.
	RouteCommittee = "/api/core/v3/committee"

	// RouteBlockIssuance is the route for getting all needed information for block creation.
	// GET returns the data needed toa attach block.
	// MIMEApplicationJSON => json.
	// MIMEApplicationVendorIOTASerializerV2 => bytes.
	RouteBlockIssuance = "/api/core/v3/blocks/issuance"

	// RouteBlock is the route for getting a block by its ID.
	// GET returns the block based on the given type in the request "Accept" header.
	// MIMEApplicationJSON => json
	// MIMEApplicationVendorIOTASerializerV2 => bytes.
	RouteBlock = "/api/core/v3/blocks/%s"

	// RouteBlockMetadata is the route for getting block metadata by its ID.
	// GET returns block metadata (including info about "promotion/reattachment needed").
	RouteBlockMetadata = "/api/core/v3/blocks/%s/metadata"

	// RouteBlocks is the route for creating new blocks.
	// POST creates a single new block and returns the ID.
	// The block is parsed based on the given type in the request "Content-Type" header.
	// MIMEApplicationJSON => json
	// MIMEApplicationVendorIOTASerializerV2 => bytes.
	RouteBlocks = "/api/core/v3/blocks"

	// RouteTransactionsIncludedBlock is the route for getting the block that was included in the ledger for a given transaction ID.
	// GET returns the block based on the given type in the request "Accept" header.
	// MIMEApplicationJSON => json
	// MIMEApplicationVendorIOTASerializerV2 => bytes.
	RouteTransactionsIncludedBlock = "/api/core/v3/transactions/%s/included-block"

	// RouteTransactionsIncludedBlockMetadata is the route for getting the block metadata that was first confirmed in the ledger for a given transaction ID.
	// GET returns block metadata (including info about "promotion/reattachment needed").
	// MIMEApplicationJSON => json.
	// MIMEApplicationVendorIOTASerializerV2 => bytes.
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
	// MIMEApplicationVendorIOTASerializerV2 => bytes.
	RouteOutput = "/api/core/v3/outputs/%s"

	// RouteOutputMetadata is the route for getting output metadata by its outputID (transactionHash + outputIndex) without getting the data again.
	// GET returns the output metadata.
	RouteOutputMetadata = "/api/core/v3/outputs/%s/metadata"

	// RouteOutputWithMetadata is the route for getting output and its metadata by its outputID (transactionHash + outputIndex).
	// GET returns the output metadata.
	RouteOutputWithMetadata = "/api/core/v3/outputs/%s/full"

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
	// ErrBlockIssuerPluginNotAvailable is returned when the BlockIssuer plugin is not available on the node.
	ErrBlockIssuerPluginNotAvailable = ierrors.New("blockissuer plugin not available on the current node")
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
	// RequestHeaderHookAcceptIOTASerializerV2 is used to set the request "Accept" header to MIMEApplicationVendorIOTASerializerV2.
	RequestHeaderHookAcceptIOTASerializerV2 = func(header http.Header) { header.Set("Accept", MIMEApplicationVendorIOTASerializerV2) }
	// RequestHeaderHookContentTypeIOTASerializerV2 is used to set the request "Content-Type" header to MIMEApplicationVendorIOTASerializerV2.
	RequestHeaderHookContentTypeIOTASerializerV2 = func(header http.Header) { header.Set("Content-Type", MIMEApplicationVendorIOTASerializerV2) }
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

const initInfoEndpointCallTimeout = 5 * time.Second

// New returns a new Client using the given base URL.
// This constructor will automatically call Client.Info() in order to initialize the Client
// with the appropriate protocol parameters and latest iotago.API version (use WithIOTAGoAPI() to override this behavior).
func New(baseURL string, opts ...ClientOption) (*Client, error) {

	options := &ClientOptions{}
	options.apply(defaultNodeAPIOptions...)
	options.apply(opts...)

	client := &Client{
		BaseURL:     baseURL,
		apiProvider: api.NewEpochBasedProvider(),
		opts:        options,
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), initInfoEndpointCallTimeout)
	defer cancelFunc()
	info, err := client.Info(ctx)
	if err != nil {
		return nil, ierrors.Errorf("unable to call info endpoint for protocol parameter init: %w", err)
	}
	for _, params := range info.ProtocolParameters {
		client.apiProvider.AddProtocolParametersAtEpoch(params.Parameters, params.StartEpoch)
	}

	return client, nil
}

// Client is a client for node HTTP REST API endpoints.
type Client struct {
	// The base URL for all API calls.
	BaseURL string

	apiProvider *api.EpochBasedProvider

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
	return do(ctx, client.CommittedAPI().Underlying(), client.opts.httpClient, client.BaseURL, client.opts.userInfo, method, route, client.opts.requestURLHook, nil, reqObj, resObj)
}

// DoWithRequestHeaderHook executes a request against the endpoint.
// This function is only meant to be used for special routes not covered through the standard API.
func (client *Client) DoWithRequestHeaderHook(ctx context.Context, method string, route string, requestHeaderHook RequestHeaderHook, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	return do(ctx, client.CommittedAPI().Underlying(), client.opts.httpClient, client.BaseURL, client.opts.userInfo, method, route, client.opts.requestURLHook, requestHeaderHook, reqObj, resObj)
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

// BlockIssuer returns the BlockIssuerClient.
// Returns ErrBlockIssuerPluginNotAvailable if the current node does not support the plugin.
func (client *Client) BlockIssuer(ctx context.Context) (BlockIssuerClient, error) {
	hasPlugin, err := client.NodeSupportsRoute(ctx, BlockIssuerPluginName)
	if err != nil {
		return nil, err
	}
	if !hasPlugin {
		return nil, ErrBlockIssuerPluginNotAvailable
	}

	return &blockIssuerClient{core: client}, nil
}

// Health returns whether the given node is healthy.
func (client *Client) Health(ctx context.Context) (bool, error) {
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, RouteHealth, nil, nil); err != nil {
		if ierrors.Is(err, ErrHTTPServiceUnavailable) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// Routes gets the routes the node supports.
func (client *Client) Routes(ctx context.Context) (*apimodels.RoutesResponse, error) {
	//nolint:bodyclose
	res := new(apimodels.RoutesResponse)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, RouteRoutes, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// Info gets the info of the node.
func (client *Client) Info(ctx context.Context) (*apimodels.InfoResponse, error) {
	res := new(apimodels.InfoResponse)
	//nolint:bodyclose
	if _, err := do(ctx, iotago.CommonSerixAPI(), client.opts.httpClient, client.BaseURL, client.opts.userInfo, http.MethodGet, RouteInfo, client.opts.requestURLHook, nil, nil, res); err != nil {
		return nil, err
	}

	client.apiProvider.SetCommittedSlot(res.Status.LatestCommitmentID.Slot())

	return res, nil
}

// BlockIssuance gets the info to issue a block.
func (client *Client) BlockIssuance(ctx context.Context) (*apimodels.IssuanceBlockHeaderResponse, error) {
	res := new(apimodels.IssuanceBlockHeaderResponse)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, RouteBlockIssuance, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (client *Client) Congestion(ctx context.Context, accountID iotago.AccountID) (*apimodels.CongestionResponse, error) {
	res := new(apimodels.CongestionResponse)
	query := fmt.Sprintf(RouteCongestion, hexutil.EncodeHex(accountID[:]))
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (client *Client) Rewards(ctx context.Context, outputID iotago.OutputID) (*apimodels.ManaRewardsResponse, error) {
	res := &apimodels.ManaRewardsResponse{}
	query := fmt.Sprintf(RouteRewards, hexutil.EncodeHex(outputID[:]))
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (client *Client) Validators(ctx context.Context) (*apimodels.ValidatorsResponse, error) {
	res := &apimodels.ValidatorsResponse{}
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, RouteValidators, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (client *Client) StakingAccount(ctx context.Context, accountID iotago.AccountID) (*apimodels.ValidatorResponse, error) {
	res := &apimodels.ValidatorResponse{}
	query := fmt.Sprintf(RouteValidatorsAccount, hexutil.EncodeHex(accountID[:]))
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (client *Client) Committee(ctx context.Context, optEpochIndex ...iotago.EpochIndex) (*apimodels.CommitteeResponse, error) {
	query := RouteCommittee
	if len(optEpochIndex) > 0 {
		query += fmt.Sprintf("?epochIndex=%d", optEpochIndex[0])
	}
	fmt.Printf("query: %s\n", query)

	res := &apimodels.CommitteeResponse{}
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
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
	// validate the block afterward.

	apiForVersion, err := client.APIForVersion(m.ProtocolVersion)
	if err != nil {
		return iotago.EmptyBlockID, err
	}

	data, err := apiForVersion.Encode(m)
	if err != nil {
		return iotago.EmptyBlockID, err
	}

	req := &RawDataEnvelope{Data: data}
	//nolint:bodyclose
	res, err := client.Do(ctx, http.MethodPost, RouteBlocks, req, nil)
	if err != nil {
		return iotago.EmptyBlockID, err
	}

	blockID, err := iotago.BlockIDFromHexString(res.Header.Get(locationHeader))
	if err != nil {
		return iotago.EmptyBlockID, err
	}

	return blockID, nil
}

// BlockMetadataByBlockID gets the metadata of a block by its ID from the node.
func (client *Client) BlockMetadataByBlockID(ctx context.Context, blockID iotago.BlockID) (*apimodels.BlockMetadataResponse, error) {
	query := fmt.Sprintf(RouteBlockMetadata, hexutil.EncodeHex(blockID[:]))

	res := new(apimodels.BlockMetadataResponse)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// BlockByBlockID get a block by its block ID from the node.
func (client *Client) BlockByBlockID(ctx context.Context, blockID iotago.BlockID) (*iotago.ProtocolBlock, error) {
	query := fmt.Sprintf(RouteBlock, hexutil.EncodeHex(blockID[:]))

	res := new(RawDataEnvelope)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptIOTASerializerV2, nil, res); err != nil {
		return nil, err
	}

	block, _, err := iotago.ProtocolBlockFromBytes(client)(res.Data)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// TransactionIncludedBlock get a block that included the given transaction ID in the ledger.
func (client *Client) TransactionIncludedBlock(ctx context.Context, txID iotago.TransactionID) (*iotago.ProtocolBlock, error) {
	query := fmt.Sprintf(RouteTransactionsIncludedBlock, hexutil.EncodeHex(txID[:]))

	res := new(RawDataEnvelope)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptIOTASerializerV2, nil, res); err != nil {
		return nil, err
	}

	block, _, err := iotago.ProtocolBlockFromBytes(client)(res.Data)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// BlockMetadataByBlockID gets the metadata of a block by its ID from the node.
func (client *Client) TransactionIncludedBlockMetadata(ctx context.Context, txID iotago.TransactionID) (*apimodels.BlockMetadataResponse, error) {
	query := fmt.Sprintf(RouteTransactionsIncludedBlockMetadata, hexutil.EncodeHex(txID[:]))

	res := new(apimodels.BlockMetadataResponse)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// OutputByID gets an output by its ID from the node.
func (client *Client) OutputByID(ctx context.Context, outputID iotago.OutputID) (iotago.Output, error) {
	query := fmt.Sprintf(RouteOutput, outputID.ToHex())

	res := new(RawDataEnvelope)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptIOTASerializerV2, nil, res); err != nil {
		return nil, err
	}

	var outputResponse apimodels.OutputResponse
	if _, err := client.CommittedAPI().Decode(res.Data, &outputResponse, serix.WithValidation()); err != nil {
		return nil, err
	}

	derivedOutputID, err := outputResponse.OutputIDProof.OutputID(outputResponse.Output)
	if err != nil {
		return nil, err
	}

	if derivedOutputID != outputID {
		return nil, ierrors.Errorf("output ID mismatch. Expected %s, got %s", outputID.ToHex(), derivedOutputID.ToHex())
	}

	return outputResponse.Output, nil
}

// OutputWithMetadataByID gets an output by its ID, together with the metadata from the node.
func (client *Client) OutputWithMetadataByID(ctx context.Context, outputID iotago.OutputID) (iotago.Output, *apimodels.OutputMetadata, error) {
	query := fmt.Sprintf(RouteOutputWithMetadata, outputID.ToHex())

	res := new(RawDataEnvelope)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptIOTASerializerV2, nil, res); err != nil {
		return nil, nil, err
	}

	var outputResponse apimodels.OutputWithMetadataResponse
	if _, err := client.CommittedAPI().Decode(res.Data, &outputResponse, serix.WithValidation()); err != nil {
		return nil, nil, err
	}

	derivedOutputID, err := outputResponse.OutputIDProof.OutputID(outputResponse.Output)
	if err != nil {
		return nil, nil, err
	}

	if derivedOutputID != outputID {
		return nil, nil, ierrors.Errorf("output ID mismatch. Expected %s, got %s", outputID.ToHex(), derivedOutputID.ToHex())
	}

	return outputResponse.Output, outputResponse.Metadata, nil
}

// OutputMetadataByID gets an output's metadata by its ID from the node without getting the output data again.
func (client *Client) OutputMetadataByID(ctx context.Context, outputID iotago.OutputID) (*apimodels.OutputMetadata, error) {
	query := fmt.Sprintf(RouteOutputMetadata, outputID.ToHex())

	res := new(apimodels.OutputMetadata)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentByID gets a commitment details by its ID.
func (client *Client) CommitmentByID(ctx context.Context, id iotago.CommitmentID) (*iotago.Commitment, error) {
	query := fmt.Sprintf(RouteCommitmentByID, id.ToHex())

	res := new(iotago.Commitment)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentUTXOChangesByID returns all UTXO changes of a commitment by its ID.
func (client *Client) CommitmentUTXOChangesByID(ctx context.Context, id iotago.CommitmentID) (*apimodels.UTXOChangesResponse, error) {
	query := fmt.Sprintf(RouteCommitmentByIDUTXOChanges, id.ToHex())

	res := new(apimodels.UTXOChangesResponse)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentByIndex gets a commitment details by its index.
func (client *Client) CommitmentByIndex(ctx context.Context, index iotago.SlotIndex) (*iotago.Commitment, error) {
	query := fmt.Sprintf(RouteCommitmentByIndex, index)

	res := new(iotago.Commitment)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentUTXOChangesByIndex returns all UTXO changes of a commitment by its index.
func (client *Client) CommitmentUTXOChangesByIndex(ctx context.Context, index iotago.SlotIndex) (*apimodels.UTXOChangesResponse, error) {
	query := fmt.Sprintf(RouteCommitmentByIndexUTXOChanges, index)

	res := new(apimodels.UTXOChangesResponse)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// PeerByID gets a peer by its identifier.
func (client *Client) PeerByID(ctx context.Context, id string) (*apimodels.PeerInfo, error) {
	query := fmt.Sprintf(RoutePeer, id)

	res := new(apimodels.PeerInfo)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// RemovePeerByID removes a peer by its identifier.
func (client *Client) RemovePeerByID(ctx context.Context, id string) error {
	query := fmt.Sprintf(RoutePeer, id)

	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodDelete, query, nil, nil); err != nil {
		return err
	}

	return nil
}

// Peers returns a list of all peers.
func (client *Client) Peers(ctx context.Context) (*apimodels.PeersResponse, error) {
	res := new(apimodels.PeersResponse)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, RoutePeers, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// AddPeer adds a new peer by libp2p multi address with optional alias.
func (client *Client) AddPeer(ctx context.Context, multiAddress string, alias ...string) (*apimodels.PeerInfo, error) {
	req := &apimodels.AddPeerRequest{
		MultiAddress: multiAddress,
	}

	if len(alias) > 0 {
		req.Alias = alias[0]
	}

	res := new(apimodels.PeerInfo)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodPost, RoutePeers, req, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (client *Client) APIForVersion(version iotago.Version) (iotago.API, error) {
	return client.apiProvider.APIForVersion(version)
}

func (client *Client) APIForEpoch(epoch iotago.EpochIndex) iotago.API {
	return client.apiProvider.APIForEpoch(epoch)
}

func (client *Client) APIForTime(t time.Time) iotago.API {
	return client.apiProvider.APIForTime(t)
}

func (client *Client) APIForSlot(slot iotago.SlotIndex) iotago.API {
	return client.apiProvider.APIForSlot(slot)
}

func (client *Client) CommittedAPI() iotago.API {
	return client.apiProvider.CommittedAPI()
}

func (client *Client) LatestAPI() iotago.API {
	return client.apiProvider.LatestAPI()
}

var _ iotago.APIProvider = new(Client)
