/*
MIT License

Copyright (c) 2016 Sascha Hanse
Copyright (c) 2017 Shinya Yagyu

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package giota

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

//PublicNodes is a list of known public nodes from http://iotasupport.com/lightwallet.shtml.
var (
	PublicNodes = []string{
		"http://service.iotasupport.com:14265",
		"http://walletservice.iota.community:14265",
		"http://eugene.iota.community:14265",
		"http://185.101.92.190:14265",
		"http://185.101.94.8:14265",
		"http://iota-na.indenodes.net:14265",
		"http://iotaserver.forobits.com:14265",
		"http://eugene.iotasupport.com:14999",
		"http://eugeneoldisoft.iotasupport.com:14265",
	}
)

//RandomNode returns a random node from public nodes.
func RandomNode() string {
	b := make([]byte, 1)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return PublicNodes[int(b[0])%len(PublicNodes)]
}

//API is for calling APIs.
type API struct {
	client   *http.Client
	endpoint string
}

// NewAPI takes an (optional) endpoint and optional http.Client and returns
// an API struct. If an empty endpoint is supplied, then "http://localhost:14265"
// is used.
func NewAPI(endpoint string, c *http.Client) *API {
	if c == nil {
		c = http.DefaultClient
	}

	if endpoint == "" {
		endpoint = "http://localhost:14265/"
	}

	return &API{client: c, endpoint: endpoint}
}

func handleError(resp *ErrorResponse, err1, err2 error) error {
	if err1 != nil {
		return err1
	}
	if resp.Exception == "" {
		return err2
	}
	return errors.New(resp.Exception)
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
	defer func() {
		if err = resp.Body.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		errResp := &ErrorResponse{}
		err = json.NewDecoder(resp.Body).Decode(errResp)
		return handleError(errResp, err, fmt.Errorf("http status %d while calling API", resp.StatusCode))
	}

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if bytes.Contains(bs, []byte(`"exception"`)) {
		errResp := &ErrorResponse{}
		err = json.Unmarshal(bs, errResp)
		return handleError(errResp, err, fmt.Errorf("exception occured while calling API"))
	}
	return json.Unmarshal(bs, out)
}

//ErrorResponse is for occuring exception while calling API.
type ErrorResponse struct {
	Exception string `json:"exception"`
}

//GetNodeInfoResponse is for GetNode APi response.
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

//GetNodeInfo calls GetNodeInfo API.
func (api *API) GetNodeInfo() (*GetNodeInfoResponse, error) {
	resp := &GetNodeInfoResponse{}
	err := api.do(map[string]string{
		"command": "getNodeInfo",
	}, resp)
	return resp, err
}

//Neighbor is a part of response of GetNeighbors API.
type Neighbor struct {
	Address                     Address `json:"address"`
	NumberOfAllTransactions     int64   `json:"numberOfAllTransactions"`
	NumberOfInvalidTransactions int64   `json:"numberOfInvalidTransactions"`
	NumberOfNewTransactions     int64   `json:"numberOfNewTransactions"`
}

//GetNeighborsResponse is for GetNeighbors API resonse.
type GetNeighborsResponse struct {
	Duration  int64
	Neighbors []Neighbor
}

//GetNeighbors calls GetNeighbors API.
func (api *API) GetNeighbors() (*GetNeighborsResponse, error) {
	resp := &GetNeighborsResponse{}
	err := api.do(map[string]string{
		"command": "getNeighbors",
	}, resp)
	return resp, err
}

//AddNeighborsRequest is for AddNeighbors API request.
type AddNeighborsRequest struct {
	// URIS is an array of strings in the form of "udp://identifier:port"
	// where identifier can be either an IP address or a domain name.
	URIS []string `json:"uris"`
}

//AddNeighborsResponse is for AddNeighbors API resonse.
type AddNeighborsResponse struct {
	Duration       int64 `json:"duration"`
	AddedNeighbors int64 `json:"addedNeighbors"`
}

//AddNeighbors calls AddNeighbors API.
func (api *API) AddNeighbors(an *AddNeighborsRequest) (*AddNeighborsResponse, error) {
	resp := &AddNeighborsResponse{}
	err := api.do(&struct {
		Command string `json:"command"`
		*AddNeighborsRequest
	}{
		"addNeighbors",
		an,
	}, resp)
	return resp, err
}

//RemoveNeighborsRequest is for RemoveNeighbors API request.
type RemoveNeighborsRequest struct {
	// URIS is an array of strings in the form of "udp://identifier:port"
	// where identifier can be either an IP address or a domain name.
	URIS []string `json:"uris"`
}

//RemoveNeighborsResponse is for RemoveNeighbors API resonse.
type RemoveNeighborsResponse struct {
	Duration         int64 `json:"duration"`
	RemovedNeighbors int64 `json:"removedNeighbors"`
}

//RemoveNeighbors calls RemoveNeighbors API.
func (api *API) RemoveNeighbors(rn *RemoveNeighborsRequest) (*RemoveNeighborsResponse, error) {
	resp := &RemoveNeighborsResponse{}
	err := api.do(&struct {
		Command string `json:"command"`
		*RemoveNeighborsRequest
	}{
		"removeNeighbors",
		rn,
	}, resp)
	return resp, err
}

//GetTipsResponse is for GetTips API resonse.
type GetTipsResponse struct {
	Duration int64    `json:"duration"`
	Hashes   []Trytes `json:"hashes"`
}

//GetTips calls GetTips API.
func (api *API) GetTips() (*GetTipsResponse, error) {
	resp := &GetTipsResponse{}
	err := api.do(map[string]string{
		"command": "getTips",
	}, resp)
	return resp, err
}

//FindTransactionsRequest is for FindTransactions API request.
type FindTransactionsRequest struct {
	Bundles   []Trytes  `json:"bundles,omitempty"`
	Addresses []Address `json:"addresses,omitempty"`
	Tags      []Trytes  `json:"tags,omitempty"`
	Approvees []Trytes  `json:"approvees,omitempty"`
}

//FindTransactionsResponse is for FindTransaction API resonse.
type FindTransactionsResponse struct {
	Duration int64    `json:"duration"`
	Hashes   []Trytes `json:"hashes"`
}

//FindTransactions calls FindTransactions API.
func (api *API) FindTransactions(ft *FindTransactionsRequest) (*FindTransactionsResponse, error) {
	resp := &FindTransactionsResponse{}
	err := api.do(&struct {
		Command string `json:"command"`
		*FindTransactionsRequest
	}{
		"findTransactions",
		ft,
	}, resp)
	return resp, err
}

//GetTrytesRequest is for GetTrytes API request.
type GetTrytesRequest struct {
	Hashes []Trytes `json:"hashes"`
}

//GetTrytesResponse is for GetTrytes API resonse.
type GetTrytesResponse struct {
	Duration int64         `json:"duration"`
	Trytes   []Transaction `json:"trytes"`
}

//GetTrytes calls GetTrytes API.
func (api *API) GetTrytes(gt *GetTrytesRequest) (*GetTrytesResponse, error) {
	resp := &GetTrytesResponse{}
	err := api.do(&struct {
		Command string `json:"command"`
		*GetTrytesRequest
	}{
		"getTrytes",
		gt,
	}, resp)
	return resp, err
}

//GetInclusionStatesRequest is for GetInclusionStates API request.
type GetInclusionStatesRequest struct {
	Transactions []Trytes `json:"transactions"`
	Tips         []string `json:"tips"`
}

//GetInclusionStatesResponse is for GetInclusionStates API resonse.
type GetInclusionStatesResponse struct {
	Duration int64  `json:"duration"`
	States   []bool `json:"states"`
}

//GetInclusionStates calls GetInclusionStates API.
func (api *API) GetInclusionStates(gis *GetInclusionStatesRequest) (*GetInclusionStatesResponse, error) {
	resp := &GetInclusionStatesResponse{}
	err := api.do(&struct {
		Command string `json:"command"`
		*GetInclusionStatesRequest
	}{
		"getInclusionStates",
		gis,
	}, resp)
	return resp, err
}

//GetBalancesRequest is for GetBalances API request.
type GetBalancesRequest struct {
	Addresses []Address `json:"addresses"`
	Threshold int64     `json:"threshold"`
}

//GetBalancesResponse is for GetBalances API resonse.
type GetBalancesResponse struct {
	Duration       int64    `json:"duration"`
	Balances       []string `json:"balances"`
	Milestone      Trytes   `json:"milestone"`
	MilestoneIndex int64    `json:"milestoneIndex"`
}

//GetBalances calls GetBalances API.
func (api *API) GetBalances(gb *GetBalancesRequest) (*GetBalancesResponse, error) {
	if gb.Threshold <= 0 {
		gb.Threshold = 100
	}
	resp := &GetBalancesResponse{}
	err := api.do(&struct {
		Command string `json:"command"`
		*GetBalancesRequest
	}{
		"getBalances",
		gb,
	}, resp)
	return resp, err
}

//GetTransactionsToApproveRequest is for GetTransactionsToApprove API request.
type GetTransactionsToApproveRequest struct {
	Depth int64 `json:"depth"`
}

//GetTransactionsToApproveResponse is for GetTransactionsToApprove API resonse.
type GetTransactionsToApproveResponse struct {
	Duration          int64  `json:"duration"`
	TrunkTransaction  Trytes `json:"trunkTransaction"`
	BranchTransaction Trytes `json:"branchTransaction"`
}

//GetTransactionsToApprove calls GetTransactionsToApprove API.
func (api *API) GetTransactionsToApprove(gtta *GetTransactionsToApproveRequest) (*GetTransactionsToApproveResponse, error) {
	resp := &GetTransactionsToApproveResponse{}
	err := api.do(&struct {
		Command string `json:"command"`
		*GetTransactionsToApproveRequest
	}{
		"getTransactionsToApprove",
		gtta,
	}, resp)
	return resp, err
}

//AttachToTangleRequest is for AttachToTangle API request.
type AttachToTangleRequest struct {
	TrunkTransaction   Trytes   `json:"trunkTransaction"`
	BranchTransaction  Trytes   `json:"branchTransaction"`
	MinWeightMagnitude int64    `json:"minWeightMagnitude"`
	Trytes             []Trytes `json:"trytes"`
}

//AttachToTangleResponse is for AttachToTangle API resonse.
type AttachToTangleResponse struct {
	Duration int64    `json:"duration"`
	Trytes   []Trytes `json:"trytes"`
}

//AttachToTangle calls AttachToTangle API.
func (api *API) AttachToTangle(att *AttachToTangleRequest) (*AttachToTangleResponse, error) {
	resp := &AttachToTangleResponse{}
	err := api.do(&struct {
		Command string `json:"command"`
		*AttachToTangleRequest
	}{
		"attachToTangle",
		att,
	}, resp)
	return resp, err
}

//InterruptAttachingToTangle calls InterruptAttachingToTangle API.
func (api *API) InterruptAttachingToTangle() error {
	err := api.do(map[string]string{
		"command": "interruptAttachingToTangle",
	}, struct{}{})
	return err
}

//BroadcastTransactionsRequest is for BroadcastTransactions API request.
type BroadcastTransactionsRequest struct {
	Trytes []Trytes `json:"trytes"`
}

//BroadcastTransactions calls BroadcastTransactions API.
func (api *API) BroadcastTransactions(bt *BroadcastTransactionsRequest) error {
	err := api.do(&struct {
		Command string `json:"command"`
		*BroadcastTransactionsRequest
	}{
		"broadcastTransactions",
		bt,
	}, struct{}{})
	return err
}

//StoreTransactionsRequest is for StoreTransactions API request.
type StoreTransactionsRequest struct {
	Trytes []Trytes `json:"trytes"`
}

//StoreTransactions calls StoreTransactions API.
func (api *API) StoreTransactions(st *StoreTransactionsRequest) error {
	err := api.do(&struct {
		Command string `json:"command"`
		*StoreTransactionsRequest
	}{
		"storeTransactions",
		st,
	}, struct{}{})
	return err
}
