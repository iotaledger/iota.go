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
	"github.com/iotaledger/iota.go/v4/nodeclient/apimodels"
)

const (
	RootAPI = "/api"

	// CorePluginName is the name for the core API plugin.
	CorePluginName = "core/v3"

	// ManagementPluginName is the name for the management plugin.
	ManagementPluginName = "management/v1"

	// IndexerPluginName is the name for the indexer plugin.
	IndexerPluginName = "indexer/v2"

	// MQTTPluginName is the name for the MQTT plugin.
	MQTTPluginName = "mqtt/v2"

	// BlockIssuerPluginName is the name for the blockissuer plugin.
	BlockIssuerPluginName = "blockissuer/v1"
)

const (
	// QueryParameterEpochIndex is used to identify an epoch by index.
	QueryParameterEpochIndex = "epochIndex"

	// QueryParameterCommitmentID is used to identify a slot commitment by its ID.
	QueryParameterCommitmentID = "commitmentID"
)

const (
	// CoreEndpointInfo is the endpoint for getting the node info.
	// GET returns the node info.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointInfo = "/info"

	// CoreEndpointBlocks is the endpoint for sending new blocks.
	// POST sends a single new block and returns the new block ID.
	// "Content-Type" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointBlocks = "/blocks"

	// CoreEndpointBlock is the endpoint for getting a block by its blockID.
	// GET returns the block.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointBlock = "/blocks/%s"

	// CoreEndpointBlockMetadata is the endpoint for getting block metadata by its blockID.
	// GET returns block metadata.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointBlockMetadata = "/blocks/%s/metadata"

	// CoreEndpointBlockWithMetadata is the endpoint for getting a block, together with its metadata by its blockID.
	// GET returns the block and metadata.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointBlockWithMetadata = "/blocks/%s/full"

	// CoreEndpointBlockIssuance is the endpoint for getting all needed information for block creation.
	// GET returns the data needed to attach a block.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointBlockIssuance = "/blocks/issuance"

	// CoreEndpointOutput is the endpoint for getting an output by its outputID (transactionHash + outputIndex). This includes the proof, that the output corresponds to the requested outputID.
	// GET returns the output.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointOutput = "/outputs/%s"

	// CoreEndpointOutputMetadata is the endpoint for getting output metadata by its outputID (transactionHash + outputIndex) without getting the output itself again.
	// GET returns the output metadata.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointOutputMetadata = "/outputs/%s/metadata"

	// CoreEndpointOutputWithMetadata is the endpoint for getting output, together with its metadata by its outputID (transactionHash + outputIndex).
	// GET returns the output and metadata.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointOutputWithMetadata = "/outputs/%s/full"

	// CoreEndpointTransactionsIncludedBlock is the endpoint for getting the block that was first confirmed for a given transaction ID.
	// GET returns the block.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointTransactionsIncludedBlock = "/transactions/%s/included-block"

	// CoreEndpointTransactionsIncludedBlockMetadata is the endpoint for getting the metadata for the block that was first confirmed in the ledger for a given transaction ID.
	// GET returns block metadata.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointTransactionsIncludedBlockMetadata = "/transactions/%s/included-block/metadata"

	// CoreEndpointCommitmentByID is the endpoint for getting a slot commitment by its ID.
	// GET returns the commitment.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointCommitmentByID = "/commitments/%s"

	// CoreEndpointCommitmentByIDUTXOChanges is the endpoint for getting all UTXO changes of a commitment by its ID.
	// GET returns the output IDs of all UTXO changes.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointCommitmentByIDUTXOChanges = "/commitments/%s/utxo-changes"

	// CoreEndpointCommitmentByIndex is the endpoint for getting a commitment by its Slot.
	// GET returns the commitment.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointCommitmentByIndex = "/commitments/by-index/%s"

	// CoreEndpointCommitmentByIndexUTXOChanges is the endpoint for getting all UTXO changes of a commitment by its Slot.
	// GET returns the output IDs of all UTXO changes.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointCommitmentByIndexUTXOChanges = "/commitments/by-index/%s/utxo-changes"

	// CoreEndpointCongestion is the endpoint for getting the current congestion state and all account related useful details as block issuance credits.
	// GET returns the congestion state related to the specified account. (optional query parameters: "QueryParameterCommitmentID" to specify the used commitment)
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointCongestion = "/accounts/%s/congestion"

	// CoreEndpointValidators is the endpoint for getting informations about the current registered validators.
	// GET returns the paginated response with the list of validators.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointValidators = "/validators"

	// CoreEndpointValidatorsAccount is the endpoint for getting details about the validator by its bech32 account address.
	// GET returns the validator details.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointValidatorsAccount = "/validators/%s"

	// CoreEndpointRewards is the endpoint for getting the rewards for staking or delegation based on staking account or delegation output.
	// Rewards are decayed up to returned epochEnd index.
	// GET returns the rewards.
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointRewards = "/rewards/%s"

	// CoreEndpointCommittee is the endpoint for getting information about the current committee.
	// GET returns the information about the current committee. (optional query parameters: "QueryParameterEpochIndex" to specify the epoch)
	// "Accept" header:
	// 		MIMEApplicationJSON => json.
	// 		MIMEApplicationVendorIOTASerializerV2 => bytes.
	CoreEndpointCommittee = "/committee"
)

