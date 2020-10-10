package iota

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var (
	// Returned for 400 bad request HTTP responses.
	ErrHTTPBadRequest = errors.New("bad request")
	// Returned for 500 internal server error HTTP responses.
	ErrHTTPInternalServerError = errors.New("internal server error")
	// Returned for 404 not found error HTTP responses.
	ErrHTTPNotFound = errors.New("not found")
	// Returned for 401 unauthorized error HTTP responses.
	ErrHTTPUnauthorized = errors.New("unauthorized")
	// Returned for unknown error HTTP responses.
	ErrHTTPUnknownError = errors.New("unknown error")
	// Returned for 501 not implemented error HTTP responses.
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
	// ParameterMessageID is used to identify a message by it's ID.
	ParameterMessageID = "messageID"

	// ParameterOutputID is used to identify an output by it's ID.
	ParameterOutputID = "outputID"

	// ParameterAddress is used to identify an address.
	ParameterAddress = "address"

	// ParameterMilestoneIndex is used to identify a milestone.
	ParameterMilestoneIndex = "milestoneIndex"
)

const (
	// NodeAPIRouteInfo is the route for getting the node info.
	// GET returns the node info.
	NodeAPIRouteInfo = "/info"

	// NodeAPIRouteTips is the route for getting two tips.
	// GET returns the tips.
	NodeAPIRouteTips = "/tips"

	// NodeAPIRouteMessageMetadata is the route for getting message metadata by it's messageID.
	// GET returns message metadata (including info about "promotion/reattachment needed").
	NodeAPIRouteMessageMetadata = "/messages/:" + ParameterMessageID + "/metadata"

	// NodeAPIRouteMessageBytes is the route for getting message raw data by it's messageID.
	// GET returns raw message data (bytes).
	NodeAPIRouteMessageBytes = "/messages/:" + ParameterMessageID + "/raw"

	// NodeAPIRouteMessageChildren is the route for getting message IDs of the children of a message, identified by it's messageID.
	// GET returns the message IDs of all children.
	NodeAPIRouteMessageChildren = "/messages/:" + ParameterMessageID + "/children"

	// NodeAPIRouteMessages is the route for getting message IDs or creating new messages.
	// GET with query parameter (mandatory) returns all message IDs that fit these filter criteria (query parameters: "index").
	// POST creates a single new message and returns the new message ID.
	NodeAPIRouteMessages = "/messages"

	// NodeAPIRouteMilestone is the route for getting a milestone by it's milestoneIndex.
	// GET returns the milestone.
	NodeAPIRouteMilestone = "/milestones/:" + ParameterMilestoneIndex

	// NodeAPIRouteOutput is the route for getting outputs by their outputID (transactionHash + outputIndex).
	// GET returns the output.
	NodeAPIRouteOutput = "/outputs/:" + ParameterOutputID

	// NodeAPIRouteAddressBalance is the route for getting the total balance of all unspent outputs of an address.
	// GET returns the balance of all unspent outputs of this address.
	NodeAPIRouteAddressBalance = "/addresses/:" + ParameterAddress

	// NodeAPIRouteAddressOutputs is the route for getting all output IDs for an address.
	// GET returns the outputIDs for all outputs of this address (optional query parameters: "include-spent").
	NodeAPIRouteAddressOutputs = "/addresses/:" + ParameterAddress + "/outputs"
)

// NewNodeAPI returns a new NodeAPI with the given BaseURL and HTTPClient.
func NewNodeAPI(baseURL string, httpClient ...http.Client) *NodeAPI {
	if len(httpClient) > 0 {
		return &NodeAPI{BaseURL: baseURL, HTTPClient: httpClient[0]}
	}
	return &NodeAPI{BaseURL: baseURL}
}

// NodeAPI is a client for node HTTP REST APIs.
type NodeAPI struct {
	// The HTTP client to use.
	HTTPClient http.Client
	// The base URL for all API calls.
	BaseURL string
}

// defines the error response schema for node API responses.
type HTTPErrorResponseEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// defines the ok response schema for node API responses.
type HTTPOkResponseEnvelope struct {
	// The encapsulated json data.
	Data interface{} `json:"data"`
}

// rawDataEnvelope is used internally to encapsulate binary data.
type rawDataEnvelope struct {
	// The encapsulated binary data.
	Data []byte
}

