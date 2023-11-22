package nodeclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/hexutil"
	"github.com/iotaledger/iota.go/v4/nodeclient/apimodels"
)

const (
	IndexerEndpointOutputs           = "/outputs"
	IndexerEndpointBasicOutputs      = "/outputs/basic"
	IndexerEndpointAccounts          = "/outputs/account"
	IndexerEndpointAccount           = "/outputs/account/%s"
	IndexerEndpointAnchors           = "/outputs/anchor"
	IndexerEndpointAnchor            = "/outputs/anchor/%s"
	IndexerEndpointFoundries         = "/outputs/foundry"
	IndexerEndpointFoundry           = "/outputs/foundry/%s"
	IndexerEndpointNFTs              = "/outputs/nft"
	IndexerEndpointNFT               = "/outputs/nft/%s"
	IndexerEndpointDelegationOutputs = "/outputs/delegation"
	IndexerEndpointDelegationOutput  = "/outputs/delegation/%s"
)

var (
	IndexerRouteOutputs           = route(IndexerPluginName, IndexerEndpointOutputs)
	IndexerRouteBasicOutputs      = route(IndexerPluginName, IndexerEndpointBasicOutputs)
	IndexerRouteAccounts          = route(IndexerPluginName, IndexerEndpointAccounts)
	IndexerRouteAccount           = route(IndexerPluginName, IndexerEndpointAccount)
	IndexerRouteAnchors           = route(IndexerPluginName, IndexerEndpointAnchors)
	IndexerRouteAnchor            = route(IndexerPluginName, IndexerEndpointAnchor)
	IndexerRouteFoundries         = route(IndexerPluginName, IndexerEndpointFoundries)
	IndexerRouteFoundry           = route(IndexerPluginName, IndexerEndpointFoundry)
	IndexerRouteNFTs              = route(IndexerPluginName, IndexerEndpointNFTs)
	IndexerRouteNFT               = route(IndexerPluginName, IndexerEndpointNFT)
	IndexerRouteDelegationOutputs = route(IndexerPluginName, IndexerEndpointDelegationOutputs)
	IndexerRouteDelegationOutput  = route(IndexerPluginName, IndexerEndpointDelegationOutput)
)

var (
	// ErrIndexerNotFound gets returned when the indexer doesn't find any result.
	// Only applicable to single element queries.
	ErrIndexerNotFound = ierrors.New("no result found")
)

type (

	// IndexerClient is a client which queries the optional indexer functionality of a node.
	IndexerClient interface {
		// Outputs returns a handle to query for outputs.
		Outputs(ctx context.Context, query IndexerQuery) (*IndexerResultSet, error)
		// Account queries for a specific iotago.AccountOutput by its identifier and returns the ledger index at which this output where available at.
		Account(ctx context.Context, accountID iotago.AccountID) (*iotago.OutputID, *iotago.AccountOutput, iotago.SlotIndex, error)
		// Anchor queries for a specific iotago.AnchorOutput by its identifier and returns the ledger index at which this output where available at.
		Anchor(ctx context.Context, anchorID iotago.AnchorID) (*iotago.OutputID, *iotago.AnchorOutput, iotago.SlotIndex, error)
		// Foundry queries for a specific iotago.FoundryOutput by its identifier and returns the ledger index at which this output where available at.
		Foundry(ctx context.Context, foundryID iotago.FoundryID) (*iotago.OutputID, *iotago.FoundryOutput, iotago.SlotIndex, error)
		// NFT queries for a specific iotago.NFTOutput by its identifier and returns the ledger index at which this output where available at.
		NFT(ctx context.Context, nftID iotago.NFTID) (*iotago.OutputID, *iotago.NFTOutput, iotago.SlotIndex, error)
		// Delegation queries for a specific iotago.DelegationOutout by its identifier and returns the ledger index at which this output where available at.
		Delegation(ctx context.Context, delegationID iotago.DelegationID) (*iotago.OutputID, *iotago.DelegationOutput, iotago.SlotIndex, error)
	}

	// IndexerQuery is a query executed against the indexer.
	IndexerQuery interface {
		// SetOffset sets the offset for the query.
		SetOffset(offset *string)
		// URLParams returns the query parameters as URL encoded query parameters.
		URLParams() (string, error)
	}

	indexerClient struct {
		core *Client
	}
)

