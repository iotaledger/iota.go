package nodeclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/api"
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
		// Account queries for a specific iotago.AccountOutput by its address and returns the ledger index at which this output where available at.
		Account(ctx context.Context, accountAddress *iotago.AccountAddress) (*iotago.OutputID, *iotago.AccountOutput, iotago.SlotIndex, error)
		// Anchor queries for a specific iotago.AnchorOutput by its address and returns the ledger index at which this output where available at.
		Anchor(ctx context.Context, anchorAddress *iotago.AnchorAddress) (*iotago.OutputID, *iotago.AnchorOutput, iotago.SlotIndex, error)
		// Foundry queries for a specific iotago.FoundryOutput by its identifier and returns the ledger index at which this output where available at.
		Foundry(ctx context.Context, foundryID iotago.FoundryID) (*iotago.OutputID, *iotago.FoundryOutput, iotago.SlotIndex, error)
		// NFT queries for a specific iotago.NFTOutput by its address and returns the ledger index at which this output where available at.
		NFT(ctx context.Context, nftAddress *iotago.NFTAddress) (*iotago.OutputID, *iotago.NFTOutput, iotago.SlotIndex, error)
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
	Response *api.IndexerResponse
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
	case *api.OutputsQuery:
		baseRoute = api.IndexerRouteOutputs
	case *api.BasicOutputsQuery:
		baseRoute = api.IndexerRouteOutputsBasic
	case *api.AccountsQuery:
		baseRoute = api.IndexerRouteOutputsAccounts
	case *api.AnchorsQuery:
		baseRoute = api.IndexerRouteOutputsAnchors
	case *api.FoundriesQuery:
		baseRoute = api.IndexerRouteOutputsFoundries
	case *api.NFTsQuery:
		baseRoute = api.IndexerRouteOutputsNFTs
	case *api.DelegationOutputsQuery:
		baseRoute = api.IndexerRouteOutputsDelegations
	default:
		return nil, ierrors.Errorf("unsupported query type: %T", query)
	}

	// this gets executed on every Next()
	nextFunc := func() error {
		res.Response = &api.IndexerResponse{}

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
	res := &api.IndexerResponse{}
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

func (client *indexerClient) Account(ctx context.Context, accountAddress *iotago.AccountAddress) (*iotago.OutputID, *iotago.AccountOutput, iotago.SlotIndex, error) {
	outputID, output, ledgerIndex, err := client.singleOutputQuery(ctx, client.core.endpointReplaceAddressParameter(api.IndexerRouteOutputsAccountByAddress, accountAddress))
	if err != nil {
		return nil, nil, ledgerIndex, err
	}

	//nolint:forcetypeassert // we can safely assume that this is an AccountOutput
	return outputID, output.(*iotago.AccountOutput), ledgerIndex, nil
}

func (client *indexerClient) Anchor(ctx context.Context, anchorAddress *iotago.AnchorAddress) (*iotago.OutputID, *iotago.AnchorOutput, iotago.SlotIndex, error) {
	outputID, output, ledgerIndex, err := client.singleOutputQuery(ctx, client.core.endpointReplaceAddressParameter(api.IndexerRouteOutputsAnchorByAddress, anchorAddress))
	if err != nil {
		return nil, nil, ledgerIndex, err
	}

	//nolint:forcetypeassert // we can safely assume that this is an AnchorOutput
	return outputID, output.(*iotago.AnchorOutput), ledgerIndex, nil
}

func (client *indexerClient) Foundry(ctx context.Context, foundryID iotago.FoundryID) (*iotago.OutputID, *iotago.FoundryOutput, iotago.SlotIndex, error) {
	outputID, output, ledgerIndex, err := client.singleOutputQuery(ctx, api.EndpointWithNamedParameterValue(api.IndexerRouteOutputsFoundryByID, api.ParameterFoundryID, foundryID.ToHex()))
	if err != nil {
		return nil, nil, ledgerIndex, err
	}

	//nolint:forcetypeassert // we can safely assume that this is an FoundryOutput
	return outputID, output.(*iotago.FoundryOutput), ledgerIndex, nil
}

func (client *indexerClient) NFT(ctx context.Context, nftAddress *iotago.NFTAddress) (*iotago.OutputID, *iotago.NFTOutput, iotago.SlotIndex, error) {
	outputID, output, ledgerIndex, err := client.singleOutputQuery(ctx, client.core.endpointReplaceAddressParameter(api.IndexerRouteOutputsNFTByAddress, nftAddress))
	if err != nil {
		return nil, nil, ledgerIndex, err
	}

	//nolint:forcetypeassert // we can safely assume that this is an NFTOutput
	return outputID, output.(*iotago.NFTOutput), ledgerIndex, nil
}

func (client *indexerClient) Delegation(ctx context.Context, delegationID iotago.DelegationID) (*iotago.OutputID, *iotago.DelegationOutput, iotago.SlotIndex, error) {
	outputID, output, ledgerIndex, err := client.singleOutputQuery(ctx, api.EndpointWithNamedParameterValue(api.IndexerRouteOutputsDelegationByID, api.ParameterDelegationID, delegationID.ToHex()))
	if err != nil {
		return nil, nil, ledgerIndex, err
	}

	//nolint:forcetypeassert // we can safely assume that this is a DelegationOutput
	return outputID, output.(*iotago.DelegationOutput), ledgerIndex, nil
}
