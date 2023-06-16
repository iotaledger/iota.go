package nodeclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

// Indexer plugin routes.
var (
	IndexerAPIRouteBasicOutputs = "/api/" + IndexerPluginName + "/outputs/basic"
	IndexerAPIRouteAccounts     = "/api/" + IndexerPluginName + "/outputs/account"
	IndexerAPIRouteAccount      = "/api/" + IndexerPluginName + "/outputs/account/%s"
	IndexerAPIRouteFoundries    = "/api/" + IndexerPluginName + "/outputs/foundry"
	IndexerAPIRouteFoundry      = "/api/" + IndexerPluginName + "/outputs/foundry/%s"
	IndexerAPIRouteNFTs         = "/api/" + IndexerPluginName + "/outputs/nft"
	IndexerAPIRouteNFT          = "/api/" + IndexerPluginName + "/outputs/nft/%s"
)

var (
	// ErrIndexerNotFound gets returned when the indexer doesn't find any result.
	// Only applicable to single element queries.
	ErrIndexerNotFound = errors.New("no result found")

	outputTypeToIndexerRoute = map[iotago.OutputType]string{
		iotago.OutputBasic:   IndexerAPIRouteBasicOutputs,
		iotago.OutputAccount: IndexerAPIRouteAccounts,
		iotago.OutputFoundry: IndexerAPIRouteFoundries,
		iotago.OutputNFT:     IndexerAPIRouteNFTs,
	}
)

type (

	// IndexerClient is a client which queries the optional indexer functionality of a node.
	IndexerClient interface {
		// Outputs returns a handle to query for outputs.
		Outputs(ctx context.Context, query IndexerQuery) (*IndexerResultSet, error)
		// Account queries for a specific iotago.AccountOutput by its identifier and returns the ledger index at which this output where available at.
		Account(ctx context.Context, accountID iotago.AccountID) (*iotago.OutputID, *iotago.AccountOutput, iotago.SlotIndex, error)
		// Foundry queries for a specific iotago.FoundryOutput by its identifier and returns the ledger index at which this output where available at.
		Foundry(ctx context.Context, foundryID iotago.FoundryID) (*iotago.OutputID, *iotago.FoundryOutput, iotago.SlotIndex, error)
		// NFT queries for a specific iotago.NFTOutput by its identifier and returns the ledger index at which this output where available at.
		NFT(ctx context.Context, nftID iotago.NFTID) (*iotago.OutputID, *iotago.NFTOutput, iotago.SlotIndex, error)
	}

	// IndexerQuery is a query executed against the indexer.
	IndexerQuery interface {
		// SetOffset sets the offset for the query.
		SetOffset(offset *string)
		// URLiotago.TxEssenceInputs returns the query parameters as URL encoded query parameters.
		URLParams() (string, error)
		// OutputType returns the output type for which the query is for.
		OutputType() iotago.OutputType
	}

	indexerClient struct {
		core *Client
	}
)

// IndexerResultSet is a handle for indexer queries.
type IndexerResultSet struct {
	client         *Client
	ctx            context.Context
	query          IndexerQuery
	firstQueryDone bool
	nextFunc       func() error
	// The error which has occurred during querying.
	Error error
	// The response from the indexer after calling Next().
	Response *IndexerResponse
}

// Next runs the next query against the indexer.
// Returns false if there are no more results to collect.
func (resultSet *IndexerResultSet) Next() bool {
	if resultSet.firstQueryDone && resultSet.Response.Cursor == nil {
		return false
	}

	if err := resultSet.nextFunc(); err != nil {
		resultSet.Error = err
		return false
	}

	// set offset for next query
	resultSet.query.SetOffset(resultSet.Response.Cursor)
	resultSet.firstQueryDone = true

	return len(resultSet.Response.Items) > 0
}

// Outputs collects/fetches the outputs result from the query.
func (resultSet *IndexerResultSet) Outputs() (iotago.Outputs[iotago.Output], error) {
	outputIDs := resultSet.Response.Items.MustOutputIDs()
	outputs := make(iotago.Outputs[iotago.Output], len(outputIDs))
	for i, outputID := range outputIDs {
		output, err := resultSet.client.OutputByID(resultSet.ctx, outputID)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch output %s: %w", outputID.ToHex(), err)
		}
		outputs[i] = output
	}
	return outputs, nil
}

