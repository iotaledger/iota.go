package nodeclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	iotago "github.com/iotaledger/iota.go/v3"
)

// Indexer plugin routes.
var (
	IndexerAPIRouteBasicOutputs = "/api/" + IndexerPluginName + "/outputs/basic"
	IndexerAPIRouteAliases      = "/api/" + IndexerPluginName + "/outputs/alias"
	IndexerAPIRouteAlias        = "/api/" + IndexerPluginName + "/outputs/alias/%s"
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
		iotago.OutputAlias:   IndexerAPIRouteAliases,
		iotago.OutputFoundry: IndexerAPIRouteFoundries,
		iotago.OutputNFT:     IndexerAPIRouteNFTs,
	}
)

type (

	// IndexerClient is a client which queries the optional indexer functionality of a node.
	IndexerClient interface {
		// Outputs returns a handle to query for outputs.
		Outputs(ctx context.Context, query IndexerQuery) (*IndexerResultSet, error)
		// Alias queries for a specific iotago.AliasOutput by its identifier.
		Alias(ctx context.Context, aliasID iotago.AliasID) (*iotago.OutputID, *iotago.AliasOutput, error)
		// Foundry queries for a specific iotago.FoundryOutput by its identifier.
		Foundry(ctx context.Context, foundryID iotago.FoundryID) (*iotago.OutputID, *iotago.FoundryOutput, error)
		// NFT queries for a specific iotago.NFTOutput by its identifier.
		NFT(ctx context.Context, nftID iotago.NFTID) (*iotago.OutputID, *iotago.NFTOutput, error)
	}

	// IndexerQuery is a query executed against the indexer.
	IndexerQuery interface {
		// SetOffset sets the offset for the query.
		SetOffset(offset *string)
		// URLParas returns the query parameters as URL encoded query parameters.
		URLParas() (string, error)
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
		res, err := resultSet.client.OutputByID(resultSet.ctx, outputID)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch output %s: %w", outputID, err)
		}
		output, err := res.Output()
		if err != nil {
			return nil, fmt.Errorf("unable to build output %s: %w", outputID, err)
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

		urlParas, err := query.URLParas()
		if err != nil {
			return err
		}

		routeWithParas := fmt.Sprintf("%s?%s", baseRoute, urlParas)
		_, reqErr := client.Do(ctx, http.MethodGet, routeWithParas, nil, res.Response)
		return reqErr
	}
	res.nextFunc = nextFunc

	return res, nil
}

func (client *indexerClient) singleOutputQuery(ctx context.Context, route string) (*iotago.OutputID, iotago.Output, error) {
	res := &IndexerResponse{}
	if _, err := client.Do(ctx, http.MethodGet, route, nil, res); err != nil {
		return nil, nil, err
	}

	if len(res.Items) == 0 {
		return nil, nil, fmt.Errorf("%w for route %s", ErrIndexerNotFound, route)
	}

	outputID := res.Items.MustOutputIDs()[0]
	outputRes, err := client.core.OutputByID(ctx, outputID)
	if err != nil {
		return nil, nil, err
	}
	out, err := outputRes.Output()
	return &outputID, out, err
}

func (client *indexerClient) Alias(ctx context.Context, aliasID iotago.AliasID) (*iotago.OutputID, *iotago.AliasOutput, error) {
	outputID, output, err := client.singleOutputQuery(ctx, fmt.Sprintf(IndexerAPIRouteAlias, iotago.EncodeHex(aliasID[:])))
	if err != nil {
		return nil, nil, err
	}
	return outputID, output.(*iotago.AliasOutput), nil
}

func (client *indexerClient) Foundry(ctx context.Context, foundryID iotago.FoundryID) (*iotago.OutputID, *iotago.FoundryOutput, error) {
	outputID, output, err := client.singleOutputQuery(ctx, fmt.Sprintf(IndexerAPIRouteFoundry, iotago.EncodeHex(foundryID[:])))
	if err != nil {
		return nil, nil, err
	}
	return outputID, output.(*iotago.FoundryOutput), nil
}

func (client *indexerClient) NFT(ctx context.Context, nftID iotago.NFTID) (*iotago.OutputID, *iotago.NFTOutput, error) {
	outputID, output, err := client.singleOutputQuery(ctx, fmt.Sprintf(IndexerAPIRouteNFT, iotago.EncodeHex(nftID[:])))
	if err != nil {
		return nil, nil, err
	}
	return outputID, output.(*iotago.NFTOutput), nil
}
