package giota

import (
	"bytes"
	"encoding/json"
	"github.com/iotaledger/giota/bundle"
	"github.com/iotaledger/giota/curl"
	"github.com/iotaledger/giota/pow"
	"github.com/iotaledger/giota/signing"
	"github.com/iotaledger/giota/transaction"
	"github.com/iotaledger/giota/trinary"
	"github.com/pkg/errors"
	"sort"

	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	ErrInvalidTailTransactionHash          = errors.New("the given transaction hash is not a trail transaction hash (current index must be 0)")
	ErrTransactionNotFound                 = errors.New("couldn't find transaction via getTrytes")
	ErrTransactionNotFoundInInclusionState = errors.New("couldn't find transactions via getInclusionStates")
	ErrNotEnoughBalance                    = errors.New("not enough balance")
	ErrInvalidAddressStartEnd              = errors.New("start/end invalid, start must be less than end and end must be less than start+500")
	ErrEmptyTransferForPromote             = errors.New("given bundle for promotion is empty")
	ErrInconsistentSubtangle               = errors.New("inconsistent subtangle")
	ErrNonOKStatusCodeFromAPIRequest       = errors.New("endpoint returned a non OK status code")
	ErrUnknownErrorFromAPIRequest          = errors.New("an unknown error occurred while calling API")
	ErrInputsNotProvided                   = errors.New("inputs must me provided for controling transfer balances")
	ErrInputsInsufficientForBundle         = errors.New("the total balance of all inputs isn't enough for the bundle transfer")
)

// API provides a basic interface to the API calls of IRI.
type API struct {
	client   *http.Client
	endpoint string
}

const localhostEndpoint = "http://localhost:14265/"

// NewAPI constructs a new API client connecting to the given node endpoint.
// If the endpoint is empty the client will connect to localhost.
func NewAPI(endpoint string, c *http.Client) *API {
	if c == nil {
		c = http.DefaultClient
	}

	if endpoint == "" {
		endpoint = localhostEndpoint
	}

	return &API{client: c, endpoint: endpoint}
}

func handleError(err *APIErrorResponse, err1, err2 error) error {
	switch {
	case err.Error != "":
		return errors.New(err.Error)
	case err.Exception != "":
		return errors.New(err.Exception)
	case err1 != nil:
		return err1
	}

	return err2
}

// does the actual API call to the node and given an out interface, initializes it with the response.
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
	req.Header.Set("X-IOTA-API-Version", "1")
	resp, err := api.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		errResp := &APIErrorResponse{}
		err = json.Unmarshal(bs, errResp)
		return handleError(errResp, err, errors.Wrapf(ErrNonOKStatusCodeFromAPIRequest, "http code %d", resp.StatusCode))
	}

	if bytes.Contains(bs, []byte(`"error"`)) || bytes.Contains(bs, []byte(`"exception"`)) {
		errResp := &APIErrorResponse{}
		err = json.Unmarshal(bs, errResp)
		return handleError(errResp, err, ErrUnknownErrorFromAPIRequest)
	}

	if out == nil {
		return nil
	}
	return json.Unmarshal(bs, out)
}

// APIErrorResponse represents the exception response for API calls.
type APIErrorResponse struct {
	Error     string `json:"error"`
	Exception string `json:"exception"`
}

type GetNodeInfoRequest struct {
	Command string `json:"command"`
}

type GetNodeInfoResponse struct {
	AppName                            string         `json:"appName"`
	AppVersion                         string         `json:"appVersion"`
	Duration                           int64          `json:"duration"`
	JREVersion                         string         `json:"jreVersion"`
	JREAvailableProcessors             int64          `json:"jreAvailableProcessors"`
	JREFreeMemory                      int64          `json:"jreFreeMemory"`
	JREMaxMemory                       int64          `json:"jreMaxMemory"`
	JRETotalMemory                     int64          `json:"jreTotalMemory"`
	LatestMilestone                    trinary.Trytes `json:"latestMilestone"`
	LatestMilestoneIndex               int64          `json:"latestMilestoneIndex"`
	LatestSolidSubtangleMilestone      trinary.Trytes `json:"latestSolidSubtangleMilestone"`
	LatestSolidSubtangleMilestoneIndex int64          `json:"latestSolidSubtangleMilestoneIndex"`
	Neighbors                          int64          `json:"neighbors"`
	PacketQueueSize                    int64          `json:"packetQueueSize"`
	Time                               int64          `json:"time"`
	Tips                               int64          `json:"tips"`
	TransactionsToRequest              int64          `json:"transactionsToRequest"`
}

// GetNodeInfo returns information about the connected node.
func (api *API) GetNodeInfo() (*GetNodeInfoResponse, error) {
	resp := &GetNodeInfoResponse{}
	err := api.do(map[string]string{
		"command": "getNodeInfo",
	}, resp)

	return resp, err
}

