package nodeclient

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2/serix"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
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
	RequestHeaderHookAcceptJSON = func(header http.Header) { header.Set("Accept", api.MIMEApplicationJSON) }
	// RequestHeaderHookAcceptIOTASerializerV2 is used to set the request "Accept" header to MIMEApplicationVendorIOTASerializerV2.
	RequestHeaderHookAcceptIOTASerializerV2 = func(header http.Header) { header.Set("Accept", api.MIMEApplicationVendorIOTASerializerV2) }
	// RequestHeaderHookContentTypeIOTASerializerV2 is used to set the request "Content-Type" header to MIMEApplicationVendorIOTASerializerV2.
	RequestHeaderHookContentTypeIOTASerializerV2 = func(header http.Header) { header.Set("Content-Type", api.MIMEApplicationVendorIOTASerializerV2) }
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
		return nil, ierrors.Wrap(err, "unable to call info endpoint for protocol parameter init")
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
	hasPlugin, err := client.NodeSupportsRoute(ctx, api.ManagementPluginName)
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
	hasPlugin, err := client.NodeSupportsRoute(ctx, api.IndexerPluginName)
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
	hasPlugin, err := client.NodeSupportsRoute(ctx, api.MQTTPluginName)
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
	hasPlugin, err := client.NodeSupportsRoute(ctx, api.BlockIssuerPluginName)
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
	if _, err := client.Do(ctx, http.MethodGet, api.RouteHealth, nil, nil); err != nil {
		if ierrors.Is(err, ErrHTTPServiceUnavailable) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// Routes gets the routes the node supports.
func (client *Client) Routes(ctx context.Context) (*api.RoutesResponse, error) {
	res := new(api.RoutesResponse)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, api.RouteRoutes, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// Info gets the info of the node.
func (client *Client) Info(ctx context.Context) (*api.InfoResponse, error) {
	res := new(api.InfoResponse)

	//nolint:bodyclose
	if _, err := do(ctx, iotago.CommonSerixAPI(), client.opts.httpClient, client.BaseURL, client.opts.userInfo, http.MethodGet, api.CoreRouteInfo, client.opts.requestURLHook, nil, nil, res); err != nil {
		return nil, err
	}

	client.apiProvider.SetCommittedSlot(res.Status.LatestCommitmentID.Slot())

	return res, nil
}

// NetworkHealth returns whether the network is healthy (finalization is not delayed).
func (client *Client) NetworkHealth(ctx context.Context) (bool, error) {
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, api.CoreRouteNetworkHealth, nil, nil); err != nil {
		if ierrors.Is(err, ErrHTTPServiceUnavailable) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// NetworkMetrics gets the current network metrics.
func (client *Client) NetworkMetrics(ctx context.Context) (*api.NetworkMetricsResponse, error) {
	res := new(api.NetworkMetricsResponse)

	//nolint:bodyclose
	if _, err := do(ctx, iotago.CommonSerixAPI(), client.opts.httpClient, client.BaseURL, client.opts.userInfo, http.MethodGet, api.CoreRouteNetworkMetrics, client.opts.requestURLHook, nil, nil, res); err != nil {
		return nil, err
	}

	return res, nil
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
	res, err := client.DoWithRequestHeaderHook(ctx, http.MethodPost, api.CoreRouteBlocks, RequestHeaderHookContentTypeIOTASerializerV2, req, nil)
	if err != nil {
		return iotago.EmptyBlockID, err
	}

	blockID, err := iotago.BlockIDFromHexString(res.Header.Get(locationHeader))
	if err != nil {
		return iotago.EmptyBlockID, err
	}

	return blockID, nil
}

// BlockByBlockID get a block by its block ID from the node.
func (client *Client) BlockByBlockID(ctx context.Context, blockID iotago.BlockID) (*iotago.Block, error) {
	query := client.endpointReplaceBlockIDParameter(api.CoreRouteBlock, blockID)

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

// BlockMetadataByBlockID gets the metadata of a block by its ID from the node.
func (client *Client) BlockMetadataByBlockID(ctx context.Context, blockID iotago.BlockID) (*api.BlockMetadataResponse, error) {
	query := client.endpointReplaceBlockIDParameter(api.CoreRouteBlockMetadata, blockID)

	res := new(api.BlockMetadataResponse)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// BlockWithMetadataByBlockID gets a block by its ID, together with the metadata from the node.
func (client *Client) BlockWithMetadataByBlockID(ctx context.Context, blockID iotago.BlockID) (*api.BlockWithMetadataResponse, error) {
	query := client.endpointReplaceBlockIDParameter(api.CoreRouteBlockWithMetadata, blockID)

	res := new(RawDataEnvelope)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptIOTASerializerV2, nil, res); err != nil {
		return nil, err
	}

	block, consumedBytes, err := iotago.BlockFromBytes(client)(res.Data)
	if err != nil {
		return nil, err
	}

	metadata := new(api.BlockMetadataResponse)
	_, err = client.APIForSlot(blockID.Slot()).Decode(res.Data[consumedBytes:], metadata)
	if err != nil {
		return nil, err
	}

	return &api.BlockWithMetadataResponse{
		Block:    block,
		Metadata: metadata,
	}, nil
}

// BlockIssuance gets the info to issue a block.
func (client *Client) BlockIssuance(ctx context.Context) (*api.IssuanceBlockHeaderResponse, error) {
	res := new(api.IssuanceBlockHeaderResponse)

	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, api.CoreRouteBlockIssuance, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// OutputByID gets an output by its ID from the node.
func (client *Client) OutputByID(ctx context.Context, outputID iotago.OutputID) (iotago.Output, error) {
	query := client.endpointReplaceOutputIDParameter(api.CoreRouteOutput, outputID)

	res := new(RawDataEnvelope)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptIOTASerializerV2, nil, res); err != nil {
		return nil, err
	}

	var outputResponse api.OutputResponse
	if _, err := client.CommittedAPI().Decode(res.Data, &outputResponse, serix.WithValidation()); err != nil {
		return nil, err
	}

	derivedOutputID, err := outputResponse.OutputIDProof.OutputID(outputResponse.Output)
	if err != nil {
		return nil, err
	}

	if derivedOutputID != outputID {
		return nil, ierrors.Errorf("requested output ID %s does not match computed output ID %s", outputID.ToHex(), derivedOutputID.ToHex())
	}

	return outputResponse.Output, nil
}

// OutputMetadataByID gets an output's metadata by its ID from the node without getting the output data again.
func (client *Client) OutputMetadataByID(ctx context.Context, outputID iotago.OutputID) (*api.OutputMetadata, error) {
	query := client.endpointReplaceOutputIDParameter(api.CoreRouteOutputMetadata, outputID)

	res := new(api.OutputMetadata)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// OutputWithMetadataByID gets an output by its ID, together with the metadata from the node.
func (client *Client) OutputWithMetadataByID(ctx context.Context, outputID iotago.OutputID) (iotago.Output, *api.OutputMetadata, error) {
	query := client.endpointReplaceOutputIDParameter(api.CoreRouteOutputWithMetadata, outputID)

	res := new(RawDataEnvelope)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptIOTASerializerV2, nil, res); err != nil {
		return nil, nil, err
	}

	var outputResponse api.OutputWithMetadataResponse
	if _, err := client.CommittedAPI().Decode(res.Data, &outputResponse, serix.WithValidation()); err != nil {
		return nil, nil, err
	}

	derivedOutputID, err := outputResponse.OutputIDProof.OutputID(outputResponse.Output)
	if err != nil {
		return nil, nil, err
	}

	if derivedOutputID != outputID {
		return nil, nil, ierrors.Errorf("requested output ID %s does not match computed output ID %s", outputID.ToHex(), derivedOutputID.ToHex())
	}

	return outputResponse.Output, outputResponse.Metadata, nil
}

// TransactionByID gets a transaction by its ID from the node.
func (client *Client) TransactionByID(ctx context.Context, txID iotago.TransactionID) (*iotago.Transaction, error) {
	query := client.endpointReplaceTransactionIDParameter(api.CoreRouteTransaction, txID)

	res := new(RawDataEnvelope)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptIOTASerializerV2, nil, res); err != nil {
		return nil, err
	}

	tx := new(iotago.Transaction)
	if _, err := client.CommittedAPI().Decode(res.Data, tx, serix.WithValidation()); err != nil {
		return nil, err
	}

	return tx, nil
}

// TransactionIncludedBlock get a block that included the given transaction ID in the ledger.
func (client *Client) TransactionIncludedBlock(ctx context.Context, txID iotago.TransactionID) (*iotago.Block, error) {
	query := client.endpointReplaceTransactionIDParameter(api.CoreRouteTransactionsIncludedBlock, txID)

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
func (client *Client) TransactionIncludedBlockMetadata(ctx context.Context, txID iotago.TransactionID) (*api.BlockMetadataResponse, error) {
	query := client.endpointReplaceTransactionIDParameter(api.CoreRouteTransactionsIncludedBlockMetadata, txID)

	res := new(api.BlockMetadataResponse)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// TransactionMetadata gets the metadata of a transaction by its ID from the node.
func (client *Client) TransactionMetadata(ctx context.Context, txID iotago.TransactionID) (*api.TransactionMetadataResponse, error) {
	query := client.endpointReplaceTransactionIDParameter(api.CoreRouteTransactionsMetadata, txID)

	res := new(api.TransactionMetadataResponse)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentByID gets a commitment details by its ID.
func (client *Client) CommitmentByID(ctx context.Context, commitmentID iotago.CommitmentID) (*iotago.Commitment, error) {
	query := client.endpointReplaceCommitmentIDParameter(api.CoreRouteCommitmentByID, commitmentID)

	res := new(iotago.Commitment)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentUTXOChangesByID returns all UTXO changes of a commitment by its ID.
func (client *Client) CommitmentUTXOChangesByID(ctx context.Context, commitmentID iotago.CommitmentID) (*api.UTXOChangesResponse, error) {
	query := client.endpointReplaceCommitmentIDParameter(api.CoreRouteCommitmentByIDUTXOChanges, commitmentID)

	res := new(api.UTXOChangesResponse)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentUTXOChangesFullByID returns all UTXO changes (including outputs) of a commitment by its ID.
func (client *Client) CommitmentUTXOChangesFullByID(ctx context.Context, commitmentID iotago.CommitmentID) (*api.UTXOChangesFullResponse, error) {
	query := client.endpointReplaceCommitmentIDParameter(api.CoreRouteCommitmentByIDUTXOChangesFull, commitmentID)

	res := new(api.UTXOChangesFullResponse)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentBySlot gets a commitment details by its slot.
func (client *Client) CommitmentBySlot(ctx context.Context, slot iotago.SlotIndex) (*iotago.Commitment, error) {
	query := client.endpointReplaceSlotParameter(api.CoreRouteCommitmentBySlot, slot)

	res := new(iotago.Commitment)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentUTXOChangesBySlot returns all UTXO changes of a commitment by its slot.
func (client *Client) CommitmentUTXOChangesBySlot(ctx context.Context, slot iotago.SlotIndex) (*api.UTXOChangesResponse, error) {
	query := client.endpointReplaceSlotParameter(api.CoreRouteCommitmentBySlotUTXOChanges, slot)

	res := new(api.UTXOChangesResponse)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// CommitmentUTXOChangesFullBySlot returns all UTXO changes (including outputs) of a commitment by its slot.
func (client *Client) CommitmentUTXOChangesFullBySlot(ctx context.Context, slot iotago.SlotIndex) (*api.UTXOChangesFullResponse, error) {
	query := client.endpointReplaceSlotParameter(api.CoreRouteCommitmentBySlotUTXOChangesFull, slot)

	res := new(api.UTXOChangesFullResponse)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// Congestion gets the congestion of the node.
func (client *Client) Congestion(ctx context.Context, accountAddress *iotago.AccountAddress, workScore iotago.WorkScore, optCommitmentID ...iotago.CommitmentID) (*api.CongestionResponse, error) {
	//nolint:contextcheck
	query := client.endpointReplaceAddressParameter(api.CoreRouteCongestion, accountAddress)
	queryParams := url.Values{}

	if workScore > 0 {
		queryParams.Add(api.ParameterWorkScore, strconv.FormatUint(uint64(workScore), 10))
	}
	if len(optCommitmentID) > 0 {
		queryParams.Add(api.ParameterCommitmentID, optCommitmentID[0].ToHex())
	}

	queryWithParams, err := encodeURLWithQueryParams(query, queryParams)
	if err != nil {
		return nil, err
	}

	res := new(api.CongestionResponse)

	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, queryWithParams, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// Validators gets the validators from the node with the given page size and cursor.
func (client *Client) Validators(ctx context.Context, pageSize uint64, cursor ...string) (*api.ValidatorsResponse, error) {
	res := new(api.ValidatorsResponse)
	query := api.CoreRouteValidators
	queryParams := url.Values{}

	if pageSize > 0 {
		queryParams.Add(api.ParameterPageSize, strconv.FormatUint(pageSize, 10))
	}

	if len(cursor) > 0 {
		queryParams.Add(api.ParameterCursor, cursor[0])
	}

	queryWithParams, err := encodeURLWithQueryParams(query, queryParams)
	if err != nil {
		return nil, err
	}

	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, queryWithParams, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// ValidatorsAll gets all validators from the node.
func (client *Client) ValidatorsAll(ctx context.Context, maxPages ...int) (validators *api.ValidatorsResponse, allRetrieved bool, err error) {
	validatorsResponses := make([]*api.ValidatorResponse, 0)
	resp, err := client.Validators(ctx, 0)
	if err != nil {
		return nil, false, err
	}
	validatorsResponses = append(validatorsResponses, resp.Validators...)

	cursor := resp.Cursor
	for count := 1; cursor != ""; count++ {
		if len(maxPages) > 0 && count >= maxPages[0] {
			return &api.ValidatorsResponse{Validators: validatorsResponses}, false, nil
		}
		resp, err = client.Validators(ctx, 0, cursor)
		if err != nil {
			return nil, false, err
		}
		validatorsResponses = append(validatorsResponses, resp.Validators...)

		cursor = resp.Cursor
	}

	return &api.ValidatorsResponse{Validators: validatorsResponses}, true, nil
}

// Validator gets the validator response of the given account address.
func (client *Client) Validator(ctx context.Context, accountAddress *iotago.AccountAddress) (*api.ValidatorResponse, error) {
	res := new(api.ValidatorResponse)

	//nolint:contextcheck
	query := client.endpointReplaceAddressParameter(api.CoreRouteValidatorsAccount, accountAddress)

	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// Rewards gets the mana rewards of the given output.
func (client *Client) Rewards(ctx context.Context, outputID iotago.OutputID) (*api.ManaRewardsResponse, error) {
	query := client.endpointReplaceOutputIDParameter(api.CoreRouteRewards, outputID)

	res := new(api.ManaRewardsResponse)
	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, query, RequestHeaderHookAcceptJSON, nil, res); err != nil {
		return nil, err
	}

	return res, nil
}

// Committee gets the committee of the given epoch index.
func (client *Client) Committee(ctx context.Context, optEpochIndex ...iotago.EpochIndex) (*api.CommitteeResponse, error) {
	query := api.CoreRouteCommittee
	queryParams := url.Values{}

	if len(optEpochIndex) > 0 {
		queryParams.Add(api.ParameterEpoch, strconv.FormatUint(uint64(optEpochIndex[0]), 10))
	}

	queryWithParams, err := encodeURLWithQueryParams(query, queryParams)
	if err != nil {
		return nil, err
	}

	res := new(api.CommitteeResponse)

	//nolint:bodyclose
	if _, err := client.DoWithRequestHeaderHook(ctx, http.MethodGet, queryWithParams, RequestHeaderHookAcceptJSON, nil, res); err != nil {
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
		if string(p) == route {
			return true, nil
		}
	}

	return false, nil
}

func (client *Client) endpointReplaceAddressParameter(endpoint string, address iotago.Address) string {
	return api.EndpointWithNamedParameterValue(endpoint, api.ParameterBech32Address, address.Bech32(client.CommittedAPI().ProtocolParameters().Bech32HRP()))
}

func (client *Client) endpointReplaceBlockIDParameter(endpoint string, blockID iotago.BlockID) string {
	return api.EndpointWithNamedParameterValue(endpoint, api.ParameterBlockID, blockID.ToHex())
}

func (client *Client) endpointReplaceTransactionIDParameter(endpoint string, txID iotago.TransactionID) string {
	return api.EndpointWithNamedParameterValue(endpoint, api.ParameterTransactionID, txID.ToHex())
}

func (client *Client) endpointReplaceOutputIDParameter(endpoint string, outputID iotago.OutputID) string {
	return api.EndpointWithNamedParameterValue(endpoint, api.ParameterOutputID, outputID.ToHex())
}

func (client *Client) endpointReplaceSlotParameter(endpoint string, slot iotago.SlotIndex) string {
	return api.EndpointWithNamedParameterValue(endpoint, api.ParameterSlot, strconv.Itoa(int(slot)))
}

func (client *Client) endpointReplaceCommitmentIDParameter(endpoint string, commitmentID iotago.CommitmentID) string {
	return api.EndpointWithNamedParameterValue(endpoint, api.ParameterCommitmentID, commitmentID.ToHex())
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
