package iotago

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/iotaledger/hive.go/serializer"
)

var (
	// ErrHTTPBadRequest gets returned for 400 bad request HTTP responses.
	ErrHTTPBadRequest = errors.New("bad request")
	// ErrHTTPInternalServerError gets returned for 500 internal server error HTTP responses.
	ErrHTTPInternalServerError = errors.New("internal server error")
	// ErrHTTPNotFound gets returned for 404 not found error HTTP responses.
	ErrHTTPNotFound = errors.New("not found")
	// ErrHTTPUnauthorized gets returned for 401 unauthorized error HTTP responses.
	ErrHTTPUnauthorized = errors.New("unauthorized")
	// ErrHTTPUnknownError gets returned for unknown error HTTP responses.
	ErrHTTPUnknownError = errors.New("unknown error")
	// ErrHTTPNotImplemented gets returned for 501 not implemented error HTTP responses.
	ErrHTTPNotImplemented = errors.New("operation not implemented/supported/available")

	httpCodeToErr = map[int]error{
		http.StatusBadRequest:          ErrHTTPBadRequest,
		http.StatusInternalServerError: ErrHTTPInternalServerError,
		http.StatusNotFound:            ErrHTTPNotFound,
		http.StatusUnauthorized:        ErrHTTPUnauthorized,
		http.StatusNotImplemented:      ErrHTTPNotImplemented,
	}
)

const (
	contentTypeJSON        = "application/json"
	contentTypeOctetStream = "application/octet-stream"
	locationHeader         = "Location"
)

const (
	// NodeAPIRouteHealth is the route for querying a node's health status.
	NodeAPIRouteHealth = "/health"

	// NodeAPIRouteInfo is the route for getting the node info.
	// GET returns the node info.
	NodeAPIRouteInfo = "/api/v1/info"

	// NodeAPIRouteTips is the route for getting two tips.
	// GET returns the tips.
	NodeAPIRouteTips = "/api/v1/tips"

	// NodeAPIRouteMessageData is the route for getting message data by its messageID.
	// GET returns message data (json).
	NodeAPIRouteMessageData = "/api/v1/messages/%s"

	// NodeAPIRouteMessageMetadata is the route for getting message metadata by its messageID.
	// GET returns message metadata (including info about "promotion/reattachment needed").
	NodeAPIRouteMessageMetadata = "/api/v1/messages/%s/metadata"

	// NodeAPIRouteMessageBytes is the route for getting message raw data by its messageID.
	// GET returns raw message data (bytes).
	NodeAPIRouteMessageBytes = "/api/v1/messages/%s/raw"

	// NodeAPIRouteMessageChildren is the route for getting message IDs of the children of a message, identified by its messageID.
	// GET returns the message IDs of all children.
	NodeAPIRouteMessageChildren = "/api/v1/messages/%s/children"

	// NodeAPIRouteMessages is the route for getting message IDs or creating new messages.
	// GET with query parameter (mandatory) returns all message IDs that fit these filter criteria (query parameters: "index").
	// POST creates a single new message and returns the new message ID.
	NodeAPIRouteMessages = "/api/v1/messages"

	// NodeAPIRouteMilestone is the route for getting a milestone by its milestoneIndex.
	// GET returns the milestone.
	NodeAPIRouteMilestone = "/api/v1/milestones/%s"

	// NodeAPIRouteMilestoneUTXOChanges is the route for getting all UTXO changes of a milestone by its milestoneIndex.
	// GET returns the output IDs of all UTXO changes.
	NodeAPIRouteMilestoneUTXOChanges = "/api/v1/milestones/%s/utxo-changes"

	// NodeAPIRouteOutput is the route for getting outputs by their outputID (transactionHash + outputIndex).
	// GET returns the output.
	NodeAPIRouteOutput = "/api/v1/outputs/%s"

	// NodeAPIRouteAddressBech32Balance is the route for getting the total balance of all unspent outputs of a Bech32 address.
	// GET returns the balance of all unspent outputs of this address.
	NodeAPIRouteAddressBech32Balance = "/api/v1/addresses/%s"

	// NodeAPIRouteAddressEd25519Balance is the route for getting the total balance of all unspent outputs of an ed25519 address.
	// The ed25519 address must be encoded in hex.
	// GET returns the balance of all unspent outputs of this address.
	NodeAPIRouteAddressEd25519Balance = "/api/v1/addresses/ed25519/%s"

	// NodeAPIRouteAddressBech32Outputs is the route for getting all output IDs for a Bech32 address.
	// GET returns the outputIDs for all outputs of this address (optional query parameters: "include-spent").
	NodeAPIRouteAddressBech32Outputs = "/api/v1/addresses/%s/outputs"

	// NodeAPIRouteAddressEd25519Outputs is the route for getting all output IDs for an ed25519 address.
	// The ed25519 address must be encoded in hex.
	// GET returns the outputIDs for all outputs of this address (optional query parameters: "include-spent").
	NodeAPIRouteAddressEd25519Outputs = "/api/v1/addresses/ed25519/%s/outputs"

	// NodeAPIRouteTreasury is the route for getting the current treasury.
	// GET returns the treasury.
	NodeAPIRouteTreasury = "/api/v1/treasury"

	// NodeAPIRouteReceipts is the route for getting all persisted receipts on a node.
	// GET returns the receipts.
	NodeAPIRouteReceipts = "/api/v1/receipts"

	// NodeAPIRouteReceiptsByMigratedAtIndex is the route for getting all persisted receipts for a given migrated at index on a node.
	// GET returns the receipts for the given migrated at index.
	NodeAPIRouteReceiptsByMigratedAtIndex = "/api/v1/receipts/%s"

	// NodeAPIRoutePeer is the route for getting peers by their peerID.
	// GET returns the peer
	// DELETE deletes the peer.
	NodeAPIRoutePeer = "/api/v1/peers/%s"

	// NodeAPIRoutePeers is the route for getting all peers of the node.
	// GET returns a list of all peers.
	// POST adds a new peer.
	NodeAPIRoutePeers = "/api/v1/peers"
)