// CheckConsistency checks if a transaction is consistent or a set of transactions are co-consistent, by calling checkConsistency command.
// Co-consistent transactions and the transactions that they approve (directly or indirectly), are not conflicting with each other and rest of the ledger.
// As long as a transaction is consistent it might be accepted by the network.
// In case transaction is inconsistent, it will not be accepted, and a reattachment is required by calling ReplayBundle().
func (api *API) CheckConsistency(tailTransactionHashes ...trinary.Trytes) (*CheckConsistencyResponse, error) {
	resp := &CheckConsistencyResponse{}
	err := api.do(&struct {
		Command string           `json:"command"`
		Tails   []trinary.Trytes `json:"tails"`
	}{
		"checkConsistency",
		tailTransactionHashes,
	}, resp)

	return resp, err
}

type CheckConsistencyResponse struct {
	Duration int64  `json:"duration"`
	State    bool   `json:"state"`
	Info     string `json:"info"`
}

// Neighbor represents a peer of the connected node.
type Neighbor struct {
	Address                           signing.Address `json:"address"`
	ConnectionType                    string          `json:"connectionType"`
	NumberOfAllTransactions           int64           `json:"numberOfAllTransactions"`
	NumberOfInvalidTransactions       int64           `json:"numberOfInvalidTransactions"`
	NumberOfNewTransactions           int64           `json:"numberOfNewTransactions"`
	NumberOfRandomTransactionRequests int64           `json:"numberOfRandomTransactionRequests"`
	NumberOfSentTransactions          int64           `json:"numberOfSentTransactions"`
}

type GetNeighborsRequest struct {
	Command string `json:"command"`
}

type GetNeighborsResponse struct {
	Duration  int64
	Neighbors []Neighbor
}

// GetNeighbors returns list of connected neighbors.
func (api *API) GetNeighbors() (*GetNeighborsResponse, error) {
	resp := &GetNeighborsResponse{}
	err := api.do(map[string]string{
		"command": "getNeighbors",
	}, resp)

	return resp, err
}

type AddNeighborsRequest struct {
	Command string `json:"command"`

	// URIs is an array of strings in the form of "udp://identifier:port"
	// where identifier can be either an IP address or a domain name.
	URIs []string `json:"uris"`
}

type AddNeighborsResponse struct {
	Duration       int64 `json:"duration"`
	AddedNeighbors int64 `json:"addedNeighbors"`
}

// AddNeighbors adds the given neighbors to the connected node.
func (api *API) AddNeighbors(uris []string) (*AddNeighborsResponse, error) {
	resp := &AddNeighborsResponse{}
	err := api.do(&struct {
		Command string   `json:"command"`
		URIS    []string `json:"uris"`
	}{
		"addNeighbors",
		uris,
	}, resp)

	return resp, err
}

type RemoveNeighborsRequest struct {
	Command string `json:"command"`

	// URIs is an array of strings in the form of "udp://identifier:port"
	// where identifier can be either an IP address or a domain name.
	URIs []string `json:"uris"`
}

type RemoveNeighborsResponse struct {
	Duration         int64 `json:"duration"`
	RemovedNeighbors int64 `json:"removedNeighbors"`
}

// RemoveNeighbors removes the given neighbors from the connected node.
func (api *API) RemoveNeighbors(uris []string) (*RemoveNeighborsResponse, error) {
	resp := &RemoveNeighborsResponse{}
	err := api.do(&struct {
		Command string   `json:"command"`
		URIS    []string `json:"uris"`
	}{
		"removeNeighbors",
		uris,
	}, resp)

	return resp, err
}

type GetTipsRequest struct {
	Command string `json:"command"`
}

type GetTipsResponse struct {
	Duration int64            `json:"duration"`
	Hashes   []trinary.Trytes `json:"hashes"`
}

// GetTips calls returns a list of tips (transactions not referenced by other transactions), as seen by the connected node.
func (api *API) GetTips() (*GetTipsResponse, error) {
	resp := &GetTipsResponse{}
	err := api.do(map[string]string{
		"command": "getTips",
	}, resp)

	return resp, err
}

type FindTransactionsRequest struct {
	Command   string            `json:"command"`
	Bundles   []trinary.Trytes  `json:"bundles,omitempty"`
	Addresses []signing.Address `json:"addresses,omitempty"`
	Tags      []trinary.Trytes  `json:"tags,omitempty"`
	Approvees []trinary.Trytes  `json:"approvees,omitempty"`
}

type FindTransactionsResponse struct {
	Duration int64            `json:"duration"`
	Hashes   []trinary.Trytes `json:"hashes"`
}