// Do executes a request against the endpoint.
// This function is only meant to be used for special routes not covered through the standard API.
func (client *indexerClient) Do(ctx context.Context, method string, route string, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	return do(ctx, client.core.opts.httpClient, client.core.BaseURL, client.core.opts.userInfo, method, route, client.core.opts.requestURLHook, nil, reqObj, resObj)
}

// DoWithRequestHeaderHook executes a request against the endpoint.
// This function is only meant to be used for special routes not covered through the standard API.
func (client *indexerClient) DoWithRequestHeaderHook(ctx context.Context, method string, route string, requestHeaderHook RequestHeaderHook, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	return do(ctx, client.core.opts.httpClient, client.core.BaseURL, client.core.opts.userInfo, method, route, client.core.opts.requestURLHook, requestHeaderHook, reqObj, resObj)
}

func (client *indexerClient) Outputs(ctx context.Context, query IndexerQuery) (*IndexerResultSet, error) {
	res := &IndexerResultSet{
		ctx:    ctx,
		client: client.core,
		query:  query,
	}

	baseRoute := outputTypeToIndexerRoute[query.OutputType()]

	// this gets executed on every Next()
	nextFunc := func() error {
		res.Response = &IndexerResponse{}

		urlParams, err := query.URLParams()
		if err != nil {
			return err
		}

		routeWithParams := fmt.Sprintf("%s?%s", baseRoute, urlParams)
		_, reqErr := client.Do(ctx, http.MethodGet, routeWithParams, nil, res.Response)
		return reqErr
	}
	res.nextFunc = nextFunc

	return res, nil
}

func (client *indexerClient) singleOutputQuery(ctx context.Context, route string) (*iotago.OutputID, iotago.Output, iotago.SlotIndex, error) {
	res := &IndexerResponse{}
	if _, err := client.Do(ctx, http.MethodGet, route, nil, res); err != nil {
		return nil, nil, 0, err
	}

	if len(res.Items) == 0 {
		return nil, nil, res.LedgerIndex, fmt.Errorf("%w for route %s", ErrIndexerNotFound, route)
	}

	outputID := res.Items.MustOutputIDs()[0]
	output, err := client.core.OutputByID(ctx, outputID)
	if err != nil {
		return nil, nil, res.LedgerIndex, err
	}
	return &outputID, output, res.LedgerIndex, err
}

func (client *indexerClient) Account(ctx context.Context, accountID iotago.AccountID) (*iotago.OutputID, *iotago.AccountOutput, iotago.SlotIndex, error) {
	outputID, output, ledgerIndex, err := client.singleOutputQuery(ctx, fmt.Sprintf(IndexerAPIRouteAccount, hexutil.EncodeHex(accountID[:])))
	if err != nil {
		return nil, nil, ledgerIndex, err
	}
	return outputID, output.(*iotago.AccountOutput), ledgerIndex, nil
}

func (client *indexerClient) Foundry(ctx context.Context, foundryID iotago.FoundryID) (*iotago.OutputID, *iotago.FoundryOutput, iotago.SlotIndex, error) {
	outputID, output, ledgerIndex, err := client.singleOutputQuery(ctx, fmt.Sprintf(IndexerAPIRouteFoundry, hexutil.EncodeHex(foundryID[:])))
	if err != nil {
		return nil, nil, ledgerIndex, err
	}
	return outputID, output.(*iotago.FoundryOutput), ledgerIndex, nil
}

func (client *indexerClient) NFT(ctx context.Context, nftID iotago.NFTID) (*iotago.OutputID, *iotago.NFTOutput, iotago.SlotIndex, error) {
	outputID, output, ledgerIndex, err := client.singleOutputQuery(ctx, fmt.Sprintf(IndexerAPIRouteNFT, hexutil.EncodeHex(nftID[:])))
	if err != nil {
		return nil, nil, ledgerIndex, err
	}
	return outputID, output.(*iotago.NFTOutput), ledgerIndex, nil
}