func interpretBody(res *http.Response, decodeTo interface{}) error {
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		if rawData, ok := decodeTo.(*rawDataEnvelope); ok {
			rawData.Data = make([]byte, len(resBody))
			copy(rawData.Data, resBody)
			return nil
		}

		okRes := &HTTPOkResponseEnvelope{Data: decodeTo}
		return json.Unmarshal(resBody, okRes)
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

func (api *NodeAPI) do(method string, route string, reqObj interface{}, resObj interface{}) (*http.Response, error) {

	// marshal request object
	var data []byte
	var raw bool

	if reqObj != nil {
		var err error

		if rawData, ok := reqObj.(*rawDataEnvelope); !ok {
			data, err = json.Marshal(reqObj)
			if err != nil {
				return nil, err
			}
		} else {
			data = rawData.Data
			raw = true
		}
	}

	// construct request
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", api.BaseURL, route), func() io.Reader {
		if data == nil {
			return nil
		}
		return bytes.NewReader(data)
	}())
	if err != nil {
		return nil, err
	}

	if data != nil {
		if !raw {
			req.Header.Set("Content-Type", contentTypeJSON)
		} else {
			req.Header.Set("Content-Type", contentTypeOctetStream)
		}
	}

	// make the request
	res, err := api.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resObj == nil {
		return res, nil
	}

	// write response into response object
	if err := interpretBody(res, resObj); err != nil {
		return nil, err
	}
	return res, nil
}

// NodeInfoResponse defines the response of a GET info REST API call.
type NodeInfoResponse struct {
	// The name of the node software.
	Name string `json:"name"`
	// The semver version of the node software.
	Version string `json:"version"`
	// Whether the node is healthy.
	IsHealthy bool `json:"isHealthy"`
	// The hex encoded public key of the coordinator.
	CoordinatorPublicKey string `json:"coordinatorPublicKey"`
	// The hex encoded message ID of the latest known milestone.
	LatestMilestoneMessageID string `json:"latestMilestoneMessageId"`
	// The latest known milestone index.
	LatestMilestoneIndex uint32 `json:"latestMilestoneIndex"`
	// The hex encoded message ID of the current solid milestone.
	SolidMilestoneMessageID string `json:"solidMilestoneMessageId"`
	// The current solid milestone's index.
	SolidMilestoneIndex uint32 `json:"solidMilestoneIndex"`
	// The milestone index at which the last pruning commenced.
	PruningIndex uint32 `json:"pruningIndex"`
	// The features this node exposes.
	Features []string `json:"features"`
}