// FindTransactions searches for transaction hashes by calling findTransactions command.
// It allows to search for transactions by passing a query object with addresses, tags and approvees fields.
// Multiple query fields are supported and FindTransactions() returns intersection of results.
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

type GetTrytesResponse struct {
	Duration int64            `json:"duration"`
	Trytes   []trinary.Trytes `json:"trytes"`
}

// GetTrytes fetches the transaction trytes given a list of transaction hashes.
func (api *API) GetTrytes(hashes ...trinary.Trytes) (*GetTrytesResponse, error) {
	resp := &GetTrytesResponse{}
	err := api.do(&struct {
		Command string           `json:"command"`
		Hashes  []trinary.Trytes `json:"hashes"`
	}{
		"getTrytes",
		hashes,
	}, resp)

	return resp, err
}

type GetInclusionStatesRequest struct {
	Command      string           `json:"command"`
	Transactions []trinary.Trytes `json:"transactions"`
	Tips         []trinary.Trytes `json:"tips"`
}

type GetInclusionStatesResponse struct {
	Duration int64  `json:"duration"`
	States   []bool `json:"states"`
}

// GetInclusionStates fetches inclusion states of given list of transactions.
func (api *API) GetInclusionStates(tx []trinary.Trytes, tips []trinary.Trytes) (*GetInclusionStatesResponse, error) {
	resp := &GetInclusionStatesResponse{}
	err := api.do(&struct {
		Command      string           `json:"command"`
		Transactions []trinary.Trytes `json:"transactions"`
		Tips         []trinary.Trytes `json:"tips"`
	}{
		"getInclusionStates",
		tx,
		tips,
	}, resp)

	return resp, err
}

type WereAddressesSpentFromResponse struct {
	States   []bool `json:"states"`
	Duration int64  `json:"duration"`
}

// WereAddressesSpentFrom takes a list of addresses and checks whether they were spent from.
// The resulting bool slice contains the answer in the same order as the given addresses.
func (api *API) WereAddressesSpentFrom(addr ...signing.Address) ([]bool, error) {
	resp := &WereAddressesSpentFromResponse{}
	err := api.do(&struct {
		Command   string            `json:"command"`
		Addresses signing.Addresses `json:"addresses"`
	}{
		"wereAddressesSpentFrom",
		addr,
	}, resp)

	return resp.States, err
}

// Balance represents the current balance of an address.
type Balance struct {
	Address  signing.Address
	Value    int64
	Security signing.SecurityLevel
}

type Balances []Balance

// Total returns the total balance.
func (bs Balances) Total() int64 {
	var total int64
	for _, b := range bs {
		total += b.Value
	}
	return total
}

type GetBalancesRequest struct {
	Command   string            `json:"command"`
	Addresses []signing.Address `json:"addresses"`
	Threshold int64             `json:"threshold"`
}

type GetBalancesResponse struct {
	Duration       int64    `json:"duration"`
	Balances       []int64  `json:"balances"`
	References     []string `json:"references"`
	MilestoneIndex int64    `json:"milestoneIndex"`
}

// Balances fetches and returns balances of the given addresses.
func (api *API) Balances(addrs signing.Addresses) (Balances, error) {
	r, err := api.GetBalances(addrs, 100)
	if err != nil {
		return nil, err
	}

	bs := make(Balances, 0, len(addrs))
	for i, bal := range r.Balances {
		b := Balance{
			Address: addrs[i],
			Value:   bal,
		}
		bs = append(bs, b)
	}
	return bs, nil
}