func route(pluginName, endpoint string) string {
	return fmt.Sprintf("%s/%s%s", RootAPI, pluginName, endpoint)
}

var (
	// RouteHealth is the route for querying a node's health status.
	RouteHealth = "/health"

	// RouteRoutes is the route for getting the routes the node supports.
	// GET returns the nodes routes.
	RouteRoutes = route("", "/routes")

	CoreRouteInfo                              = route(CorePluginName, CoreEndpointInfo)
	CoreRouteBlocks                            = route(CorePluginName, CoreEndpointBlocks)
	CoreRouteBlock                             = route(CorePluginName, CoreEndpointBlock)
	CoreRouteBlockMetadata                     = route(CorePluginName, CoreEndpointBlockMetadata)
	CoreRouteBlockWithMetadata                 = route(CorePluginName, CoreEndpointBlockWithMetadata)
	CoreRouteBlockIssuance                     = route(CorePluginName, CoreEndpointBlockIssuance)
	CoreRouteOutput                            = route(CorePluginName, CoreEndpointOutput)
	CoreRouteOutputMetadata                    = route(CorePluginName, CoreEndpointOutputMetadata)
	CoreRouteOutputWithMetadata                = route(CorePluginName, CoreEndpointOutputWithMetadata)
	CoreRouteTransactionsIncludedBlock         = route(CorePluginName, CoreEndpointTransactionsIncludedBlock)
	CoreRouteTransactionsIncludedBlockMetadata = route(CorePluginName, CoreEndpointTransactionsIncludedBlockMetadata)
	CoreRouteCommitmentByID                    = route(CorePluginName, CoreEndpointCommitmentByID)
	CoreRouteCommitmentByIDUTXOChanges         = route(CorePluginName, CoreEndpointCommitmentByIDUTXOChanges)
	CoreRouteCommitmentByIndex                 = route(CorePluginName, CoreEndpointCommitmentByIndex)
	CoreRouteCommitmentByIndexUTXOChanges      = route(CorePluginName, CoreEndpointCommitmentByIndexUTXOChanges)
	CoreRouteCongestion                        = route(CorePluginName, CoreEndpointCongestion)
	CoreRouteValidators                        = route(CorePluginName, CoreEndpointValidators)
	CoreRouteValidatorsAccount                 = route(CorePluginName, CoreEndpointValidatorsAccount)
	CoreRouteRewards                           = route(CorePluginName, CoreEndpointRewards)
	CoreRouteCommittee                         = route(CorePluginName, CoreEndpointCommittee)
)