// RequestURLHook is a function to modify the URL before sending a request.
type RequestURLHook func(url string) string

// the default options applied to the NodeHTTPAPIClient.
var defaultNodeAPIOptions = []NodeHTTPAPIClientOption{
	WithNodeHTTPAPIClientHTTPClient(http.DefaultClient),
	WithNodeHTTPAPIClientUserInfo(nil),
	WithNodeHTTPAPIClientRequestURLHook(nil),
}

// NodeHTTPAPIClientOptions define options for the NodeHTTPAPIClient.
type NodeHTTPAPIClientOptions struct {
	// The HTTP client to use.
	httpClient *http.Client
	// The username and password information.
	userInfo *url.Userinfo
	// The hook to modify the URL before sending a request.
	requestURLHook RequestURLHook
}

// applies the given NodeHTTPAPIClientOption.
func (no *NodeHTTPAPIClientOptions) apply(opts ...NodeHTTPAPIClientOption) {
	for _, opt := range opts {
		opt(no)
	}
}

// WithNodeHTTPAPIClientHTTPClient sets the used HTTP Client.
func WithNodeHTTPAPIClientHTTPClient(httpClient *http.Client) NodeHTTPAPIClientOption {
	return func(opts *NodeHTTPAPIClientOptions) {
		opts.httpClient = httpClient
	}
}

// WithNodeHTTPAPIClientUserInfo sets the Userinfo used to add basic auth "Authorization" headers to the requests.
func WithNodeHTTPAPIClientUserInfo(userInfo *url.Userinfo) NodeHTTPAPIClientOption {
	return func(opts *NodeHTTPAPIClientOptions) {
		opts.userInfo = userInfo
	}
}

// WithNodeHTTPAPIClientRequestURLHook is used to modify the URL before sending a request.
func WithNodeHTTPAPIClientRequestURLHook(requestURLHook RequestURLHook) NodeHTTPAPIClientOption {
	return func(opts *NodeHTTPAPIClientOptions) {
		opts.requestURLHook = requestURLHook
	}
}

// NodeHTTPAPIClientOption is a function setting a NodeHTTPAPIClient option.
type NodeHTTPAPIClientOption func(opts *NodeHTTPAPIClientOptions)

// NewNodeHTTPAPIClient returns a new NodeHTTPAPIClient with the given BaseURL.
func NewNodeHTTPAPIClient(baseURL string, opts ...NodeHTTPAPIClientOption) *NodeHTTPAPIClient {

	options := &NodeHTTPAPIClientOptions{}
	options.apply(defaultNodeAPIOptions...)
	options.apply(opts...)

	return &NodeHTTPAPIClient{
		BaseURL: baseURL,
		opts:    options,
	}
}

// NodeHTTPAPIClient is a client for node HTTP REST API endpoints.
type NodeHTTPAPIClient struct {
	// The base URL for all API calls.
	BaseURL string
	// holds the NodeHTTPAPIClient options.
	opts *NodeHTTPAPIClientOptions
}