// GetBalances fetches confirmed balances of the given addresses at the latest solid milestone.
func (api *API) GetBalances(adr []signing.Address, threshold int64) (*GetBalancesResponse, error) {
	if threshold <= 0 {
		threshold = 100
	}

	type getBalancesResponse struct {
		Duration       int64    `json:"duration"`
		Balances       []string `json:"balances"`
		References     []string `json:"references"`
		MilestoneIndex int64    `json:"milestoneIndex"`
	}

	resp := &getBalancesResponse{}
	err := api.do(&struct {
		Command   string            `json:"command"`
		Addresses []signing.Address `json:"addresses"`
		Threshold int64             `json:"threshold"`
	}{
		"getBalances",
		adr,
		threshold,
	}, resp)

	r := &GetBalancesResponse{
		Duration:       resp.Duration,
		Balances:       make([]int64, len(resp.Balances)),
		References:     resp.References,
		MilestoneIndex: resp.MilestoneIndex,
	}

	for i, ba := range resp.Balances {
		r.Balances[i], err = strconv.ParseInt(ba, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	return r, err
}

type GetTransactionsToApproveRequest struct {
	Command string `json:"command"`
	Depth   int64  `json:"depth"`
}

type GetTransactionsToApproveResponse struct {
	Duration          int64          `json:"duration"`
	TrunkTransaction  trinary.Trytes `json:"trunkTransaction"`
	BranchTransaction trinary.Trytes `json:"branchTransaction"`
}

// GetTransactionsToApprove does the tip selection by calling the getTransactionsToApprove command.
// Returns a pair of approved transactions, which are chosen randomly after validating the transaction trytes,
// the signatures and cross-checking for conflicting transactions.
//
// Tip selection is executed by a Random Walk (RW) starting at random point in given depth ending up to the pair
// of selected tips. For more information about tip selection please refer to the whitepaper.
//
// The reference option allows to select tips in a way that the reference transaction is being approved too.
// This is useful for promoting transactions, for example with promoteTransaction.
func (api *API) GetTransactionsToApprove(depth int, reference trinary.Trytes) (*GetTransactionsToApproveResponse, error) {
	resp := &GetTransactionsToApproveResponse{}
	err := api.do(&struct {
		Command   string         `json:"command"`
		Depth     int            `json:"depth"`
		Reference trinary.Trytes `json:"reference,omitempty"`
	}{
		"getTransactionsToApprove",
		depth,
		reference,
	}, resp)

	return resp, err
}

type AttachToTangleRequest struct {
	Command            string                    `json:"command"`
	TrunkTransaction   trinary.Trytes            `json:"trunkTransaction"`
	BranchTransaction  trinary.Trytes            `json:"branchTransaction"`
	MinWeightMagnitude int64                     `json:"minWeightMagnitude"`
	Trytes             []transaction.Transaction `json:"trytes"`
}

type AttachToTangleResponse struct {
	Duration int64                     `json:"duration"`
	Trytes   []transaction.Transaction `json:"trytes"`
}

// AttachToTangle performs the Proof-of-Work required to attach a transaction to the Tangle.
// Returns list of transaction trytes and overwrites the following fields:
// hash, nonce, attachmentTimestamp, attachmentTimestampLowerBound, attachmentTimestampUpperBound.
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

type InterruptAttachingToTangleRequest struct {
	Command string `json:"command"`
}

// InterruptAttachingToTangle interrupts the current ongoing Proof-Of-Work done by the connected node.
func (api *API) InterruptAttachingToTangle() error {
	err := api.do(map[string]string{
		"command": "interruptAttachingToTangle",
	}, nil)

	return err
}

// AccountData represents an account's current state determined by FindTransactions(), WereAddressesSpentFrom()
// and GetBalances()
type AccountData struct {
	LatestAddress signing.Address   `json:"latest_address"`
	Transfers     bundle.Bundles    `json:"transfers"`
	Transactions  []trinary.Trytes  `json:"transactions"`
	Inputs        Balances          `json:"inputs"`
	Addresses     signing.Addresses `json:"addresses"`
	Balance       int64             `json:"balance"`
}

// GetBundlesFromAddresses returns a by attachment timestamp ordered list of bundles of the given addresses.
func (api *API) GetBundlesFromAddresses(addrs signing.Addresses) (bundle.Bundles, error) {
	// fetch transactions which operated on the given addresses
	txs, err := api.FindTransactionObjects(&FindTransactionsRequest{Addresses: addrs})
	if err != nil {
		return nil, err
	}

	// fetch all transactions associated with the bundle of every transaction
	// use a map as a ghetto set
	bundleHashesSet := map[trinary.Trytes]struct{}{}
	for i := range txs {
		bundleHashesSet[txs[i].Bundle] = struct{}{}
	}

	bundleHashes := make([]trinary.Trytes, len(bundleHashesSet))
	for hash := range bundleHashesSet {
		bundleHashes = append(bundleHashes, hash)
	}

	allTxs, err := api.FindTransactionObjects(&FindTransactionsRequest{Bundles: bundleHashes})
	if err != nil {
		return nil, err
	}

	bundles := bundle.GroupTransactionsIntoBundles(allTxs)
	sort.Sort(bundle.BundlesByTimestamp(bundles))
	return bundles, err
}

func firstNonNulErr(errs ...error) error {
	for x := range errs {
		if errs[x] != nil {
			return errs[x]
		}
	}
	return nil
}

// GetAccountData returns an AccountData object containing account information about addresses,
// transactions, inputs and total account balance.
func (api *API) GetAccountData(seed trinary.Trytes, startIndex uint, endIndex uint, securityLvl signing.SecurityLevel) (*AccountData, error) {

	// 1. generate addresses up to first unused address
	unspentAddr, spentAddrs, err := api.GetUntilFirstUnusedAddress(seed, securityLvl)
	if err != nil {
		return nil, err
	}

	var err1, err2, err3 error
	var bundles bundle.Bundles
	var balances Balances
	var spentState []bool

	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		bundles, err1 = api.GetBundlesFromAddresses(spentAddrs)
	}()

	go func() {
		defer wg.Done()
		balances, err2 = api.Balances(spentAddrs)
	}()

	go func() {
		defer wg.Done()
		spentState, err3 = api.WereAddressesSpentFrom(spentAddrs...)
	}()

	wg.Wait()
	if err := firstNonNulErr(err1, err2, err3); err != nil {
		return nil, err
	}

	// get all transaction hashes of our corresponding addresses
	var txsHashes []trinary.Trytes
	for i := range bundles {
		bndl := &bundles[i]
		for j := range *bndl {
			tx := &(*bndl)[j]
			for x := range spentAddrs {
				if tx.Address == spentAddrs[x] {
					txsHashes = append(txsHashes, tx.Hash())
					break
				}
			}
		}
	}

	// compute balances
	inputs := Balances{}
	var totalBalance int64
	for i := range spentAddrs {
		value := balances[i].Value
		// this works because the balances and spent states are ordered
		if spentState[i] || value <= 0 {
			continue
		}
		totalBalance += value
		balanceCopy := balances[i]
		balanceCopy.Security = securityLvl
		inputs = append(inputs, balanceCopy)
	}

	// finally add the unused addr to the used ones
	spentAddrs = append(spentAddrs, unspentAddr)

	accountData := &AccountData{
		LatestAddress: unspentAddr,
		Transfers:     bundles,
		Transactions:  txsHashes,
		Inputs:        inputs,
		Addresses:     spentAddrs,
		Balance:       totalBalance,
	}

	return accountData, nil
}

// BroadcastBundle re-broadcasts all transactions in a bundle given the tail transaction hash.
// It might be useful when transactions did not properly propagate, particularly in the case of large bundles.
func (api *API) BroadcastBundle(tailTransactionHash trinary.Trytes) error {
	bndl, err := api.GetBundle(tailTransactionHash)
	if err != nil {
		return err
	}
	return api.BroadcastTransactions(bndl)
}

// TraverseBundle fetches the bundle of the given tail transaction hash by traversing through the trunk transactions.
// It does not validate the bundle.
func (api *API) TraverseBundle(tailTransactionHash trinary.Trytes) (bundle.Bundle, error) {
	txs, err := api.GetTransactionObjects(tailTransactionHash)
	if err != nil {
		return nil, err
	}

	tx := txs[0]

	// check whether we actually got the tail transaction passed in
	if tx.CurrentIndex != 0 {
		return nil, ErrInvalidTailTransactionHash
	}

	txsInBundle := int(tx.LastIndex + 1)
	bndl := make(bundle.Bundle, txsInBundle)
	bndl[0] = tx

	for i := 1; i < txsInBundle; i++ {
		txs, err := api.GetTransactionObjects(tailTransactionHash)
		if err != nil {
			return nil, err
		}

		tx = txs[0]
		bndl[i] = tx
	}

	return bndl, nil
}

// GetBundle fetches and validates the bundle given a tail transaction hash, by calling TraverseBundle()
func (api *API) GetBundle(tailTransactionHash trinary.Trytes) (bundle.Bundle, error) {
	bndl, err := api.TraverseBundle(tailTransactionHash)
	if err != nil {
		return nil, err
	}

	return bndl, bndl.IsValid()
}

type BroadcastTransactionsRequest struct {
	Command string                    `json:"command"`
	Trytes  []transaction.Transaction `json:"trytes"`
}

// BroadcastTransactions broadcasts a list of attached transaction trytes to the network.
// Tip selection and Proof-of-Work must be done first, by calling GetTransactionsToApprove() and AttachToTangle() or
// an equivalent attach method or remote PoWbox, which is a development tool.
// You may use this method to increase odds of effective transaction propagation.
//
// Persist the transaction trytes in local storage before calling this command for first time, to ensure that
// reattachment is possible, until your bundle has been included.
func (api *API) BroadcastTransactions(trytes []transaction.Transaction) error {
	err := api.do(&struct {
		Command string                    `json:"command"`
		Trytes  []transaction.Transaction `json:"trytes"`
	}{
		"broadcastTransactions",
		trytes,
	}, nil)

	return err
}

type StoreTransactionsRequest struct {
	Command string                    `json:"command"`
	Trytes  []transaction.Transaction `json:"trytes"`
}

// StoreTransactions persists a list of attached transaction trytes in the store of the connected node.
// Tip selection and Proof-of-Work must be done first, by calling
// getTransactionsToApprove and attachToTangle or an equivalent attach method or remote PoWbox.
//
// Persist the transaction trytes in local storage before calling this command, to ensure reattachment is possible, until your bundle has been included.
// Any transactions stored with this command will eventually be erased, as a result of a snapshot.
func (api *API) StoreTransactions(trytes transaction.Transactions) error {
	err := api.do(&struct {
		Command string                    `json:"command"`
		Trytes  []transaction.Transaction `json:"trytes"`
	}{
		"storeTransactions",
		trytes,
	}, nil)

	return err
}

// StoreAndBroadcast stores and broadcasts a list of attached transaction trytes by calling StoreTransactions() and BroadcastTransactions().
//
// Note: Persist the transaction trytes in local storage before calling this command, to ensure that reattachment is possible,
// until your bundle has been included.
//
// Any transactions stored with this command will eventually be erased, as a result of a snapshot.
func (api *API) StoreAndBroadcast(trytes transaction.Transactions) error {
	if err := api.StoreTransactions(trytes); err != nil {
		return err
	}
	if err := api.BroadcastTransactions(trytes); err != nil {
		return err
	}
	return nil
}

// GetLatestInclusion fetches inclusion states of given transactions and a list of tips
// by calling GetInclusionStates on the latest solid subtangle milestone (provided by GetNodeInfo()).
func (api *API) GetLatestInclusion(hash []trinary.Trytes) ([]bool, error) {
	var gt *GetTrytesResponse
	var ni *GetNodeInfoResponse
	var err1, err2 error

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		gt, err1 = api.GetTrytes(hash...)
	}()

	go func() {
		defer wg.Done()
		ni, err2 = api.GetNodeInfo()
	}()
	wg.Wait()

	switch {
	case err1 != nil:
		return nil, err1
	case err2 != nil:
		return nil, err2
	case len(gt.Trytes) == 0:
		return nil, ErrTransactionNotFound
	}

	resp, err := api.GetInclusionStates(hash, []trinary.Trytes{ni.LatestMilestone})
	if err != nil {
		return nil, err
	}

	if len(resp.States) == 0 {
		return nil, ErrTransactionNotFoundInInclusionState
	}
	return resp.States, nil
}