var (
	// ErrManagementPluginNotAvailable is returned when the Management plugin is not available on the node.
	ErrManagementPluginNotAvailable = ierrors.New("management plugin not available on the current node")
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
		apiProvider: iotago.NewEpochBasedProvider(),
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

	apiProvider *iotago.EpochBasedProvider

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

// Management returns the ManagementClient.
// Returns ErrManagementPluginNotAvailable if the current node does not support the plugin.
func (client *Client) Management(ctx context.Context) (ManagementClient, error) {
	hasPlugin, err := client.NodeSupportsRoute(ctx, ManagementPluginName)
	if err != nil {
		return nil, err
	}
	if !hasPlugin {
		return nil, ErrManagementPluginNotAvailable
	}

	return &managementClient{core: client}, nil
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
	if _, err := do(ctx, iotago.CommonSerixAPI(), client.opts.httpClient, client.BaseURL, client.opts.userInfo, http.MethodGet, CoreRouteInfo, client.opts.requestURLHook, nil, nil, res); err != nil {
		return nil, err
	}

	client.apiProvider.SetCommittedSlot(res.Status.LatestCommitmentID.Slot())

	return res, nil
}

// BlockIssuance gets the info to issue a block.
func (client *Client) BlockIssuance(ctx context.Context) (*apimodels.IssuanceBlockHeaderResponse, error) {
	res := new(apimodels.IssuanceBlockHeaderResponse)

	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, CoreRouteBlockIssuance, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (client *Client) Congestion(ctx context.Context, accountAddress *iotago.AccountAddress, optCommitmentID ...iotago.CommitmentID) (*apimodels.CongestionResponse, error) {
	res := new(apimodels.CongestionResponse)

	//nolint:contextcheck
	query := fmt.Sprintf(CoreRouteCongestion, accountAddress.Bech32(client.CommittedAPI().ProtocolParameters().Bech32HRP()))

	if len(optCommitmentID) > 0 {
		query += fmt.Sprintf("?%s=%s", QueryParameterCommitmentID, optCommitmentID[0].ToHex())
	}

	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (client *Client) Rewards(ctx context.Context, outputID iotago.OutputID) (*apimodels.ManaRewardsResponse, error) {
	res := &apimodels.ManaRewardsResponse{}
	query := fmt.Sprintf(CoreRouteRewards, hexutil.EncodeHex(outputID[:]))
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (client *Client) Validators(ctx context.Context) (*apimodels.ValidatorsResponse, error) {
	res := &apimodels.ValidatorsResponse{}
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, CoreRouteValidators, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (client *Client) StakingAccount(ctx context.Context, accountAddress *iotago.AccountAddress) (*apimodels.ValidatorResponse, error) {
	res := &apimodels.ValidatorResponse{}

	//nolint:contextcheck
	query := fmt.Sprintf(CoreRouteValidatorsAccount, accountAddress.Bech32(client.CommittedAPI().ProtocolParameters().Bech32HRP()))
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (client *Client) Committee(ctx context.Context, optEpochIndex ...iotago.EpochIndex) (*apimodels.CommitteeResponse, error) {
	query := CoreRouteCommittee
	if len(optEpochIndex) > 0 {
		query += fmt.Sprintf("?%s=%d", QueryParameterEpochIndex, optEpochIndex[0])
	}

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
func (client *Client) SubmitBlock(ctx context.Context, m *iotago.Block) (iotago.BlockID, error) {
	// do not check the block because the validation would fail if
	// no parents were given. The node will first add this missing information and
	// validate the block afterward.

	apiForVersion, err := client.APIForVersion(m.Header.ProtocolVersion)
	if err != nil {
		return iotago.EmptyBlockID, err
	}

	data, err := apiForVersion.Encode(m)
	if err != nil {
		return iotago.EmptyBlockID, err
	}

	req := &RawDataEnvelope{Data: data}
	//nolint:bodyclose
	res, err := client.Do(ctx, http.MethodPost, CoreRouteBlocks, req, nil)
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
	query := fmt.Sprintf(CoreRouteBlockMetadata, hexutil.EncodeHex(blockID[:]))

	res := new(apimodels.BlockMetadataResponse)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// BlockByBlockID get a block by its block ID from the node.
func (client *Client) BlockByBlockID(ctx context.Context, blockID iotago.BlockID) (*iotago.Block, error) {
	query := fmt.Sprintf(CoreRouteBlock, hexutil.EncodeHex(blockID[:]))

	res := new(RawDataEnvelope)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptIOTASerializerV2, nil, res); err != nil {
		return nil, err
	}

	block, _, err := iotago.BlockFromBytes(client)(res.Data)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// TransactionIncludedBlock get a block that included the given transaction ID in the ledger.
func (client *Client) TransactionIncludedBlock(ctx context.Context, txID iotago.TransactionID) (*iotago.Block, error) {
	query := fmt.Sprintf(CoreRouteTransactionsIncludedBlock, hexutil.EncodeHex(txID[:]))

	res := new(RawDataEnvelope)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptIOTASerializerV2, nil, res); err != nil {
		return nil, err
	}

	block, _, err := iotago.BlockFromBytes(client)(res.Data)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// TransactionIncludedBlockMetadata gets the metadata of a block by its ID from the node.
func (client *Client) TransactionIncludedBlockMetadata(ctx context.Context, txID iotago.TransactionID) (*apimodels.BlockMetadataResponse, error) {
	query := fmt.Sprintf(CoreRouteTransactionsIncludedBlockMetadata, hexutil.EncodeHex(txID[:]))

	res := new(apimodels.BlockMetadataResponse)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// OutputByID gets an output by its ID from the node.
func (client *Client) OutputByID(ctx context.Context, outputID iotago.OutputID) (iotago.Output, error) {
	query := fmt.Sprintf(CoreRouteOutput, outputID.ToHex())

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
	query := fmt.Sprintf(CoreRouteOutputWithMetadata, outputID.ToHex())

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
	query := fmt.Sprintf(CoreRouteOutputMetadata, outputID.ToHex())

	res := new(apimodels.OutputMetadata)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentByID gets a commitment details by its ID.
func (client *Client) CommitmentByID(ctx context.Context, id iotago.CommitmentID) (*iotago.Commitment, error) {
	query := fmt.Sprintf(CoreRouteCommitmentByID, id.ToHex())

	res := new(iotago.Commitment)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentUTXOChangesByID returns all UTXO changes of a commitment by its ID.
func (client *Client) CommitmentUTXOChangesByID(ctx context.Context, id iotago.CommitmentID) (*apimodels.UTXOChangesResponse, error) {
	query := fmt.Sprintf(CoreRouteCommitmentByIDUTXOChanges, id.ToHex())

	res := new(apimodels.UTXOChangesResponse)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentByIndex gets a commitment details by its index.
func (client *Client) CommitmentByIndex(ctx context.Context, index iotago.SlotIndex) (*iotago.Commitment, error) {
	query := fmt.Sprintf(CoreRouteCommitmentByIndex, index)

	res := new(iotago.Commitment)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentUTXOChangesByIndex returns all UTXO changes of a commitment by its index.
func (client *Client) CommitmentUTXOChangesByIndex(ctx context.Context, index iotago.SlotIndex) (*apimodels.UTXOChangesResponse, error) {
	query := fmt.Sprintf(CoreRouteCommitmentByIndexUTXOChanges, index)

	res := new(apimodels.UTXOChangesResponse)
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, query, nil, res); err != nil {
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