// IndexerResultSet is a handle for indexer queries.
type IndexerResultSet struct {
	client         *Client
	query          IndexerQuery
	firstQueryDone bool
	nextFunc       func() error
	// The error which has occurred during querying.
	Error error
	// The response from the indexer after calling Next().
	Response *apimodels.IndexerResponse
}

// Next runs the next query against the indexer.
// Returns false if there are no more results to collect.
func (resultSet *IndexerResultSet) Next() bool {
	if resultSet.firstQueryDone && resultSet.Response.Cursor == "" {
		return false
	}

	if err := resultSet.nextFunc(); err != nil {
		resultSet.Error = err

		return false
	}

	// set offset for next query
	resultSet.query.SetOffset(&resultSet.Response.Cursor)
	resultSet.firstQueryDone = true

	return len(resultSet.Response.Items) > 0
}

// Outputs collects/fetches the outputs result from the query.
func (resultSet *IndexerResultSet) Outputs(ctx context.Context) (iotago.Outputs[iotago.Output], error) {
	outputIDs := resultSet.Response.Items.MustOutputIDs()
	outputs := make(iotago.Outputs[iotago.Output], len(outputIDs))
	for i, outputID := range outputIDs {
		output, err := resultSet.client.OutputByID(ctx, outputID)
		if err != nil {
			return nil, ierrors.Errorf("unable to fetch output %s: %w", outputID.ToHex(), err)
		}
		outputs[i] = output
	}

	return outputs, nil
}

// Do executes a request against the endpoint.
// This function is only meant to be used for special routes not covered through the standard API.
func (client *indexerClient) Do(ctx context.Context, method string, route string, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	return client.core.Do(ctx, method, route, reqObj, resObj)
}

// DoWithRequestHeaderHook executes a request against the endpoint.
// This function is only meant to be used for special routes not covered through the standard API.
func (client *indexerClient) DoWithRequestHeaderHook(ctx context.Context, method string, route string, requestHeaderHook RequestHeaderHook, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	return client.core.DoWithRequestHeaderHook(ctx, method, route, requestHeaderHook, reqObj, resObj)
}

func (client *indexerClient) Outputs(ctx context.Context, query IndexerQuery) (*IndexerResultSet, error) {
	res := &IndexerResultSet{
		client: client.core,
		query:  query,
	}

	var baseRoute string
	switch query.(type) {
	case *apimodels.OutputsQuery:
		baseRoute = IndexerRouteOutputs
	case *apimodels.BasicOutputsQuery:
		baseRoute = IndexerRouteBasicOutputs
	case *apimodels.AccountsQuery:
		baseRoute = IndexerRouteAccounts
	case *apimodels.AnchorsQuery:
		baseRoute = IndexerRouteAnchors
	case *apimodels.FoundriesQuery:
		baseRoute = IndexerRouteFoundries
	case *apimodels.NFTsQuery:
		baseRoute = IndexerRouteNFTs
	case *apimodels.DelegationOutputsQuery:
		baseRoute = IndexerRouteDelegationOutputs
	default:
		return nil, ierrors.Errorf("unsupported query type: %T", query)
	}

	// this gets executed on every Next()
	nextFunc := func() error {
		res.Response = &apimodels.IndexerResponse{}

		urlParams, err := query.URLParams()
		if err != nil {
			return err
		}

		routeWithParams := fmt.Sprintf("%s?%s", baseRoute, urlParams)
		//nolint:bodyclose
		_, reqErr := client.Do(ctx, http.MethodGet, routeWithParams, nil, res.Response)

		return reqErr
	}
	res.nextFunc = nextFunc

	return res, nil
}