// GetNewAddress generates and returns a new address by calling FindTransactions() and WereAddressesSpentFrom() until the first
// unused address is detected. This stops working after a snapshot.
func (api *API) GetNewAddress(seed trinary.Trytes, security signing.SecurityLevel) (signing.Address, error) {
	unusedAddr, _, err := api.GetUntilFirstUnusedAddress(seed, security)
	return unusedAddr, err
}

// GetUntilFirstUnusedAddress generates and returns all addresses up to the first unused addresses including it.
// An unused address is an address which didn't receive any transactions and is unspent.
func (api *API) GetUntilFirstUnusedAddress(seed trinary.Trytes, security signing.SecurityLevel) (signing.Address, []signing.Address, error) {
	var all []signing.Address
	var index uint
	for ; ; index++ {
		adr, err := signing.NewAddress(seed, index, security)
		if err != nil {
			return "", nil, err
		}

		var err1, err2 error
		var findTxResp *FindTransactionsResponse
		var spentStates []bool
		wg := sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			findTxResp, err1 = api.FindTransactions(&FindTransactionsRequest{
				Addresses: []signing.Address{adr},
			})
		}()

		go func() {
			defer wg.Done()
			spentStates, err2 = api.WereAddressesSpentFrom(adr)
		}()
		wg.Wait()

		if err := firstNonNulErr(err1, err2); err != nil {
			return "", nil, err
		}

		if len(findTxResp.Hashes) == 0 && spentStates[0] == false {
			return adr, all, nil
		}

		// reached the end of the loop, so must be used address, repeat until return
		all = append(all, adr)
	}
}

