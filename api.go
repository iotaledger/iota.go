package giota

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type API struct {
	client   *http.Client
	endpoint string
}

// NewAPI takes an (optional) endpoint and optional http.Client and returns
// an API struct. If an empty endpoint is supplied, then "http://localhost:14265"
// is used.
func NewAPI(endpoint string, c *http.Client) (*API, error) {
	if c == nil {
		c = http.DefaultClient
	}

	if endpoint == "" {
		endpoint = "http://localhost:14265/"
	}

	return &API{client: c, endpoint: endpoint}, nil
}

func (api *API) do(cmd interface{}, out interface{}) error {
	b, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	rd := bytes.NewReader(b)
	req, err := http.NewRequest("POST", api.endpoint, rd)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := api.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errResp := &ErrorResponse{}
		err = json.NewDecoder(resp.Body).Decode(errResp)
		if err != nil {
			return err
		}
		return errors.New(errResp.Exception)
	}

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if bytes.Contains(bs, []byte(`"exception"`)) {
		errResp := &ErrorResponse{}
		err = json.Unmarshal(bs, errResp)
		if err != nil {
			return err
		}
		return errors.New(errResp.Exception)
	}

	return json.Unmarshal(bs, out)
}

type ErrorResponse struct {
	Exception string `json:"exception"`
}

type GetNodeInfoRequest struct {
	Command string `json:"command"`
}

type GetNodeInfoResponse struct {
	AppName                            string `json:"appName"`
	AppVersion                         string `json:"appVersion"`
	Duration                           int64  `json:"duration"`
	JREAvailableProcessors             int64  `json:"jreAvailableProcessors"`
	JREFreeMemory                      int64  `json:"jreFreeMemory"`
	JREMaxMemory                       int64  `json:"jreMaxMemory"`
	JRETotalMemory                     int64  `json:"jreTotalMemory"`
	LatestMilestone                    string `json:"latestMilestone"`
	LatestMilestoneIndex               int64  `json:"latestMilestoneIndex"`
	LatestSolidSubtangleMilestone      string `json:"latestSolidSubtangleMilestone"`
	LatestSolidSubtangleMilestoneIndex int64  `json:"latestSolidSubtangleMilestoneIndex"`
	Neighbors                          int64  `json:"neighbors"`
	PacketQueueSize                    int64  `json:"packetQueueSize"`
	Time                               int64  `json:"time"`
	Tips                               int64  `json:"tips"`
	TransactionsToRequest              int64  `json:"transactionsToRequest"`
}