func (client *indexerClient) singleOutputQuery(ctx context.Context, route string) (*iotago.OutputID, iotago.Output, iotago.SlotIndex, error) {
	res := &apimodels.IndexerResponse{}
	//nolint:bodyclose
	if _, err := client.Do(ctx, http.MethodGet, route, nil, res); err != nil {
		return nil, nil, 0, err
	}

	if len(res.Items) == 0 {
		return nil, nil, res.CommittedSlot, ierrors.Errorf("%w for route %s", ErrIndexerNotFound, route)
	}

	outputID := res.Items.MustOutputIDs()[0]
	output, err := client.core.OutputByID(ctx, outputID)
	if err != nil {
		return nil, nil, res.CommittedSlot, err
	}

	return &outputID, output, res.CommittedSlot, err
}

func (client *indexerClient) Account(ctx context.Context, accountID iotago.AccountID) (*iotago.OutputID, *iotago.AccountOutput, iotago.SlotIndex, error) {
	outputID, output, ledgerIndex, err := client.singleOutputQuery(ctx, fmt.Sprintf(IndexerRouteAccount, hexutil.EncodeHex(accountID[:])))
	if err != nil {
		return nil, nil, ledgerIndex, err
	}

	//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
	return outputID, output.(*iotago.AccountOutput), ledgerIndex, nil
}

func (client *indexerClient) Anchor(ctx context.Context, anchorID iotago.AnchorID) (*iotago.OutputID, *iotago.AnchorOutput, iotago.SlotIndex, error) {
	outputID, output, ledgerIndex, err := client.singleOutputQuery(ctx, fmt.Sprintf(IndexerRouteAnchor, hexutil.EncodeHex(anchorID[:])))
	if err != nil {
		return nil, nil, ledgerIndex, err
	}

	//nolint:forcetypeassert // we can safely assume that this is an AnchorOutput
	return outputID, output.(*iotago.AnchorOutput), ledgerIndex, nil
}

func (client *indexerClient) Foundry(ctx context.Context, foundryID iotago.FoundryID) (*iotago.OutputID, *iotago.FoundryOutput, iotago.SlotIndex, error) {
	outputID, output, ledgerIndex, err := client.singleOutputQuery(ctx, fmt.Sprintf(IndexerRouteFoundry, hexutil.EncodeHex(foundryID[:])))
	if err != nil {
		return nil, nil, ledgerIndex, err
	}

	//nolint:forcetypeassert // we can safely assume that this is an FoundryOutput
	return outputID, output.(*iotago.FoundryOutput), ledgerIndex, nil
}

func (client *indexerClient) NFT(ctx context.Context, nftID iotago.NFTID) (*iotago.OutputID, *iotago.NFTOutput, iotago.SlotIndex, error) {
	outputID, output, ledgerIndex, err := client.singleOutputQuery(ctx, fmt.Sprintf(IndexerRouteNFT, hexutil.EncodeHex(nftID[:])))
	if err != nil {
		return nil, nil, ledgerIndex, err
	}

	//nolint:forcetypeassert // we can safely assume that this is an NFTOutput
	return outputID, output.(*iotago.NFTOutput), ledgerIndex, nil
}

func (client *indexerClient) Delegation(ctx context.Context, delegationID iotago.DelegationID) (*iotago.OutputID, *iotago.DelegationOutput, iotago.SlotIndex, error) {
	outputID, output, ledgerIndex, err := client.singleOutputQuery(ctx, fmt.Sprintf(IndexerRouteDelegationOutput, hexutil.EncodeHex(delegationID[:])))
	if err != nil {
		return nil, nil, ledgerIndex, err
	}

	//nolint:forcetypeassert // we can safely assume that this is a DelegationOutput
	return outputID, output.(*iotago.DelegationOutput), ledgerIndex, nil
}
