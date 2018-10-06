package api

import (
	. "github.com/iotaledger/giota/trinary"
)

type IRICommand string

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

type AddNeighborsCommand struct {
	Command IRICommand `json:"command"`
	URIs    []string   `json:"uris"`
}

type AddNeighborsResponse struct {
	AddedNeighbors int64
	Duration       int64
}

type AttachToTangleCommand struct {
	Command            IRICommand `json:"command"`
	TrunkTransaction   Hash       `json:"trunkTransaction"`
	BranchTransaction  Hash       `json:"branchTransaction"`
	MinWeightMagnitude uint64     `json:"minWeightMagnitude"`
	Trytes             []Trytes   `json:"trytes"`
}

type AttachToTangleResponse struct {
	Trytes []Trytes `json:"trytes"`
}

type BroadcastTransactionsCommand struct {
	Command IRICommand `json:"command"`
	Trytes  []Trytes   `json:"trytes"`
}

type CheckConsistencyCommand struct {
	Command IRICommand `json:"command"`
	Tails   Hashes     `json:"tails"`
}

type CheckConsistencyResponse struct {
	State bool   `json:"state"`
	Info  string `json:"info"`
}

type FindTransactionsCommand struct {
	FindTransactionsQuery
	Command IRICommand `json:"command"`
}

type FindTransactionsResponse struct {
	Hashes Hashes `json:"hashes"`
}

type GetBalancesCommand struct {
	Command   IRICommand `json:"command"`
	Addresses Hashes     `json:"addresses"`
	Threshold uint64     `json:"threshold"`
}

type GetBalancesResponse struct {
	Balances       []string `json:"balances"`
	Duration       int64    `json:"duration"`
	Milestone      string   `json:"milestone"`
	MilestoneIndex int64    `json:"milestoneIndex"`
}

type GetInclusionStateCommand struct {
	Command      IRICommand `json:"command"`
	Transactions Hashes     `json:"transactions"`
	Tips         Hashes     `json:"tips"`
}

type GetInclusionStatesResponse struct {
	States []bool `json:"states"`
}

type GetNeighborsCommand struct {
	Command IRICommand `json:"command"`
}

type GetNeighborsResponse struct {
	Neighbors Neighbors `json:"neighbors"`
}

type GetNodeInfoCommand struct {
	Command IRICommand `json:"command"`
}

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

type GetTipsCommand struct {
	Command IRICommand `json:"command"`
}

type GetTipsResponse struct {
	Hashes Hashes `json:"hashes"`
}

type GetTransactionsToApproveCommand struct {
	Command   IRICommand `json:"command"`
	Depth     uint64     `json:"depth"`
	Reference Hash       `json:"reference,omitempty"`
}

type GetTransactionsToApproveResponse struct {
	TransactionsToApprove
	Duration int64 `json:"duration"`
}

type GetTrytesCommand struct {
	Command IRICommand `json:"command"`
	Hashes  Hashes     `json:"hashes"`
}

type GetTrytesResponse struct {
	Trytes []Trytes `json:"trytes"`
}

type InterruptAttachToTangleCommand struct {
	Command IRICommand `json:"command"`
}

type RemoveNeighborsCommand struct {
	Command IRICommand `json:"command"`
	URIs    []string   `json:"uris"`
}

type RemoveNeighborsResponse struct {
	RemovedNeighbors int64 `json:"removedNeighbors"`
	Duration         int64 `json:"duration"`
}

type StoreTransactionsCommand struct {
	Command IRICommand `json:"command"`
	Trytes  []Trytes   `json:"trytes"`
}

type WereAddressesSpentFromCommand struct {
	Command   IRICommand `json:"command"`
	Addresses Hashes     `json:"addresses"`
}

type WereAddressesSpentFromResponse struct {
	States []bool `json:"states"`
}
