package api

import (
	. "github.com/iotaledger/iota.go/trinary"
)

// IRICommand defines a command name for IRI API calls.
type IRICommand string

// API command names.
const (
	GetNodeInfoCmd              IRICommand = "getNodeInfo"
	GetNeighborsCmd             IRICommand = "getNeighbors"
	AddNeighborsCmd             IRICommand = "addNeighbors"
	RemoveNeighborsCmd          IRICommand = "removeNeighbors"
	GetTipsCmd                  IRICommand = "getTips"
	FindTransactionsCmd         IRICommand = "findTransactions"
	GetTrytesCmd                IRICommand = "getTrytes"
	GetInclusionStatesCmd       IRICommand = "getInclusionStates"
	GetBalancesCmd              IRICommand = "getBalances"
	GetTransactionsToApproveCmd IRICommand = "getTransactionsToApprove"
	AttachToTangleCmd           IRICommand = "attachToTangle"
	InterruptAttachToTangleCmd  IRICommand = "interruptAttachToTangle"
	BroadcastTransactionsCmd    IRICommand = "broadcastTransactions"
	StoreTransactionsCmd        IRICommand = "storeTransactions"
	CheckConsistencyCmd         IRICommand = "checkConsistency"
	WereAddressesSpentFromCmd   IRICommand = "wereAddressesSpentFrom"
)

type Commander interface {
	Cmd() IRICommand
}

type Command struct {
	Command IRICommand `json:"command"`
}

func (c *Command) Cmd() IRICommand {
	return c.Command
}

// AddNeighborsCommand represents the payload to the AddNeighbor API call.
type AddNeighborsCommand struct {
	Command
	URIs []string `json:"uris"`
}

// AddNeighborsResponse is the response from the AddNeighbor API call.
type AddNeighborsResponse struct {
	AddedNeighbors int64
	Duration       int64
}

// AttachToTangleCommand represents the payload to the AttachToTangle API call.
type AttachToTangleCommand struct {
	Command
	TrunkTransaction   Hash     `json:"trunkTransaction"`
	BranchTransaction  Hash     `json:"branchTransaction"`
	MinWeightMagnitude uint64   `json:"minWeightMagnitude"`
	Trytes             []Trytes `json:"trytes"`
}

// AttachToTangleResponse is the response from the AttachToTangle API call.
type AttachToTangleResponse struct {
	Trytes []Trytes `json:"trytes"`
}

// BroadcastTransactionsCommand represents the payload to the BroadcastTransactions API call.
type BroadcastTransactionsCommand struct {
	Command
	Trytes []Trytes `json:"trytes"`
}

// CheckConsistencyCommand represents the payload to the CheckConsistency API call.
type CheckConsistencyCommand struct {
	Command
	Tails Hashes `json:"tails"`
}

// CheckConsistencyResponse is the response from the CheckConsistency API call.
type CheckConsistencyResponse struct {
	State bool   `json:"state"`
	Info  string `json:"info"`
}

// FindTransactionsCommand represents the payload to the FindTransactions API call.
type FindTransactionsCommand struct {
	FindTransactionsQuery
	Command
}

// FindTransactionsResponse is the response from the FindTransactions API call.
type FindTransactionsResponse struct {
	Hashes Hashes `json:"hashes"`
}

// GetBalancesCommand represents the payload to the GetBalances API call.
type GetBalancesCommand struct {
	Command
	Addresses Hashes `json:"addresses"`
	Threshold uint64 `json:"threshold"`
	Tips      []Hash `json:"tips,omitempty"`
}

// GetBalancesResponse is the response from the GetBalances API call.
type GetBalancesResponse struct {
	Balances       []string `json:"balances"`
	Duration       int64    `json:"duration"`
	Milestone      string   `json:"milestone"`
	MilestoneIndex int64    `json:"milestoneIndex"`
}

// GetInclusionStatesCommand represents the payload to the GetInclusionStates API call.
type GetInclusionStatesCommand struct {
	Command
	Transactions Hashes `json:"transactions"`
	Tips         Hashes `json:"tips"`
}

// GetInclusionStatesResponse is the response from the GetInclusionStates API call.
type GetInclusionStatesResponse struct {
	States []bool `json:"states"`
}

// GetNeighborsCommand represents the payload to the GetNeighbors API call.
type GetNeighborsCommand struct {
	Command
}