// HTTPErrorResponseEnvelope defines the error response schema for node API responses.
type HTTPErrorResponseEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// HTTPOkResponseEnvelope defines the ok response schema for node API responses.
type HTTPOkResponseEnvelope struct {
	// The encapsulated json data.
	Data interface{} `json:"data"`
}

// RawDataEnvelope is used internally to encapsulate binary data.
type RawDataEnvelope struct {
	// The encapsulated binary data.
	Data []byte
}

func readBody(res *http.Response) ([]byte, error) {
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}
	return resBody, nil
}

func interpretBody(res *http.Response, decodeTo interface{}) error {
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		if decodeTo == nil {
			return nil
		}

		resBody, err := readBody(res)
		if err != nil {
			return err
		}

		if rawData, ok := decodeTo.(*RawDataEnvelope); ok {
			rawData.Data = make([]byte, len(resBody))
			copy(rawData.Data, resBody)
			return nil
		}

		okRes := &HTTPOkResponseEnvelope{Data: decodeTo}
		return json.Unmarshal(resBody, okRes)
	}

	if res.StatusCode == http.StatusServiceUnavailable {
		return nil
	}

	resBody, err := readBody(res)
	if err != nil {
		return err
	}

	errRes := &HTTPErrorResponseEnvelope{}
	if err := json.Unmarshal(resBody, errRes); err != nil {
		return fmt.Errorf("unable to read error from response body: %w", err)
	}

	err, ok := httpCodeToErr[res.StatusCode]
	if !ok {
		err = ErrHTTPUnknownError
	}

	return fmt.Errorf("%w: url %s, error message: %s", err, res.Request.URL.String(), errRes.Error.Message)
}

func (api *NodeHTTPAPIClient) Do(ctx context.Context, method string, route string, reqObj interface{}, resObj interface{}) (*http.Response, error) {
	// marshal request object
	var data []byte
	var raw bool

	if reqObj != nil {
		var err error

		if rawData, ok := reqObj.(*RawDataEnvelope); !ok {
			data, err = json.Marshal(reqObj)
			if err != nil {
				return nil, fmt.Errorf("unable to serialize request object to JSON: %w", err)
			}
		} else {
			data = rawData.Data
			raw = true
		}
	}

	// construct request URL
	url := fmt.Sprintf("%s%s", api.BaseURL, route)
	if api.opts.requestURLHook != nil {
		url = api.opts.requestURLHook(url)
	}

	// construct request
	req, err := http.NewRequestWithContext(ctx, method, url, func() io.Reader {
		if data == nil {
			return nil
		}
		return bytes.NewReader(data)
	}())
	if err != nil {
		return nil, fmt.Errorf("unable to build http request: %w", err)
	}

	if api.opts.userInfo != nil {
		// set the userInfo for basic auth
		req.URL.User = api.opts.userInfo
	}

	if data != nil {
		if !raw {
			req.Header.Set("Content-Type", contentTypeJSON)
		} else {
			req.Header.Set("Content-Type", contentTypeOctetStream)
		}
	}

	// make the request
	res, err := api.opts.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// write response into response object
	if err := interpretBody(res, resObj); err != nil {
		return nil, err
	}
	return res, nil
}

// Health returns whether the given node is healthy.
func (api *NodeHTTPAPIClient) Health(ctx context.Context) (bool, error) {
	res, err := api.Do(ctx, http.MethodGet, NodeAPIRouteHealth, nil, nil)
	if err != nil {
		return false, err
	}
	if res.StatusCode == http.StatusServiceUnavailable {
		return false, nil
	}
	return true, nil
}

// NodeInfoResponse defines the response of a GET info REST API call.
type NodeInfoResponse struct {
	// The name of the node software.
	Name string `json:"name"`
	// The semver version of the node software.
	Version string `json:"version"`
	// Whether the node is healthy.
	IsHealthy bool `json:"isHealthy"`
	// The human friendly name of the network ID on which the node operates on.
	NetworkID string `json:"networkId"`
	// The HRP prefix used for Bech32 addresses in the node's network.
	Bech32HRP string `json:"bech32HRP"`
	// The minimum pow score of the network.
	MinPowScore float64 `json:"minPoWScore"`
	// The current rate of new messages per second.
	MessagesPerSecond float64 `json:"messagesPerSecond"`
	// The current rate of referenced messages per second.
	ReferencedMessagesPerSecond float64 `json:"referencedMessagesPerSecond"`
	// The ratio of referenced messages in relation to new messages of the last confirmed milestone.
	ReferencedRate float64 `json:"referencedRate"`
	// The timestamp of the latest known milestone.
	LatestMilestoneTimestamp int64 `json:"latestMilestoneTimestamp"`
	// The latest known milestone index.
	LatestMilestoneIndex uint32 `json:"latestMilestoneIndex"`
	// The current confirmed milestone's index.
	ConfirmedMilestoneIndex uint32 `json:"confirmedMilestoneIndex"`
	// The milestone index at which the last pruning commenced.
	PruningIndex uint32 `json:"pruningIndex"`
	// The features this node exposes.
	Features []string `json:"features"`
}