// GetInputs gets all possible inputs of a seed and returns them with the total balance. end must be under start+500.
func (api *API) GetInputs(seed trinary.Trytes, start, end uint, security signing.SecurityLevel) (Balances, error) {
	var err error
	var addrs []signing.Address

	if start > end || end > (start+500) {
		return nil, ErrInvalidAddressStartEnd
	}

	switch {
	case end > 0:
		addrs, err = signing.NewAddresses(seed, start, end-start, security)
	default:
		_, addrs, err = api.GetUntilFirstUnusedAddress(seed, security)
	}

	if err != nil {
		return nil, err
	}

	return api.Balances(addrs)
}

// gets all balances of the given inputs. the supplied total must be less or equal to the actual computed balance.
func (api *API) verifyInputs(seed trinary.Trytes, inputs bundle.AddressInputs, security signing.SecurityLevel, total int64) (Balances, bundle.AddressInputs, error) {
	var balances Balances
	var err error

	if inputs == nil {
		return nil, nil, ErrInputsNotProvided
	}

	// gather all addresses and balances from the provided address inputs
	addrs := make([]signing.Address, len(inputs))
	for i, ai := range inputs {
		addrs[i], err = ai.Address()
		if err != nil {
			return nil, nil, err
		}
	}

	// validate the inputs by calling getBalances
	balances, err = api.Balances(addrs)
	if err != nil {
		return nil, nil, err
	}

	// not enough balance
	if total > balances.Total() {
		return nil, nil, ErrNotEnoughBalance
	}
	return balances, inputs, nil
}

