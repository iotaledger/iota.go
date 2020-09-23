package iota

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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
	contentTypeJSON = "application/json"
	locationHeader  = "Location"
)

const (
	// The route for the info HTTP API call.
	NodeAPIRouteInfo = "/info"
	// The route for the tips HTTP API call.
	NodeAPIRouteTips = "/tips"
	// The route for checking whether messages are referenced by a milestone HTTP API call.
	NodeAPIRouteMessagesReferencedByMilestone = "/messages/by-hash/is-referenced-by-milestone"
	// The route for retrieving messages by their hash HTTP API call.
	NodeAPIRouteMessagesByHash = "/messages/by-hash"
	// The route for submitting a new message HTTP API call.
	NodeAPIRouteMessageSubmit = "/messages"
	// The route for checking whether transactions are referenced by a milestone HTTP API call.
	NodeAPIRouteTransactionReferencedByMilestone = "/transaction-messages/is-confirmed"
	// The route for retrieving outputs by their identifier HTTP API call.
	NodeAPIRouteOutputsByID = "/outputs/by-id"
	// The route for retrieving outputs by their addresses HTTP API call.
	NodeAPIRouteOutputsByAddress = "/outputs/by-address"
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
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// defines the ok response schema for node API responses.
type HTTPOkResponseEnvelope struct {
	Data interface{} `json:"data"`
}

func interpretBody(res *http.Response, decodeTo interface{}) error {
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
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
	if reqObj != nil {
		var err error
		data, err = json.Marshal(reqObj)
		if err != nil {
			return nil, err
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
		req.Header.Set("Content-Type", contentTypeJSON)
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

// NodeInfoResponse defines the response of a node info HTTP API call.
type NodeInfoResponse struct {
	// The name of the node software.
	Name string `json:"name"`
	// The semver version of the node software.
	Version string `json:"version"`
	// Whether the node is healthy.
	IsHealthy bool `json:"isHealthy"`
	// The network in which the node operates in.
	OperatingNetwork string `json:"operatingNetwork"`
	// The amount of currently connected peers.
	Peers int `json:"peers"`
	// The used coordinator address.
	CoordinatorAddress string `json:"coordinatorAddress"`
	// Whether the node is synchronized.
	IsSynced bool `json:"isSynced"`
	// The latest known milestone hash.
	LatestMilestoneHash string `json:"latestMilestoneHash"`
	// The latest known milestone index.
	LatestMilestoneIndex uint64 `json:"latestMilestoneIndex"`
	// The current solid milestone's hash.
	LatestSolidMilestoneHash string `json:"latestSolidMilestoneHash"`
	// The current solid milestone's index.
	LatestSolidMilestoneIndex uint64 `json:"latestSolidMilestoneIndex"`
	// The milestone index at which the last pruning commenced.
	PruningIndex uint64 `json:"pruningIndex"`
	// The current time from the point of view of the node.
	Time uint64 `json:"time"`
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

// NodeTipsResponse defines the response of a node tips HTTP API call.
type NodeTipsResponse struct {
	// The hex encoded hash of the first tip message.
	Tip1 string `json:"tip1"`
	// The hex encoded hash of the second tip message.
	Tip2 string `json:"tip2"`
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

// MessagesByHash gets messages by their hashes from the node.
func (api *NodeAPI) MessagesByHash(hashes MessageHashes) ([]*Message, error) {
	var query strings.Builder
	query.WriteString(NodeAPIRouteMessagesByHash)
	query.WriteString("?hashes=")
	query.WriteString(strings.Join(HashesToHex(hashes), ","))

	var res []*Message
	_, err := api.do(http.MethodGet, query.String(), nil, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// SubmitMessage submits the given Message to the node.
// The node will take care of filling missing information.
// This function returns the finalized message created by the node.
func (api *NodeAPI) SubmitMessage(m *Message) (*Message, error) {
	res, err := api.do(http.MethodPost, NodeAPIRouteMessageSubmit, m, nil)
	if err != nil {
		return nil, err
	}

	msgHashes, err := MessagesHashFromHexString(res.Header.Get(locationHeader))
	if err != nil {
		return nil, err
	}

	msgs, err := api.MessagesByHash(msgHashes)
	if err != nil {
		return nil, err
	}
	return msgs[0], nil
}

// NodeObjectReferencedResponse defines the response for an object which is potentially
// referenced by a milestone node HTTP API call.
type NodeObjectReferencedResponse struct {
	// Tells whether the given object is referenced by a milestone.
	IsReferencedByMilestone bool `json:"isReferencedByMilestone"`
	// The index of the milestone which referenced the object.
	MilestoneIndex uint64 `json:"milestoneIndex"`
	// The timestamp of the milestone which referenced the object.
	MilestoneTimestamp uint64 `json:"milestoneTimestamp"`
}

// AreMessagesReferencedByMilestone tells whether the given messages are referenced by milestones.
// The response slice is ordered by the provided input hashes.
func (api *NodeAPI) AreMessagesReferencedByMilestone(hashes MessageHashes) ([]NodeObjectReferencedResponse, error) {
	var query strings.Builder
	query.WriteString(NodeAPIRouteMessagesReferencedByMilestone)
	query.WriteString("?hashes=")
	query.WriteString(strings.Join(HashesToHex(hashes), ","))

	var res []NodeObjectReferencedResponse
	_, err := api.do(http.MethodGet, query.String(), nil, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// AreTransactionsReferencedByMilestone tells whether the given transactions are referenced by milestones.
// The response slice is ordered by the provided input hashes.
func (api *NodeAPI) AreTransactionsReferencedByMilestone(hashes SignedTransactionPayloadHashes) ([]NodeObjectReferencedResponse, error) {
	var query strings.Builder
	query.WriteString(NodeAPIRouteTransactionReferencedByMilestone)
	query.WriteString("?hashes=")
	query.WriteString(strings.Join(HashesToHex(hashes), ","))

	var res []NodeObjectReferencedResponse
	_, err := api.do(http.MethodGet, query.String(), nil, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// NodeOutputResponse defines the construct of an RawOutput in a a node HTTP API call.
type NodeOutputResponse struct {
	// The output in its serialized form.
	RawOutput *json.RawMessage `json:"output"`
	// Whether this RawOutput is spent.
	Spent bool `json:"spent"`
}

// Output deserializes the RawOutput to its output form.
func (nor *NodeOutputResponse) Output() (*SigLockedSingleDeposit, error) {
	jsonSeri, err := DeserializeObjectFromJSON(nor.RawOutput, jsonoutputselector)
	if err != nil {
		return nil, err
	}
	seri, err := jsonSeri.ToSerializable()
	if err != nil {
		return nil, err
	}
	return seri.(*SigLockedSingleDeposit), nil
}

// OutputsByID gets outputs by their ID from the node.
func (api *NodeAPI) OutputsByID(utxosID UTXOInputIDs) ([]NodeOutputResponse, error) {
	var query strings.Builder
	query.WriteString(NodeAPIRouteOutputsByID)
	query.WriteString("?ids=")
	query.WriteString(strings.Join(utxosID.ToHex(), ","))

	var res []NodeOutputResponse
	_, err := api.do(http.MethodGet, query.String(), nil, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// OutputsByAddress gets outputs by their ID from the node.
func (api *NodeAPI) OutputsByAddress(addrs ...string) (map[string][]NodeOutputResponse, error) {
	var query strings.Builder
	query.WriteString(NodeAPIRouteOutputsByAddress)
	query.WriteString("?addresses=")
	query.WriteString(strings.Join(addrs, ","))

	res := map[string][]NodeOutputResponse{}
	_, err := api.do(http.MethodGet, query.String(), nil, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