// Info gets the info of the node.
func (api *NodeHTTPAPIClient) Info(ctx context.Context) (*NodeInfoResponse, error) {
	res := &NodeInfoResponse{}
	_, err := api.Do(ctx, http.MethodGet, NodeAPIRouteInfo, nil, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// NodeTipsResponse defines the response of a GET tips REST API call.
type NodeTipsResponse struct {
	// The hex encoded message IDs of the tips.
	TipsHex []string `json:"tipMessageIds"`
}

// Tips returns the hex encoded tips as MessageIDs.
func (ntr *NodeTipsResponse) Tips() (MessageIDs, error) {
	msgIDs := make(MessageIDs, len(ntr.TipsHex))
	for i, tip := range ntr.TipsHex {
		msgID, err := hex.DecodeString(tip)
		if err != nil {
			return nil, err
		}
		copy(msgIDs[i][:], msgID)
	}
	return msgIDs, nil
}

// Tips gets the two tips from the node.
func (api *NodeHTTPAPIClient) Tips(ctx context.Context) (*NodeTipsResponse, error) {
	res := &NodeTipsResponse{}
	_, err := api.Do(ctx, http.MethodGet, NodeAPIRouteTips, nil, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// SubmitMessage submits the given Message to the node.
// The node will take care of filling missing information.
// This function returns the finalized message created by the node.
func (api *NodeHTTPAPIClient) SubmitMessage(ctx context.Context, m *Message) (*Message, error) {
	// Do not check the message because the validation would fail if
	// no parents were given. The node will first add this missing information and
	// validate the message afterwards.
	data, err := m.Serialize(serializer.DeSeriModeNoValidation)
	if err != nil {
		return nil, err
	}

	req := &RawDataEnvelope{Data: data}
	res, err := api.Do(ctx, http.MethodPost, NodeAPIRouteMessages, req, nil)
	if err != nil {
		return nil, err
	}

	messageID, err := MessageIDFromHexString(res.Header.Get(locationHeader))
	if err != nil {
		return nil, err
	}

	msg, err := api.MessageByMessageID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// MessageIDsByIndexResponse defines the response of a GET messages REST API call.
type MessageIDsByIndexResponse struct {
	// The index of the messages.
	Index string `json:"index"`
	// The maximum count of results that are returned by the node.
	MaxResults uint32 `json:"maxResults"`
	// The actual count of results that are returned.
	Count uint32 `json:"count"`
	// The hex encoded message IDs of the found messages with this index.
	MessageIDs []string `json:"messageIds"`
}

// MessageIDsByIndex gets message IDs filtered by index from the node.
func (api *NodeHTTPAPIClient) MessageIDsByIndex(ctx context.Context, index []byte) (*MessageIDsByIndexResponse, error) {
	var query strings.Builder
	query.WriteString(NodeAPIRouteMessages)
	query.WriteString("?index=")
	query.WriteString(hex.EncodeToString(index))

	res := &MessageIDsByIndexResponse{}
	_, err := api.Do(ctx, http.MethodGet, query.String(), nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// MessageMetadataResponse defines the response of a GET message metadata REST API call.
type MessageMetadataResponse struct {
	// The hex encoded message ID of the message.
	MessageID string `json:"messageId"`
	// The hex encoded message IDs of the parents the message references.
	Parents []string `json:"parentMessageIds"`
	// Whether the message is solid.
	Solid bool `json:"isSolid"`
	// The milestone index that references this message.
	ReferencedByMilestoneIndex *uint32 `json:"referencedByMilestoneIndex,omitempty"`
	// If this message represents a milestone this is the milestone index
	MilestoneIndex *uint32 `json:"milestoneIndex,omitempty"`
	// The ledger inclusion state of the transaction payload.
	LedgerInclusionState *string `json:"ledgerInclusionState,omitempty"`
	// Whether the message should be promoted.
	ShouldPromote *bool `json:"shouldPromote,omitempty"`
	// Whether the message should be reattached.
	ShouldReattach *bool `json:"shouldReattach,omitempty"`
	// The reason why this message is marked as conflicting.
	ConflictReason uint8 `json:"conflictReason,omitempty"`
}

// MessageMetadataByMessageID gets the metadata of a message by its message ID from the node.
func (api *NodeHTTPAPIClient) MessageMetadataByMessageID(ctx context.Context, msgID MessageID) (*MessageMetadataResponse, error) {
	query := fmt.Sprintf(NodeAPIRouteMessageMetadata, hex.EncodeToString(msgID[:]))

	res := &MessageMetadataResponse{}
	_, err := api.Do(ctx, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// MessageJSONByMessageID get a message by its message ID from the node (json).
func (api *NodeHTTPAPIClient) MessageJSONByMessageID(ctx context.Context, msgID MessageID) (*Message, error) {
	query := fmt.Sprintf(NodeAPIRouteMessageData, hex.EncodeToString(msgID[:]))

	res := &Message{}
	_, err := api.Do(ctx, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// MessageByMessageID get a message by its message ID from the node (bytes).
func (api *NodeHTTPAPIClient) MessageByMessageID(ctx context.Context, msgID MessageID) (*Message, error) {
	query := fmt.Sprintf(NodeAPIRouteMessageBytes, hex.EncodeToString(msgID[:]))

	res := &RawDataEnvelope{}
	_, err := api.Do(ctx, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	msg := &Message{}
	_, err = msg.Deserialize(res.Data, serializer.DeSeriModePerformValidation)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// ChildrenResponse defines the response of a GET children REST API call.
type ChildrenResponse struct {
	// The hex encoded message ID of the message.
	MessageID string `json:"messageId"`
	// The maximum count of results that are returned by the node.
	MaxResults uint32 `json:"maxResults"`
	// The actual count of results that are returned.
	Count uint32 `json:"count"`
	// The hex encoded message IDs of the children of this message.
	Children []string `json:"childrenMessageIds"`
}

// ChildrenByMessageID get a message by its message ID from the node.
func (api *NodeHTTPAPIClient) ChildrenByMessageID(ctx context.Context, msgID MessageID) (*ChildrenResponse, error) {
	query := fmt.Sprintf(NodeAPIRouteMessageChildren, hex.EncodeToString(msgID[:]))

	res := &ChildrenResponse{}
	_, err := api.Do(ctx, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// NodeOutputResponse defines the response of a GET outputs REST API call.
type NodeOutputResponse struct {
	// The hex encoded message ID of the message.
	MessageID string `json:"messageId"`
	// The hex encoded transaction id from which this output originated.
	TransactionID string `json:"transactionId"`
	// The index of the output.
	OutputIndex uint16 `json:"outputIndex"`
	// Whether this output is spent.
	Spent bool `json:"isSpent"`
	// The ledger index at which this output was available at.
	LedgerIndex uint64 `json:"ledgerIndex"`
	// The output in its serialized form.
	RawOutput *json.RawMessage `json:"output"`
}

// TxID returns the TransactionID.
func (nor *NodeOutputResponse) TxID() (*TransactionID, error) {
	txIDBytes, err := hex.DecodeString(nor.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("unable to decode raw transaction ID from JSON to transaction ID: %w", err)
	}
	var txID TransactionID
	copy(txID[:], txIDBytes)
	return &txID, nil
}

// Output deserializes the RawOutput to an Output.
func (nor *NodeOutputResponse) Output() (Output, error) {
	jsonSeri, err := DeserializeObjectFromJSON(nor.RawOutput, jsonOutputSelector)
	if err != nil {
		return nil, err
	}
	seri, err := jsonSeri.ToSerializable()
	if err != nil {
		return nil, err
	}
	output, isOutput := seri.(Output)
	if !isOutput {
		return nil, ErrUnknownOutputType
	}
	return output, nil
}

// OutputByID gets an outputs by its ID from the node.
func (api *NodeHTTPAPIClient) OutputByID(ctx context.Context, utxoID UTXOInputID) (*NodeOutputResponse, error) {
	query := fmt.Sprintf(NodeAPIRouteOutput, utxoID.ToHex())

	res := &NodeOutputResponse{}
	_, err := api.Do(ctx, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// AddressBalanceResponse defines the response of a GET addresses REST API call.
type AddressBalanceResponse struct {
	// The type of the address.
	AddressType byte `json:"addressType"`
	// The hex encoded address.
	Address string `json:"address"`
	// The balance of the address.
	Balance uint64 `json:"balance"`
	// Indicates if dust is allowed on this address.
	DustAllowed bool `json:"dustAllowed"`
	// The ledger index at which this balance was queried at.
	LedgerIndex uint64 `json:"ledgerIndex"`
}

// BalanceByBech32Address returns the balance of the given Bech32 address.
func (api *NodeHTTPAPIClient) BalanceByBech32Address(ctx context.Context, bech32Addr string) (*AddressBalanceResponse, error) {
	query := fmt.Sprintf(NodeAPIRouteAddressBech32Balance, bech32Addr)

	res := &AddressBalanceResponse{}
	_, err := api.Do(ctx, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// BalanceByEd25519Address returns the balance of an Ed25519 address.
func (api *NodeHTTPAPIClient) BalanceByEd25519Address(ctx context.Context, addr *Ed25519Address) (*AddressBalanceResponse, error) {
	query := fmt.Sprintf(NodeAPIRouteAddressEd25519Balance, addr.String())

	res := &AddressBalanceResponse{}
	_, err := api.Do(ctx, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// AddressOutputsResponse defines the response of a GET outputs by address REST API call.
type AddressOutputsResponse struct {
	// The type of the address.
	AddressType byte `json:"addressType"`
	// The hex encoded address.
	Address string `json:"address"`
	// The maximum count of results that are returned by the node.
	MaxResults uint32 `json:"maxResults"`
	// The actual count of results that are returned.
	Count uint32 `json:"count"`
	// The output IDs (transaction ID + output index) of the outputs on this address.
	OutputIDs []OutputIDHex `json:"outputIDs"`
	// The ledger index at which these outputs where available at.
	LedgerIndex uint64 `json:"ledgerIndex"`
}

// OutputIDsByBech32Address gets output IDs of outputs residing on the given Bech32 address.
// Per default only unspent outputs IDs are returned. Set includeSpentOutputs to true to also return spent output IDs.
func (api *NodeHTTPAPIClient) OutputIDsByBech32Address(ctx context.Context, bech32Addr string, includeSpentOutputs bool) (*AddressOutputsResponse, error) {
	query := fmt.Sprintf(NodeAPIRouteAddressBech32Outputs, bech32Addr)
	if includeSpentOutputs {
		query += "?include-spent=true"
	}

	res := &AddressOutputsResponse{}
	_, err := api.Do(ctx, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// OutputsByBech32Address gets the outputs residing on the given Bech32 address.
// Per default only unspent outputs are returned. Set includeSpentOutputs to true to also return spent outputs.
func (api *NodeHTTPAPIClient) OutputsByBech32Address(ctx context.Context, bech32Addr string, includeSpentOutputs bool) (*AddressOutputsResponse, map[*UTXOInput]Output, error) {
	res, err := api.OutputIDsByBech32Address(ctx, bech32Addr, includeSpentOutputs)
	if err != nil {
		return nil, nil, err
	}

	return api.outputIDsToOutputs(ctx, res)
}

// OutputIDsByEd25519Address gets output IDs of outputs residing on the given Ed25519Address.
// Per default only unspent output IDs are returned. Set includeSpentOutputs to true to also return spent output IDs.
func (api *NodeHTTPAPIClient) OutputIDsByEd25519Address(ctx context.Context, addr *Ed25519Address, includeSpentOutputs bool) (*AddressOutputsResponse, error) {
	query := fmt.Sprintf(NodeAPIRouteAddressEd25519Outputs, addr.String())
	if includeSpentOutputs {
		query += "?include-spent=true"
	}

	res := &AddressOutputsResponse{}
	_, err := api.Do(ctx, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// OutputsByEd25519Address gets the outputs residing on the given Ed25519Address.
// Per default only unspent outputs are returned. Set includeSpentOutputs to true to also return spent outputs.
func (api *NodeHTTPAPIClient) OutputsByEd25519Address(ctx context.Context, addr *Ed25519Address, includeSpentOutputs bool) (*AddressOutputsResponse, map[*UTXOInput]Output, error) {
	res, err := api.OutputIDsByEd25519Address(ctx, addr, includeSpentOutputs)
	if err != nil {
		return nil, nil, err
	}

	return api.outputIDsToOutputs(ctx, res)
}

// queries the actual outputs given an AddressOutputsResponse.
func (api *NodeHTTPAPIClient) outputIDsToOutputs(ctx context.Context, res *AddressOutputsResponse) (*AddressOutputsResponse, map[*UTXOInput]Output, error) {
	outputs := make(map[*UTXOInput]Output)
	for _, outputIDHex := range res.OutputIDs {
		utxoInput, err := outputIDHex.AsUTXOInput()
		if err != nil {
			return nil, nil, err
		}

		outputRes, err := api.OutputByID(ctx, utxoInput.ID())
		if err != nil {
			return nil, nil, err
		}

		output, err := outputRes.Output()
		if err != nil {
			return nil, nil, err
		}
		outputs[utxoInput] = output
	}

	return res, outputs, nil
}

// TreasuryResponse defines the response of a GET treasury REST API call.
type TreasuryResponse struct {
	MilestoneID string `json:"milestoneId"`
	Amount      uint64 `json:"amount"`
}

// Treasury gets the current treasury.
func (api *NodeHTTPAPIClient) Treasury(ctx context.Context) (*TreasuryResponse, error) {
	res := &TreasuryResponse{}
	_, err := api.Do(ctx, http.MethodGet, NodeAPIRouteTreasury, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ReceiptsResponse defines the response for receipts GET related REST API calls.
type ReceiptsResponse struct {
	Receipts []*ReceiptTuple `json:"receipts"`
}

// ReceiptTuple represents a receipt and the milestone index in which it was contained.
type ReceiptTuple struct {
	Receipt        *Receipt `json:"receipt"`
	MilestoneIndex uint32   `json:"milestoneIndex"`
}

// Receipts gets all receipts persisted on the node.
func (api *NodeHTTPAPIClient) Receipts(ctx context.Context) ([]*ReceiptTuple, error) {
	res := &ReceiptsResponse{}
	_, err := api.Do(ctx, http.MethodGet, NodeAPIRouteReceipts, nil, res)
	if err != nil {
		return nil, err
	}

	return res.Receipts, nil
}

// ReceiptsByMigratedAtIndex gets all receipts for the given migrated at index persisted on the node.
func (api *NodeHTTPAPIClient) ReceiptsByMigratedAtIndex(ctx context.Context, index uint32) ([]*ReceiptTuple, error) {
	query := fmt.Sprintf(NodeAPIRouteReceiptsByMigratedAtIndex, strconv.FormatUint(uint64(index), 10))

	res := &ReceiptsResponse{}
	_, err := api.Do(ctx, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res.Receipts, nil
}

// MilestoneResponse defines the response of a GET milestones REST API call.
type MilestoneResponse struct {
	// The index of the milestone.
	Index uint32 `json:"index"`
	// The hex encoded message ID of the message.
	MessageID string `json:"messageId"`
	// The unix time of the milestone payload.
	Time int64 `json:"timestamp"`
}

// MilestoneByIndex gets a milestone by its index.
func (api *NodeHTTPAPIClient) MilestoneByIndex(ctx context.Context, index uint32) (*MilestoneResponse, error) {
	query := fmt.Sprintf(NodeAPIRouteMilestone, strconv.FormatUint(uint64(index), 10))

	res := &MilestoneResponse{}
	_, err := api.Do(ctx, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// MilestoneUTXOChangesResponse defines the response of a GET milestone UTXO changes REST API call.
type MilestoneUTXOChangesResponse struct {
	// The index of the milestone.
	Index uint32 `json:"index"`
	// The output IDs (transaction hash + output index) of the newly created outputs.
	CreatedOutputs []string `json:"createdOutputs"`
	// The output IDs (transaction hash + output index) of the consumed (spent) outputs.
	ConsumedOutputs []string `json:"consumedOutputs"`
}

// MilestoneUTXOChangesByIndex returns all UTXO changes of a milestone by its milestoneIndex.
func (api *NodeHTTPAPIClient) MilestoneUTXOChangesByIndex(ctx context.Context, index uint32) (*MilestoneUTXOChangesResponse, error) {
	query := fmt.Sprintf(NodeAPIRouteMilestoneUTXOChanges, strconv.FormatUint(uint64(index), 10))

	res := &MilestoneUTXOChangesResponse{}
	_, err := api.Do(ctx, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// AddPeerRequest defines the request for a POST peer REST API call.
type AddPeerRequest struct {
	// The libp2p multi address of the peer.
	MultiAddress string `json:"multiAddress"`
	// The alias of to iditify the peer.
	Alias *string `json:"alias,omitempty"`
}

// PeerResponse defines the response of a GET peer REST API call.
type PeerResponse struct {
	// The libp2p identifier of the peer.
	ID string `json:"id"`
	// The libp2p multi addresses of the peer.
	MultiAddresses []string `json:"multiAddresses"`
	// The alias to identify the peer.
	Alias *string `json:"alias,omitempty"`
	// The relation (static, autopeered) of the peer.
	Relation string `json:"relation"`
	// Whether the peer is connected.
	Connected bool `json:"connected"`
	// The gossip related information about this peer.
	Gossip *GossipInfo `json:"gossip,omitempty"`
}

// GossipInfo represents information about an ongoing gossip protocol.
type GossipInfo struct {
	// The last received heartbeat by the given node.
	Heartbeat *GossipHeartbeat `json:"heartbeat"`
	// The metrics about sent and received protocol messages.
	Metrics PeerGossipMetrics `json:"metrics"`
}

// GossipHeartbeat represents a gossip heartbeat message.
// Peers send each other this gossip protocol message when their
// state is updated, such as when:
//	- a new milestone was received
//	- the solid milestone changed
//	- the node performed pruning of data
type GossipHeartbeat struct {
	// The solid milestone of the node.
	SolidMilestoneIndex uint32 `json:"solidMilestoneIndex"`
	// The milestone index at which the node pruned its data.
	PrunedMilestoneIndex uint32 `json:"prunedMilestoneIndex"`
	// The latest known milestone index by the node.
	LatestMilestoneIndex uint32 `json:"latestMilestoneIndex"`
	// The amount of currently connected neighbors.
	ConnectedNeighbors int `json:"connectedNeighbors"`
	// The amount of currently connected neighbors who also
	// are synchronized with the network.
	SyncedNeighbors int `json:"syncedNeighbors"`
}

// PeerGossipMetrics defines the peer gossip metrics.
type PeerGossipMetrics struct {
	// The total amount of received new messages.
	NewMessages uint32 `json:"newMessages"`
	// The total amount of received known messages.
	KnownMessages uint32 `json:"knownMessages"`
	// The total amount of received messages.
	ReceivedMessages uint32 `json:"receivedMessages"`
	// The total amount of received message requests.
	ReceivedMessageRequests uint32 `json:"receivedMessageRequests"`
	// The total amount of received milestone requests.
	ReceivedMilestoneRequests uint32 `json:"receivedMilestoneRequests"`
	// The total amount of received heartbeats.
	ReceivedHeartbeats uint32 `json:"receivedHeartbeats"`
	// The total amount of sent messages.
	SentMessages uint32 `json:"sentMessages"`
	// The total amount of sent message request.
	SentMessageRequests uint32 `json:"sentMessageRequests"`
	// The total amount of sent milestone request.
	SentMilestoneRequests uint32 `json:"sentMilestoneRequests"`
	// The total amount of sent heartbeats.
	SentHeartbeats uint32 `json:"sentHeartbeats"`
	// The total amount of packets which couldn't be sent.
	DroppedPackets uint32 `json:"droppedPackets"`
}

// PeerByID gets a peer by its identifier.
func (api *NodeHTTPAPIClient) PeerByID(ctx context.Context, id string) (*PeerResponse, error) {
	query := fmt.Sprintf(NodeAPIRoutePeer, id)

	res := &PeerResponse{}
	_, err := api.Do(ctx, http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// RemovePeerByID removes a peer by its identifier.
func (api *NodeHTTPAPIClient) RemovePeerByID(ctx context.Context, id string) error {
	query := fmt.Sprintf(NodeAPIRoutePeer, id)

	_, err := api.Do(ctx, http.MethodDelete, query, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// Peers returns a list of all peers.
func (api *NodeHTTPAPIClient) Peers(ctx context.Context) ([]*PeerResponse, error) {
	res := []*PeerResponse{}
	_, err := api.Do(ctx, http.MethodGet, NodeAPIRoutePeers, nil, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// AddPeer adds a new peer by libp2p multi address with optional alias.
func (api *NodeHTTPAPIClient) AddPeer(ctx context.Context, multiAddress string, alias ...string) (*PeerResponse, error) {
	req := &AddPeerRequest{
		MultiAddress: multiAddress,
	}

	if len(alias) > 0 {
		req.Alias = &alias[0]
	}

	res := &PeerResponse{}
	_, err := api.Do(ctx, http.MethodPost, NodeAPIRoutePeers, req, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
