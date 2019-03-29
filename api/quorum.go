package api

import (
	"bytes"
	"encoding/json"
	"github.com/cespare/xxhash"
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/pow"
	"github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

// QuorumLevel defines the percentage needed for a quorum.
type QuorumLevel float64

// package quorum levels
const (
	QuorumHigh   = 0.95
	QuorumMedium = 0.75
	QuorumLow    = 0.60
)

// errors produced by the quorum http provider
var (
	ErrInvalidQuorumThreshold                 = errors.New("quorum threshold is set too low, must be >0.5")
	ErrQuorumNotReached                       = errors.New("the quorum didn't reach a satisfactory result")
	ErrExceededNoResponseTolerance            = errors.New("exceeded no-response tolerance for quorum")
	ErrNotEnoughNodesForQuorum                = errors.New("at least 2 nodes must be defined for quorum")
	ErrNoLatestSolidSubtangleInfo             = errors.New("no latest solid subtangle info found")
	ErrExceededMaxSubtangleMilestoneDelta     = errors.New("exceeded max subtangle milestone delta between nodes")
	ErrNonOkStatusCodeSubtangleMilestoneQuery = errors.New("non ok status code for subtangle milestone query")
)

// MinimumQuorumThreshold is the minimum threshold the quorum settings
// must have.
const MinimumQuorumThreshold = 0.5

// NewQuorumHTTPClient creates a new quorum based Http Provider.
func NewQuorumHTTPClient(settings interface{}) (Provider, error) {
	client := &quorumhttpclient{}
	if err := client.SetSettings(settings); err != nil {
		return nil, err
	}
	return client, nil
}

// QuorumDefaults defines optional default values when a quorum couldn't be reached.
type QuorumDefaults struct {
	WereAddressesSpentFrom *bool
	GetInclusionStates     *bool
	GetBalances            *uint64
}

// QuorumHTTPClientSettings defines a set of settings for when constructing a new Http Provider.
type QuorumHTTPClientSettings struct {
	// The threshold/majority percentage which must be reached in the responses
	// to form a quorum. Define the threshold as 0<x<1, i.e. 0.8 => 80%.
	// A threshold of 1 would mean that all nodes must give the same response.
	Threshold float64

	// Defines the max percentage of nodes which can fail to give a response
	// when a quorum is built. For example, if 4 nodes are specified and
	// a NoResponseTolerance of 0.25/25% is set, then 1 node of those 4
	// is tolerated to fail to give a response and the quorum is built.
	NoResponseTolerance float64

	// For certain commands for which a quorum doesn't make sense
	// this node will be used. For example GetTransactionsToApprove
	// would always fail when queried via a quorum.
	// If no PrimaryNode is set, then a node is randomly selected from Nodes
	// for executing calls for which no quorum can be done.
	// The primary node is not used for forming the quorum and must be
	// explicitly set in the Nodes field a second time.
	PrimaryNode *string

	// The nodes to which the client connects to.
	Nodes []string

	// The underlying HTTPClient to use. Defaults to http.DefaultClient.
	Client HTTPClient

	// The Proof-of-Work implementation function. Defaults to use the AttachToTangle IRI API call.
	LocalProofOfWorkFunc pow.ProofOfWorkFunc

	// A list of commands which will be executed in quorum even though they are not
	// particularly made for such scenario. Good candidates are 'BroadcastTransactionsCmd'
	// or 'StoreTransactionsCmd'
	ForceQuorumSend map[IRICommand]struct{}

	// When querying for the latest solid subtangle milestone in quorum, MaxSubtangleMilestoneDelta
	// defines how far apart the highest and lowest latest solid subtangle milestone are allowed
	// to be a apart. A low MaxSubtangleMilestoneDelta ensures that the nodes have mostly the same ledger state.
	// A high MaxSubtangleMilestoneDelta allows for higher desynchronisation between the used nodes.
	// Recommended value is 1, meaning that the highest and lowest latest solid subtangle milestones
	// must only be apart max. 1 index.
	MaxSubtangleMilestoneDelta uint64

	// Default values which are returned when no quorum could be reached
	// for certain types of calls.
	Defaults *QuorumDefaults
}

// ProofOfWorkFunc returns the defined Proof-of-Work function.
func (hcs QuorumHTTPClientSettings) ProofOfWorkFunc() pow.ProofOfWorkFunc {
	return hcs.LocalProofOfWorkFunc
}

// QuorumHTTPClient defines an object being able to do Http calls.
type QuorumHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type quorumhttpclient struct {
	primary     Provider
	randClients []Provider
	client      HTTPClient
	nodesCount  int
	settings    *QuorumHTTPClientSettings
}

// ignore
func (hc *quorumhttpclient) SetSettings(settings interface{}) error {
	quSettings, ok := settings.(QuorumHTTPClientSettings)
	if !ok {
		return errors.Wrapf(ErrInvalidSettingsType, "expected %T", QuorumHTTPClientSettings{})
	}

	if len(quSettings.Nodes) < 2 {
		return ErrNotEnoughNodesForQuorum
	}

	// verify the urls of all given nodes
	for i := range quSettings.Nodes {
		if _, err := url.Parse(quSettings.Nodes[i]); err != nil {
			return errors.Wrap(ErrInvalidURI, quSettings.Nodes[i])
		}
	}

	// set default client
	if quSettings.Client != nil {
		hc.client = quSettings.Client
	} else {
		hc.client = http.DefaultClient
	}

	// verify that the quorum threshold makes sense
	if quSettings.Threshold != 0 {
		if quSettings.Threshold <= MinimumQuorumThreshold {
			return ErrInvalidQuorumThreshold
		}
	} else {
		quSettings.Threshold = QuorumHigh
	}

	// set the primary node of our provider
	if quSettings.PrimaryNode != nil {
		// initialize the primary client
		httpSettings := HTTPClientSettings{
			URI:                  *quSettings.PrimaryNode,
			Client:               quSettings.Client,
			LocalProofOfWorkFunc: quSettings.LocalProofOfWorkFunc,
		}

		httpProvider, err := NewHTTPClient(httpSettings)
		if err != nil {
			return err
		}
		hc.primary = httpProvider
	} else {
		// instantiate a new provider for each single node
		randClients := make([]Provider, len(quSettings.Nodes))
		for i := range quSettings.Nodes {
			httpSettings := HTTPClientSettings{
				URI:                  quSettings.Nodes[i],
				Client:               quSettings.Client,
				LocalProofOfWorkFunc: quSettings.LocalProofOfWorkFunc,
			}
			httpProvider, err := NewHTTPClient(httpSettings)
			if err != nil {
				return err
			}
			randClients[i] = httpProvider
		}
		hc.randClients = randClients
	}
	hc.nodesCount = len(quSettings.Nodes)
	hc.settings = &quSettings
	return nil
}

var nonQuorumCommands = map[IRICommand]struct{}{
	"getNodeInfo":              {},
	"getNeighbors":             {},
	"addNeighbors":             {},
	"removeNeighbors":          {},
	"getTips":                  {},
	"getTransactionsToApprove": {},
	"attachToTangle":           {},
	"interruptAttachToTangle":  {},
	"broadcastTransactions":    {},
	"storeTransactions":        {},
}

var subMileHashKey = [30]byte{34, 108, 97, 116, 101, 115, 116, 83, 111, 108, 105, 100, 83, 117, 98, 116, 97, 110, 103, 108, 101, 77, 105, 108, 101, 115, 116, 111, 110, 101}
var subMileIndexKey = [34]byte{108, 97, 116, 101, 115, 116, 83, 111, 108, 105, 100, 83, 117, 98, 116, 97, 110, 103, 108, 101, 77, 105, 108, 101, 115, 116, 111, 110, 101, 73, 110, 100, 101, 120}
var durationKey = [11]byte{34, 100, 117, 114, 97, 116, 105, 111, 110, 34, 58}
var infoKey = [7]byte{34, 105, 110, 102, 111, 34, 58}
var emptyInfoKey = [9]byte{34, 105, 110, 102, 111, 34, 58, 34, 34}
var milestoneIndexKey = [14]byte{109, 105, 108, 101, 115, 116, 111, 110, 101, 73, 110, 100, 101, 120}

const (
	commaAscii             = 44
	colonAscii             = 58
	closingCurlyBraceAscii = 125
	quoteAscii             = 34
)

var commaSep = [1]byte{commaAscii}
var quoteSep = [1]byte{quoteAscii}
var colonSep = [1]byte{colonAscii}

// reduces the data of a getNodeInfo response to just the latest milestone data
func reduceToLatestSolidSubtangleData(data []byte) ([]byte, error) {
	indexOfSubtangleHashKey := bytes.Index(data, subMileHashKey[:])
	if indexOfSubtangleHashKey == -1 {
		return nil, errors.Wrap(ErrNoLatestSolidSubtangleInfo, "subtangle milestone hash field not found")
	}
	indexOfSubtangleIndexKey := bytes.Index(data, subMileIndexKey[:])
	if indexOfSubtangleIndexKey == -1 {
		return nil, errors.Wrap(ErrNoLatestSolidSubtangleInfo, "subtangle milestone index field not found")
	}
	commaIndex := bytes.Index(data[indexOfSubtangleIndexKey:], commaSep[:])
	if commaIndex == -1 {
		return nil, errors.Wrap(ErrNoLatestSolidSubtangleInfo, "ending comma after subtangle milestone index field not found")
	}
	if indexOfSubtangleIndexKey+commaIndex > len(data) {
		return nil, errors.Wrap(ErrNoLatestSolidSubtangleInfo, "comma position larger than data")
	}
	part := data[indexOfSubtangleHashKey : indexOfSubtangleIndexKey+commaIndex]
	// add opening/closing bracket
	part = append([]byte{123}, part...)
	part = append(part, 125)
	return part, nil
}

// extracts the latest solid subtangle milestone hash of a reduced getNodeInfo response
func extractSubTangleMilestoneHash(data []byte) (string, error) {
	firstColonIndex := bytes.Index(data, colonSep[:])
	if firstColonIndex == -1 {
		return "", errors.Wrap(ErrNoLatestSolidSubtangleInfo, "unable to find hash")
	}
	leadingQuoteIndex := bytes.Index(data[firstColonIndex+1:], quoteSep[:])
	if leadingQuoteIndex == -1 {
		return "", errors.Wrap(ErrNoLatestSolidSubtangleInfo, "unable to find leading quote in hash search")
	}
	leadingQuoteIndex = firstColonIndex + 1 + leadingQuoteIndex
	endingQuoteIndex := bytes.Index(data[leadingQuoteIndex+1:], quoteSep[:])
	if endingQuoteIndex == -1 {
		return "", errors.Wrap(ErrNoLatestSolidSubtangleInfo, "unable to find ending quote in hash search")
	}
	endingQuoteIndex = leadingQuoteIndex + 1 + endingQuoteIndex
	return string(data[leadingQuoteIndex+1 : endingQuoteIndex]), nil
}

// extracts the latest solid subtangle milestone index of a reduced getNodeInfo response
func extractSubTangleMilestoneIndex(data []byte) (uint64, error) {
	lastColonIndex := bytes.LastIndex(data, colonSep[:])
	if lastColonIndex == -1 {
		return 0, errors.Wrap(ErrNoLatestSolidSubtangleInfo, "unable to find index")
	}
	index := data[lastColonIndex+1 : len(data)-1]
	return strconv.ParseUint(string(index), 10, 64)
}

// injects the optional default set data into the response
func (hc *quorumhttpclient) injectDefault(cmd interface{}, out interface{}) bool {
	// use defaults for non quorum results
	if hc.settings.Defaults == nil {
		return false
	}
	switch x := cmd.(type) {
	case *WereAddressesSpentFromCommand:
		if hc.settings.Defaults.WereAddressesSpentFrom != nil {
			states := make([]bool, len(x.Addresses))
			for i := range states {
				states[i] = *hc.settings.Defaults.WereAddressesSpentFrom
			}
			out.(*WereAddressesSpentFromResponse).States = states
			return true
		}
	case *GetInclusionStatesCommand:
		if hc.settings.Defaults.GetInclusionStates != nil {
			states := make([]bool, len(x.Transactions))
			for i := range states {
				states[i] = *hc.settings.Defaults.GetInclusionStates
			}
			out.(*GetInclusionStatesResponse).States = states
			return true
		}
	case *GetBalancesCommand:
		if hc.settings.Defaults.GetBalances != nil {
			balances := make([]string, len(x.Addresses))
			for i := range balances {
				balances[i] = strconv.Itoa(int(*hc.settings.Defaults.GetBalances))
			}
			out.(*GetBalancesResponse).Balances = balances
			return true
		}
	}
	return false
}

// slices out a byte slice without the duration field.
// querying multiple nodes will always lead to different durations
// and hence must be removed when hasing the entire response.
func sliceOutDurationField(data []byte) []byte {
	indexOfDurationField := bytes.LastIndex(data, durationKey[:])
	if indexOfDurationField == -1 {
		return data
	}
	curlyBraceIndex := bytes.Index(data[indexOfDurationField:], []byte{closingCurlyBraceAscii})
	return append(data[:indexOfDurationField-1], data[indexOfDurationField+curlyBraceIndex:]...)
}

// slices out a byte slice without the info field of check consistency calls.
// querying multiple nodes will lead to different info messages when the state
// is false of the responses.
func sliceOutInfoField(data []byte) []byte {
	infoIndex := bytes.LastIndex(data, infoKey[:])
	if infoIndex == -1 {
		return data
	}
	// if the info field is empty, don't create a copy
	if i := bytes.LastIndex(data, emptyInfoKey[:]); i != -1 {
		return data
	}
	// create a copy as we want to keep the schematics
	// of the original slice
	c := make([]byte, len(data))
	copy(c, data)
	return append(c[:infoIndex-1], closingCurlyBraceAscii)
}

// slices out a byte slice without the milestone index field for get balances calls.
// even when querying multiple nodes with the same reference, the IRI node's latest snapshot index
// is returned in the response, which must thereby be filtered out.
func sliceOutMilestoneIndexField(data []byte) []byte {
	indexOfMilestoneIndex := bytes.LastIndex(data, milestoneIndexKey[:])
	if indexOfMilestoneIndex == -1 {
		return data
	}
	c := make([]byte, len(data))
	copy(c, data)
	return append(c[:indexOfMilestoneIndex-2], closingCurlyBraceAscii)
}

type quorumcheck struct {
	votes map[uint64]*quorumvote
	mu    sync.Mutex
}

type quorumvote struct {
	votes  float64
	data   []byte
	status int
}

func (q *quorumcheck) add(hash uint64, data []byte, code int) {
	q.mu.Lock()
	_, ok := q.votes[hash]
	if ok {
		q.votes[hash].votes++
	} else {
		q.votes[hash] = &quorumvote{votes: 1, status: code, data: data}
	}
	q.mu.Unlock()
}

type subtanglecheck struct {
	highest     uint64
	highestNode *string
	lowest      uint64
	lowestHash  trinary.Hash
	lowestNode  *string
	mu          sync.Mutex
}

func (s *subtanglecheck) add(data []byte, node *string) error {
	// reduce data
	subtangleData, err := reduceToLatestSolidSubtangleData(data)
	if err != nil {
		return err
	}

	// extract index
	index, err := extractSubTangleMilestoneIndex(subtangleData)
	if err != nil {
		return err
	}

	// mutate check
	s.mu.Lock()
	defer s.mu.Unlock()
	if index < s.lowest || s.lowest == 0 {
		s.lowest = index
		hash, err := extractSubTangleMilestoneHash(subtangleData)
		if err != nil {
			return err
		}
		s.lowestNode = node
		s.lowestHash = hash
	}
	if index > s.highest || s.highest == 0 {
		s.highest = index
		s.highestNode = node
	}
	return nil
}

// ignore
func (hc *quorumhttpclient) Send(cmd interface{}, out interface{}) error {
	comm, ok := cmd.(Commander)
	if !ok {
		panic("non Commander interface passed into Send()")
	}

	// check whether we are specifically asking for the latest solid subtangle
	_, isLatestSolidSubtangleQuery := cmd.(*GetLatestSolidSubtangleMilestoneCommand)

	if !isLatestSolidSubtangleQuery {
		// execute non quorum command on the primary or random node
		command := comm.Cmd()
		_, forced := hc.settings.ForceQuorumSend[command]
		if _, ok := nonQuorumCommands[command]; ok && !forced {
			// randomly pick up as no primary is defined
			if hc.primary == nil {
				provider := hc.randClients[rand.Int()%hc.nodesCount]
				return provider.Send(cmd, out)
			}
			// use primary node
			return hc.primary.Send(cmd, out)
		}
	}

	// serialize
	b, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	// for any errors which occurred during sending the request
	errMu := sync.Mutex{}
	anyErrors := []error{}
	wg := sync.WaitGroup{}
	wg.Add(len(hc.settings.Nodes))

	// depending on the command to execute we do different checks
	var quorumCheck *quorumcheck
	var subtangleCheck *subtanglecheck

	if isLatestSolidSubtangleQuery {
		subtangleCheck = &subtanglecheck{}
	} else {
		quorumCheck = &quorumcheck{
			votes: make(map[uint64]*quorumvote),
		}
	}

	// query each not in parallel
	for i := range hc.settings.Nodes {
		go func(i int) {
			defer wg.Done()
			// add the error which occurred during this call
			var anyError error
			defer func() {
				if anyError != nil {
					errMu.Lock()
					anyErrors = append(anyErrors, anyError)
					errMu.Unlock()
				}
			}()

			rd := bytes.NewReader(b)
			req, err := http.NewRequest("POST", hc.settings.Nodes[i], rd)
			if err != nil {
				anyError = err
				return
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-IOTA-API-Version", "1")
			resp, err := hc.client.Do(req)
			if err != nil {
				anyError = err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK && isLatestSolidSubtangleQuery {
				anyError = ErrNonOkStatusCodeSubtangleMilestoneQuery;
				return
			}

			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				anyError = err
				return
			}

			// extract only latest solid subtangle data from get node info
			// call to be able to form a quorum around that response
			if isLatestSolidSubtangleQuery {
				if err := subtangleCheck.add(data, &hc.settings.Nodes[i]); err != nil {
					anyError = err
				}
				return
			}

			// remove the duration field from the response
			// as multiple nodes will always give a different answer
			data = sliceOutDurationField(data)

			var hash uint64

			switch cmd.(type) {
			case *GetBalancesCommand:
				// get balances responses contain the latest milestone index of the given node,
				// thereby it needs to be removed for hashing the result correctly
				hash = xxhash.Sum64(sliceOutMilestoneIndexField(data))
			case *FindTransactionsCommand:
				// as findTransactions responses don't guarantee ordering, we just simply
				// sum up the bytes of the reduced response. it's highly unlikely that responses
				// with different hashes will sum up to the same number.
				var sum uint64
				for _, b := range data {
					sum += uint64(b)
				}
				// shift the sum by the length of the data, thereby
				// distancing responses with different lengths
				sum *= uint64(len(data))
				hash = sum
			case *CheckConsistencyCommand:
				// we slice out the info field from check consistency calls
				// but use whatever first info response was given when actually
				// returning the result from this API call
				hash = xxhash.Sum64(sliceOutInfoField(data))
			default:
				hash = xxhash.Sum64(data)
			}
			// add quorum vote
			quorumCheck.add(hash, data, resp.StatusCode)
		}(i)
	}
	wg.Wait()

	// check how many nodes failed to give a response
	// and then check whether we violated the no-response tolerance
	errorCount := len(anyErrors)
	percOfFailedResp := float64(errorCount) / float64(hc.nodesCount)
	if percOfFailedResp > hc.settings.NoResponseTolerance {
		perc := math.Round(percOfFailedResp * 100)
		return errors.Wrapf(ErrExceededNoResponseTolerance, "%d%% of nodes failed to give a response, first error '%v'", int(perc), anyErrors[0].Error())
	}

	// when querying for the latest solid subtangle milestone,
	// we do not apply the default quorum behavior but take the MaxSubtangleMilestoneDelta into consideration.
	// note that we explicitly check the status code in the response against the NoResponseTolerance.
	if isLatestSolidSubtangleQuery {
		delta := subtangleCheck.highest - subtangleCheck.lowest
		if delta > hc.settings.MaxSubtangleMilestoneDelta {
			return errors.Wrapf(ErrExceededMaxSubtangleMilestoneDelta, "lowest node (%s) has %d, highest node (%s) has %d, max. allowed delta %d",
				*subtangleCheck.lowestNode, subtangleCheck.lowest,
				*subtangleCheck.highestNode, subtangleCheck.highest, hc.settings.MaxSubtangleMilestoneDelta)
		}
		o := out.(*GetLatestSolidSubtangleMilestoneResponse)
		o.LatestSolidSubtangleMilestone = subtangleCheck.lowestHash
		o.LatestSolidSubtangleMilestoneIndex = int64(subtangleCheck.lowest)
		return nil
	}

	var mostVotes float64
	var selected uint64
	for key, v := range quorumCheck.votes {
		if mostVotes < v.votes {
			mostVotes = v.votes
			selected = key
		}
	}

	// check whether quorum is over threshold
	percentage := mostVotes / float64(hc.nodesCount-errorCount)
	if percentage < hc.settings.Threshold {
		// automatically inject the default value set by the library user
		// in case no quorum was reached. If no defaults are set, then
		// the default error is returned indicating that no quorum was reached
		if hc.injectDefault(cmd, out) {
			return nil
		}
		return errors.Wrapf(ErrQuorumNotReached, "%0.2f of needed %0.2f reached, query (%T)", percentage, hc.settings.Threshold, cmd)
	}

	// extract final result and status code
	statusCode := quorumCheck.votes[selected].status
	result := quorumCheck.votes[selected].data

	if statusCode != http.StatusOK {
		errResp := &ErrRequestError{Code: statusCode}
		json.Unmarshal(result, errResp)
		return errResp
	}

	if out == nil {
		return nil
	}
	return json.Unmarshal(result, out)
}