// PrepareTransfers gets an array of transfer objects as input, and then prepares the transfer by generating the correct
// bundle as well as choosing and signing the inputs if necessary (if it is a value transfer).
func (api *API) PrepareTransfers(seed trinary.Trytes, transfers bundle.Transfers, inputs bundle.AddressInputs, remainder signing.Address, security signing.SecurityLevel) (bundle.Bundle, error) {
	var err error

	bndl, outputTxsFragments, total := transfers.CreateBundle()

	// if no input is required, don't sign and simply finalize the bundle (0 value transfer)
	if total <= 0 {
		bndl.Finalize(outputTxsFragments)
		return bndl, nil
	}

	// check whether the balance of all inputs is sufficient for the transfer
	balances, inputs, err := api.verifyInputs(seed, inputs, security, total)
	if err != nil {
		return nil, err
	}

	// add the inputs to the bundle(and the remainder address if needed)
	if err := api.AddInputsToBundle(seed, security, balances, &bndl, remainder, total); err != nil {
		return nil, err
	}

	bndl.Finalize(outputTxsFragments)
	err = bndl.SignInputs(inputs)
	return bndl, err
}

// AddInputsToBundle adds the given inputs to the bundle and if there's a remainder, adds an additional output
// transaction for it. If no remainder address is supplied GetUntilFirstUnusedAddress() is used to determine the
// remainder address. After calling this function, the provided bundle is ready for finalisation.
func (api *API) AddInputsToBundle(seed trinary.Trytes, security signing.SecurityLevel, inputs Balances, bundle *bundle.Bundle, remainderAddr signing.Address, transferValue int64) error {
	var err error
	var totalInput int64
	for i := range inputs {
		balance := &inputs[i]

		// add an entry for the input address (the value is negated because inputs are withdraws)
		bundle.AddEntry(int(security), balance.Address, -balance.Value, time.Now(), "")
		totalInput += balance.Value
	}

	remainder := totalInput - transferValue
	if remainder < 0 {
		return ErrInputsInsufficientForBundle
	}

	// inputs got spent entirely
	if remainder == 0 {
		return nil
	}

	// generate a new remainder address as non was supplied
	if len(remainderAddr) == 0 {
		remainderAddr, _, err = api.GetUntilFirstUnusedAddress(seed, security)
		if err != nil {
			return err
		}
	}

	bundle.AddEntry(1, remainderAddr, remainder, time.Now(), curl.EmptyHash)
	return nil
}

// SendTrytes attaches, stores and broadcasts a list of transaction trytes to the Tangle.
func (api *API) SendTrytes(depth int, trytes bundle.Bundle, mwm int64, pow pow.PowFunc, reference ...trinary.Trytes) (bundle.Bundle, error) {
	var ref trinary.Trytes
	if len(reference) > 0 {
		ref = reference[0]
	}
	tra, err := api.GetTransactionsToApprove(depth, ref)
	if err != nil {
		return nil, err
	}

	// if no powFunc is supplied, let the remote connected node do the proof of work
	if pow == nil {
		at := AttachToTangleRequest{
			TrunkTransaction:   tra.TrunkTransaction,
			BranchTransaction:  tra.BranchTransaction,
			MinWeightMagnitude: mwm,
			Trytes:             trytes,
		}

		attached, err := api.AttachToTangle(&at)
		if err != nil {
			return nil, err
		}

		trytes = attached.Trytes
	} else {
		if err := bundle.DoPoW(tra.TrunkTransaction, tra.BranchTransaction, trytes, mwm, pow); err != nil {
			return nil, err
		}
	}

	if err := api.BroadcastTransactions(trytes); err != nil {
		return nil, err
	}

	if err := api.StoreTransactions(transaction.Transactions(trytes)); err != nil {
		return nil, err
	}

	return trytes, nil
}

func GenerateEmptySpamTransaction() bundle.Bundle {
	bndl, _, _ := bundle.Transfers{
		{
			Address: signing.EmptyAddress,
			Tag:     "PROMOTE",
			Value:   0,
			Message: "",
		},
	}.CreateBundle()
	return bndl
}