// Info gets the info of the node.
func (api *NodeAPI) Info() (*NodeInfoResponse, error) {
	res := &NodeInfoResponse{}
	_, err := api.do(http.MethodGet, NodeAPIRouteInfo, nil, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// NodeTipsResponse defines the response of a GET tips REST API call.
type NodeTipsResponse struct {
	// The hex encoded message ID of the 1st tip.
	Tip1 string `json:"tip1MessageId"`
	// The hex encoded message ID of the 2nd tip.
	Tip2 string `json:"tip2MessageId"`
}

// Tips gets the two tips from the node.
func (api *NodeAPI) Tips() (*NodeTipsResponse, error) {
	res := &NodeTipsResponse{}
	_, err := api.do(http.MethodGet, NodeAPIRouteTips, nil, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// MessageMetadataResponse defines the response of a GET message metadata REST API call.
type MessageMetadataResponse struct {
	// The hex encoded message ID of the message.
	MessageID string `json:"messageId"`
	// The hex encoded message ID of the 1st parent the message references.
	Parent1 string `json:"parent1MessageId"`
	// The hex encoded message ID of the 2nd parent the message references.
	Parent2 string `json:"parent2MessageId"`
	// Whether the message is solid.
	Solid bool `json:"isSolid"`
	// The milestone index that references this message.
	ReferencedByMilestoneIndex *uint32 `json:"referencedByMilestoneIndex,omitempty"`
	// The ledger inclusion state of the transaction payload.
	LedgerInclusionState *string `json:"ledgerInclusionState,omitempty"`
	// Whether the message should be promoted.
	ShouldPromote *bool `json:"shouldPromote,omitempty"`
	// Whether the message should be reattached.
	ShouldReattach *bool `json:"shouldReattach,omitempty"`
}

// MessageByMessageID gets the metadata of a message by it's message ID from the node.
func (api *NodeAPI) MessageMetadataByMessageID(hash MessageID) (*MessageMetadataResponse, error) {

	query := strings.Replace(NodeAPIRouteMessageMetadata, ParameterMessageID, hex.EncodeToString(hash[:]), 1)

	res := &MessageMetadataResponse{}
	_, err := api.do(http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// MessageByMessageID get a message by it's message ID from the node.
func (api *NodeAPI) MessageByMessageID(hash MessageID) (*Message, error) {

	query := strings.Replace(NodeAPIRouteMessageBytes, ParameterMessageID, hex.EncodeToString(hash[:]), 1)

	res := &rawDataEnvelope{}
	_, err := api.do(http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	msg := &Message{}
	_, err = msg.Deserialize(res.Data, DeSeriModePerformValidation)
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

// MessageByMessageID get a message by it's message ID from the node.
func (api *NodeAPI) ChildrenByMessageID(hash MessageID) (*ChildrenResponse, error) {

	query := strings.Replace(NodeAPIRouteMessageChildren, ParameterMessageID, hex.EncodeToString(hash[:]), 1)

	res := &ChildrenResponse{}
	_, err := api.do(http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
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
func (api *NodeAPI) MessageIDsByIndex(index string) (*MessageIDsByIndexResponse, error) {
	var query strings.Builder
	query.WriteString(NodeAPIRouteMessages)
	query.WriteString("?index=")
	query.WriteString(index)

	res := &MessageIDsByIndexResponse{}
	_, err := api.do(http.MethodGet, query.String(), nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// SubmitMessage submits the given Message to the node.
// The node will take care of filling missing information.
// This function returns the finalized message created by the node.
func (api *NodeAPI) SubmitMessage(m *Message) (*Message, error) {

	data, err := m.Serialize(DeSeriModePerformValidation)
	if err != nil {
		return nil, err
	}

	req := &rawDataEnvelope{Data: data}

	res, err := api.do(http.MethodPost, NodeAPIRouteMessages, req, nil)
	if err != nil {
		return nil, err
	}

	messageID, err := MessageIDFromHexString(res.Header.Get(locationHeader))
	if err != nil {
		return nil, err
	}

	msg, err := api.MessageByMessageID(messageID)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// MilestoneResponse defines the response of a GET milestones REST API call.
type MilestoneResponse struct {
	// The index of the milestone.
	Index uint32 `json:"milestoneIndex"`
	// The hex encoded message ID of the message.
	MessageID string `json:"messageId"`
	// The unix time of the milestone payload.
	Time int64 `json:"timestamp"`
}

// MilestoneByIndex gets a milestone by its index.
func (api *NodeAPI) MilestoneByIndex(index uint32) (*MilestoneResponse, error) {

	query := strings.Replace(NodeAPIRouteMilestone, ParameterMilestoneIndex, strconv.FormatUint(uint64(index), 10), 1)

	res := &MilestoneResponse{}
	_, err := api.do(http.MethodGet, query, nil, res)
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

// Output deserializes the RawOutput to its output form.
func (nor *NodeOutputResponse) Output() (*SigLockedSingleOutput, error) {
	jsonSeri, err := DeserializeObjectFromJSON(nor.RawOutput, jsonoutputselector)
	if err != nil {
		return nil, err
	}
	seri, err := jsonSeri.ToSerializable()
	if err != nil {
		return nil, err
	}
	return seri.(*SigLockedSingleOutput), nil
}

// OutputByID gets an outputs by its ID from the node.
func (api *NodeAPI) OutputByID(utxoID UTXOInputID) (*NodeOutputResponse, error) {

	query := strings.Replace(NodeAPIRouteOutput, ParameterOutputID, utxoID.ToHex(), 1)

	res := &NodeOutputResponse{}
	_, err := api.do(http.MethodGet, query, nil, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// AddressBalanceResponse defines the response of a GET addresses REST API call.
type AddressBalanceResponse struct {
	// The hex encoded address.
	Address string `json:"address"`
	// The maximum count of results that are returned by the node.
	MaxResults uint32 `json:"maxResults"`
	// The actual count of results that are returned.
	Count uint32 `json:"count"`
	// The balance of the address.
	Balance uint64 `json:"balance"`
}

// BalanceByAddress returns the balance of an address.
func (api *NodeAPI) BalanceByAddress(address string) (*AddressBalanceResponse, error) {

	query := strings.Replace(NodeAPIRouteAddressBalance, ParameterAddress, address, 1)

	res := &AddressBalanceResponse{}
	_, err := api.do(http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// AddressOutputsResponse defines the response of a GET outputs by address REST API call.
type AddressOutputsResponse struct {
	// The hex encoded address.
	Address string `json:"address"`
	// The maximum count of results that are returned by the node.
	MaxResults uint32 `json:"maxResults"`
	// The actual count of results that are returned.
	Count uint32 `json:"count"`
	// The output IDs (transaction hash + output index) of the outputs on this address.
	OutputIDs []string `json:"outputIDs"`
}

// OutputIDsByAddress gets outputs IDs by addresses (unspent outputs) from the node.
func (api *NodeAPI) OutputIDsByAddress(address string) (*AddressOutputsResponse, error) {

	query := strings.Replace(NodeAPIRouteAddressOutputs, ParameterAddress, address, 1)

	res := &AddressOutputsResponse{}
	_, err := api.do(http.MethodGet, query, nil, res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