func (api *API) GetNodeInfo() (*GetNodeInfoResponse, error) {
	gni := &GetNodeInfoRequest{
		Command: "getNodeInfo",
	}

	resp := &GetNodeInfoResponse{}
	err := api.do(gni, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type GetNeighborsRequest struct {
	Command string `json:"command"`
}

type Neighbor struct {
	Address                     string `json:"address"`
	NumberOfAllTransactions     int64  `json:"numberOfAllTransactions"`
	NumberOfInvalidTransactions int64  `json:"numberOfInvalidTransactions"`
	NumberOfNewTransactions     int64  `json:"numberOfNewTransactions"`
}

type GetNeighborsResponse struct {
	Duration  int64
	Neighbors []Neighbor
}

func (api *API) GetNeighbors() (*GetNeighborsResponse, error) {
	gn := &GetNeighborsRequest{
		Command: "getNeighbors",
	}

	resp := &GetNeighborsResponse{}
	err := api.do(gn, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type AddNeighborsRequest struct {
	Command string `json:"command"`

	// URIS is an array of strings in the form of "udp://identifier:port"
	// where identifier can be either an IP address or a domain name.
	URIS []string `json:"uris"`
}

type AddNeighborsResponse struct {
	Duration       int64 `json"duration"`
	AddedNeighbors int64 `json"addedNeighbors"`
}

func (api *API) AddNeighbors(an *AddNeighborsRequest) (*AddNeighborsResponse, error) {
	an.Command = "addNeighbors"

	resp := &AddNeighborsResponse{}
	err := api.do(an, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type RemoveNeighborsRequest struct {
	Command string `json:"command"`

	// URIS is an array of strings in the form of "udp://identifier:port"
	// where identifier can be either an IP address or a domain name.
	URIS []string `json:"uris"`
}

type RemoveNeighborsResponse struct {
	Duration         int64 `json:"duration"`
	RemovedNeighbors int64 `json:"removedNeighbors"`
}

func (api *API) RemoveNeighbors(rn *RemoveNeighborsRequest) (*RemoveNeighborsResponse, error) {
	rn.Command = "removeNeighbors"

	resp := &RemoveNeighborsResponse{}
	err := api.do(rn, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type GetTipsRequest struct {
	Command string `json:"command"`
}

type GetTipsResponse struct {
	Duration int64    `json:"duration"`
	Hashes   []string `json:"hashes"`
}

func (api *API) GetTips() (*GetTipsResponse, error) {
	gt := &GetTipsRequest{
		Command: "getTips",
	}

	resp := &GetTipsResponse{}
	err := api.do(gt, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type FindTransactionsRequest struct {
	Command   string    `json:"command"`
	Bundles   *[]string `json:"bundles,omitempty"`
	Addresses *[]string `json:"addresses,omitempty"`
	Tags      *[]string `json:"tags,omitempty"`
	Approvees *[]string `json:"approvees,omitempty"`
}

type FindTransactionsResponse struct {
	Duration int64    `json:"duration"`
	Hashes   []string `json:"hashes"`
}

func (api *API) FindTransactions(ft *FindTransactionsRequest) (*FindTransactionsResponse, error) {
	ft.Command = "findTransactions"

	resp := &FindTransactionsResponse{}
	err := api.do(ft, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type GetTrytesRequest struct {
	Command string   `json:"command"`
	Hashes  []string `json:"hashes"`
}

type GetTrytesResponse struct {
	Duration int64    `json:"duration"`
	Trytes   []string `json:"trytes"`
}

func (api *API) GetTrytes(gt *GetTrytesRequest) (*GetTrytesResponse, error) {
	gt.Command = "getTrytes"

	resp := &GetTrytesResponse{}
	err := api.do(gt, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type GetInclusionStatesRequest struct {
	Command      string   `json:"command"`
	Transactions []string `json:"transactions"`
	Tips         []string `json:"tips"`
}

type GetInclusionStatesResponse struct {
	Duration int64  `json:"duration"`
	States   []bool `json:"states"`
}

func (api *API) GetInclusionStates(gis *GetInclusionStatesRequest) (*GetInclusionStatesResponse, error) {
	gis.Command = "getInclusionStates"

	resp := &GetInclusionStatesResponse{}
	err := api.do(gis, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type GetBalancesRequest struct {
	Command   string   `json:"command"`
	Addresses []string `json:"addresses"`
	Treshold  int64    `json:"treshold"`
}

type GetBalancesResponse struct {
	Duration       int64    `json:"duration"`
	Balances       []string `json:"balances"`
	Milestone      string   `json:"milestone"`
	MilestoneIndex int64    `json:"milestoneIndex"`
}

func (api *API) GetBalances(gb *GetBalancesRequest) (*GetBalancesResponse, error) {
	gb.Command = "getBalances"

	resp := &GetBalancesResponse{}
	err := api.do(gb, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type GetTransactionsToApproveRequest struct {
	Command string `json:"command"`
	Depth   int64  `json:"depth"`
}

type GetTransactionsToApproveResponse struct {
	Duration           int64  `json:"duration"`
	TrunkTransactions  string `json:"trunkTransactions"`
	BranchTransactions string `json:"branchTransactions"`
}

func (api *API) GetTransactionsToApprove(gtta *GetTransactionsToApproveRequest) (*GetTransactionsToApproveResponse, error) {
	gtta.Command = "getTransactionsToApprove"

	resp := &GetTransactionsToApproveResponse{}
	err := api.do(gtta, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type AttachToTangleRequest struct {
	Command            string   `json:"command"`
	TrunkTransactions  string   `json:"trunkTransactions"`
	BranchTransactions string   `json:"branchTransactions"`
	MinWeightMagnitude int64    `json:"minWeightMagnitude"`
	Trytes             []string `json:"trytes"`
}

type AttachToTangleResponse struct {
	Duration int64    `json:"duration"`
	Trytes   []string `json:"trytes"`
}

func (api *API) AttachToTangle(att *AttachToTangleRequest) (*AttachToTangleResponse, error) {
	att.Command = "attachToTangle"

	resp := &AttachToTangleResponse{}
	err := api.do(att, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type InterruptAttachingToTangleRequest struct {
	Command string `json:"command"`
}

type InterruptAttachingToTangleResponse struct{}

func (api *API) InterruptAttachingToTangle() (*InterruptAttachingToTangleResponse, error) {
	iatt := &InterruptAttachingToTangleRequest{
		Command: "interruptAttachingToTangle",
	}

	resp := &InterruptAttachingToTangleResponse{}
	err := api.do(iatt, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type BroadcastTransactionsRequest struct {
	Command string   `json:"command"`
	Trytes  []string `json:"trytes"`
}

type BroadcastTransactionsResponse struct{}

func (api *API) BroadcastTransactions(bt *BroadcastTransactionsRequest) (*BroadcastTransactionsResponse, error) {
	bt.Command = "broadcastTransactions"

	resp := &BroadcastTransactionsResponse{}
	err := api.do(bt, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type StoreTransactionsRequest struct {
	Command string   `json:"command"`
	Trytes  []string `json:"trytes"`
}

type StoreTransactionsResponse struct{}

func (api *API) StoreTransactions(st *StoreTransactionsRequest) (*StoreTransactionsResponse, error) {
	st.Command = "storeTransactions"

	resp := &StoreTransactionsResponse{}
	err := api.do(st, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
