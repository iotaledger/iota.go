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

// Command structure of the AddNeighbor API call.
type AddNeighborsCommand struct {
	Command IRICommand `json:"command"`
	URIs    []string   `json:"uris"`
}

// Response returned by the AddNeighbor API call.
type AddNeighborsResponse struct {
	AddedNeighbors int64
	Duration       int64
}

// Command structure of the AttachToTangle API call.
type AttachToTangleCommand struct {
	Command            IRICommand `json:"command"`
	TrunkTransaction   Hash       `json:"trunkTransaction"`
	BranchTransaction  Hash       `json:"branchTransaction"`
	MinWeightMagnitude uint64     `json:"minWeightMagnitude"`
	Trytes             []Trytes   `json:"trytes"`
}

// Response returned by the AttachToTangle API call.
type AttachToTangleResponse struct {
	Trytes []Trytes `json:"trytes"`
}

// Command structure of the BroadcastTransactions API call.
type BroadcastTransactionsCommand struct {
	Command IRICommand `json:"command"`
	Trytes  []Trytes   `json:"trytes"`
}

// Command structure of the CheckConsistency API call.
type CheckConsistencyCommand struct {
	Command IRICommand `json:"command"`
	Tails   Hashes     `json:"tails"`
}

// Response returned by the CheckConsistency API call.
type CheckConsistencyResponse struct {
	State bool   `json:"state"`
	Info  string `json:"info"`
}

// Command structure of the FindTransactions API call.
type FindTransactionsCommand struct {
	FindTransactionsQuery
	Command IRICommand `json:"command"`
}

// Response returned by the FindTransactions API call.
type FindTransactionsResponse struct {
	Hashes Hashes `json:"hashes"`
}

// Command structure of the GetBalances API call.
type GetBalancesCommand struct {
	Command   IRICommand `json:"command"`
	Addresses Hashes     `json:"addresses"`
	Threshold uint64     `json:"threshold"`
}

// Response returned by the GetBalances API call.
type GetBalancesResponse struct {
	Balances       []string `json:"balances"`
	Duration       int64    `json:"duration"`
	Milestone      string   `json:"milestone"`
	MilestoneIndex int64    `json:"milestoneIndex"`
}

// Command structure of the GetInclusionStates API call.
type GetInclusionStatesCommand struct {
	Command      IRICommand `json:"command"`
	Transactions Hashes     `json:"transactions"`
	Tips         Hashes     `json:"tips"`
}

// Response returned by the GetInclusionStates API call.
type GetInclusionStatesResponse struct {
	States []bool `json:"states"`
}

// Command structure of the GetNeighbors API call.
type GetNeighborsCommand struct {
	Command IRICommand `json:"command"`
}

// Response returned by the GetNeighbors API call.
type GetNeighborsResponse struct {
	Neighbors Neighbors `json:"neighbors"`
}

// Command structure of the GetNodeInfo API call.
type GetNodeInfoCommand struct {
	Command IRICommand `json:"command"`
}

// Response returned by the GetNodeInfo API call.
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

// Command structure of the GetTips API call.
type GetTipsCommand struct {
	Command IRICommand `json:"command"`
}

// Response returned by the GetTips API call.
type GetTipsResponse struct {
	Hashes Hashes `json:"hashes"`
}

// Command structure of the GetTransactionsToApprove API call.
type GetTransactionsToApproveCommand struct {
	Command   IRICommand `json:"command"`
	Depth     uint64     `json:"depth"`
	Reference Hash       `json:"reference,omitempty"`
}

// Response returned by the GetTransactionsToApprove API call.
type GetTransactionsToApproveResponse struct {
	TransactionsToApprove
	Duration int64 `json:"duration"`
}

// Command structure of the GetTrytes API call.
type GetTrytesCommand struct {
	Command IRICommand `json:"command"`
	Hashes  Hashes     `json:"hashes"`
}

// Response returned by the GetTrytes API call.
type GetTrytesResponse struct {
	Trytes []Trytes `json:"trytes"`
}

// Command structure of the InterruptAttachToTangle API call.
type InterruptAttachToTangleCommand struct {
	Command IRICommand `json:"command"`
}

// Command structure of the RemoveNeighbors API call.
type RemoveNeighborsCommand struct {
	Command IRICommand `json:"command"`
	URIs    []string   `json:"uris"`
}

// Response returned by the RemoveNeighbors API call.
type RemoveNeighborsResponse struct {
	RemovedNeighbors int64 `json:"removedNeighbors"`
	Duration         int64 `json:"duration"`
}

// Command structure of the StoreTransactions API call.
type StoreTransactionsCommand struct {
	Command IRICommand `json:"command"`
	Trytes  []Trytes   `json:"trytes"`
}

// Command structure of the WereAddressesSpentFrom API call.
type WereAddressesSpentFromCommand struct {
	Command   IRICommand `json:"command"`
	Addresses Hashes     `json:"addresses"`
}

// Response returned by the WereAddressesSpentFrom API call.
type WereAddressesSpentFromResponse struct {
	States []bool `json:"states"`
}