// GetNeighborsResponse is the response from the GetNeighbors API call.
type GetNeighborsResponse struct {
	Neighbors Neighbors `json:"neighbors"`
}

// GetNodeInfoCommand represents the payload to the GetNodeInfo API call.
type GetNodeInfoCommand struct {
	Command
}

// GetNodeInfoResponse is the response from the GetNodeInfo API call.
type GetNodeInfoResponse struct {
	AppName                            string `json:"appName"`
	AppVersion                         string `json:"appVersion"`
	Duration                           int64  `json:"duration"`
	JREAvailableProcessors             int64  `json:"jreAvailableProcessors"`
	JREFreeMemory                      int64  `json:"jreFreeMemory"`
	JREMaxMemory                       int64  `json:"jreMaxMemory"`
	JRETotalMemory                     int64  `json:"jreTotalMemory"`
	LatestMilestone                    Hash   `json:"latestMilestone"`
	LatestMilestoneIndex               int64  `json:"latestMilestoneIndex"`
	LatestSolidSubtangleMilestone      Hash   `json:"latestSolidSubtangleMilestone"`
	LatestSolidSubtangleMilestoneIndex int64  `json:"latestSolidSubtangleMilestoneIndex"`
	Neighbors                          int64  `json:"neighbors"`
	PacketsQueueSize                   int64  `json:"packetsQueueSize"`
	Time                               int64  `json:"time"`
	Tips                               int64  `json:"tips"`
	TransactionsToRequest              int64  `json:"transactionsToRequest"`
}

// GetLatestSolidSubtangleMilestoneCommand represents the payload to the GetNodeInfo API call.
type GetLatestSolidSubtangleMilestoneCommand struct {
	Command
}

// GetLatestSolidSubtangleMilestoneResponse is the response from the GetNodeInfo API call
// but reduced to just the latest subtangle milestone data.
type GetLatestSolidSubtangleMilestoneResponse struct {
	LatestSolidSubtangleMilestone      Hash  `json:"latestSolidSubtangleMilestone"`
	LatestSolidSubtangleMilestoneIndex int64 `json:"latestSolidSubtangleMilestoneIndex"`
}

// GetTipsCommand represents the payload to the GetTips API call.
type GetTipsCommand struct {
	Command
}

// GetTipsResponse is the response from the GetTips API call.
type GetTipsResponse struct {
	Hashes Hashes `json:"hashes"`
}

// GetTransactionsToApproveCommand represents the payload to the GetTransactionsToApprove API call.
type GetTransactionsToApproveCommand struct {
	Command
	Depth     uint64 `json:"depth"`
	Reference Hash   `json:"reference,omitempty"`
}

// GetTransactionsToApproveResponse is the response from the GetTransactionsToApprove API call.
type GetTransactionsToApproveResponse struct {
	TransactionsToApprove
	Duration int64 `json:"duration"`
}

// GetTrytesCommand represents the payload to the GetTrytes API call.
type GetTrytesCommand struct {
	Command
	Hashes Hashes `json:"hashes"`
}

// GetTrytesResponse is the response from the GetTrytes API call.
type GetTrytesResponse struct {
	Trytes []Trytes `json:"trytes"`
}

// InterruptAttachToTangleCommand represents the payload to the InterruptAttachToTangle API call.
type InterruptAttachToTangleCommand struct {
	Command
}

// RemoveNeighborsCommand represents the payload to the RemoveNeighbors API call.
type RemoveNeighborsCommand struct {
	Command
	URIs []string `json:"uris"`
}

// RemoveNeighborsResponse is the response from the RemoveNeighbors API call.
type RemoveNeighborsResponse struct {
	RemovedNeighbors int64 `json:"removedNeighbors"`
	Duration         int64 `json:"duration"`
}

// StoreTransactionsCommand represents the payload to the StoreTransactions API call.
type StoreTransactionsCommand struct {
	Command
	Trytes []Trytes `json:"trytes"`
}

// WereAddressesSpentFromCommand represents the payload to the WereAddressesSpentFrom API call.
type WereAddressesSpentFromCommand struct {
	Command
	Addresses Hashes `json:"addresses"`
}

// WereAddressesSpentFromResponse is the response from the WereAddressesSpentFrom API call.
type WereAddressesSpentFromResponse struct {
	States []bool `json:"states"`
}