// PromoteTransaction promotes a transaction by adding other transactions (spam by default) on top of it.
// It will promote maximum transfers on top of the current one with delay interval.
// Promotion can be interrupted through the passed in ctx.
func (api *API) PromoteTransaction(tail trinary.Trytes, depth int, trytes bundle.Bundle, mwm int64, pow pow.PowFunc) error {
	if len(trytes) == 0 {
		return ErrEmptyTransferForPromote
	}

	resp, err := api.CheckConsistency(tail)
	if err != nil {
		return err
	} else if !resp.State {
		return ErrInconsistentSubtangle
	}

	tips, err := api.GetTransactionsToApprove(depth, tail)
	if err != nil {
		return err
	}

	switch {
	case pow == nil:
		at := AttachToTangleRequest{
			TrunkTransaction:   tips.TrunkTransaction,
			BranchTransaction:  tips.BranchTransaction,
			MinWeightMagnitude: mwm,
			Trytes:             trytes,
		}

		// attach to tangle - do pow
		attached, err := api.AttachToTangle(&at)
		if err != nil {
			return err
		}

		trytes = attached.Trytes
	default:
		err := bundle.DoPoW(tips.TrunkTransaction, tips.BranchTransaction, trytes, mwm, pow)
		if err != nil {
			return err
		}
	}

	if err := api.BroadcastTransactions(trytes); err != nil {
		return err
	}

	return api.StoreTransactions(transaction.Transactions(trytes))
}

// Send sends tokens. If you need to do pow locally, you must specify pow func, otherwise this calls the AttachToTangle API
func (api *API) Send(seed trinary.Trytes, security signing.SecurityLevel, depth int, transfers bundle.Transfers, mwm int64, pow pow.PowFunc) (bundle.Bundle, error) {
	bd, err := api.PrepareTransfers(seed, transfers, nil, "", security)
	if err != nil {
		return nil, err
	}

	return api.SendTrytes(depth, []transaction.Transaction(bd), mwm, pow)
}

// GetTransactionObjects fetches transaction objects, given an array of transaction hashes.
func (api *API) GetTransactionObjects(txHashes ...trinary.Trytes) (transaction.Transactions, error) {
	res, err := api.GetTrytes(txHashes...)
	if err != nil {
		return nil, err
	}

	txs := transaction.Transactions{}
	for i := range res.Trytes {
		tx, err := transaction.NewTransaction(res.Trytes[i])
		if err != nil {
			return nil, err
		}
		txs = append(txs, *tx)
	}

	return txs, nil
}

// FindTransactionObjects is a wrapper function for findTransactions and getTrytes.
// Searches for transactions given a query object with addresses, tags and approvees fields.
// Multiple query fields are supported and findTransactionObjects returns intersection of results
func (api *API) FindTransactionObjects(findTxsReq *FindTransactionsRequest) (transaction.Transactions, error) {
	findTxResp, err := api.FindTransactions(findTxsReq)
	if err != nil {
		return nil, err
	}
	return api.GetTransactionObjects(findTxResp.Hashes...)
}

const MilestoneInterval = 2 * 60 * 1000
const OneWayDelay = 1 * 60 * 1000
const maxDepth = 6

// checks whether by the given timestamp the transaction is to deep to be promoted
func isAboveMaxDepth(attachmentTimestamp trinary.Trytes) bool {
	nowMilli := time.Now().UnixNano() / int64(time.Millisecond)
	timestamp := attachmentTimestamp.Trits().Value()
	return timestamp < nowMilli && nowMilli-timestamp < maxDepth*MilestoneInterval*OneWayDelay
}

// IsPromotable checks if a transaction is promotable, by calling checkConsistency and verifying that attachmentTimestamp
// is above a lower bound. Lower bound is calculated based on number of milestones issued since transaction attachment.
func (api *API) IsPromotable(tailTransactionHash trinary.Trytes) (bool, error) {
	var checkConsistencyResp *CheckConsistencyResponse
	var txs transaction.Transactions
	var err1, err2 error

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		checkConsistencyResp, err1 = api.CheckConsistency(tailTransactionHash)
	}()

	go func() {
		defer wg.Done()
		txs, err2 = api.GetTransactionObjects(tailTransactionHash)
	}()
	wg.Wait()

	switch {
	case err1 != nil:
		return false, err1
	case err2 != nil:
		return false, err2
	}

	consistent := checkConsistencyResp.State
	tx := txs[0]

	return consistent && isAboveMaxDepth(tx.AttachmentTimestamp), nil
}

// ReplayBundle reattaches a transfer to tangle by selecting tips & performing the Proof-of-Work again.
// Reattachments are useful in case original transactions are pending, and can be done securely as many times as needed.
func (api *API) ReplayBundle(tailTransactionHash trinary.Trytes, depth int, mwm int64, pow pow.PowFunc, reference ...trinary.Trytes) (bundle.Bundle, error) {
	bndl, err := api.GetBundle(tailTransactionHash)
	if err != nil {
		return nil, err
	}

	return api.SendTrytes(depth, bndl, mwm, pow, reference...)
}
