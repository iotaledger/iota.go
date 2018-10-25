package api

import (
	"context"
	"github.com/iotaledger/iota.go/bundle"
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/transaction"
	. "github.com/iotaledger/iota.go/trinary"
	"time"
)

type FindTransactionsQuery struct {
	Addresses Hashes   `json:"addresses,omitempty"`
	Approvees Hashes   `json:"approvees,omitempty"`
	Bundles   Hashes   `json:"bundles,omitempty"`
	Tags      []Trytes `json:"tags,omitempty"`
}

type Balance = uint64

type Address struct {
	Balance
	Address  Hash
	KeyIndex uint64
	Security SecurityLevel
}

type Balances struct {
	Balances       []uint64 `json:"balances"`
	Milestone      string   `json:"milestone"`
	MilestoneIndex int64    `json:"milestoneIndex"`
}

type Neighbors = []Neighbor
type Neighbor struct {
	Address                     string
	NumberOfAllTransactions     int64
	NumberOfInvalidTransactions int64
	NumberOfNewTransactions     int64
}

type TransactionsToApprove struct {
	TrunkTransaction  Hash
	BranchTransaction Hash
}

func getAccountDAtaDefaultOptions(options GetAccountDataOptions) GetAccountDataOptions {
	if options.Security == 0 {
		options.Security = SecurityLevelMedium
	}
	return options
}

type AccountData struct {
	Addresses     Hashes
	Inputs        []Address
	Transfers     bundle.Bundles
	Transactions  Hashes
	LatestAddress Hash
	Balance       uint64
}

type GetNewAddressOptions struct {
	Index     uint64
	Security  SecurityLevel
	Checksum  bool
	Total     *uint64
	ReturnAll bool
}

func getNewAddressDefaultOptions(options GetNewAddressOptions) GetNewAddressOptions {
	if options.Security == 0 {
		options.Security = SecurityLevelMedium
	}
	return options
}

func getInputDefaultOptions(options GetInputOptions) GetInputOptions {
	if options.Security == 0 {
		options.Security = SecurityLevelMedium
	}
	return options
}

type GetInputOptions struct {
	Start     uint64
	End       *uint64
	Threshold *uint64
	Security  SecurityLevel
}

func (gio GetInputOptions) ToGetNewAddressOptions() GetNewAddressOptions {
	if gio.End != nil {
		total := *gio.End - gio.Start
		return GetNewAddressOptions{
			Index: gio.Start, Total: &total, Security: gio.Security, ReturnAll: true,
		}
	} else {
		return GetNewAddressOptions{
			Index: gio.Start, Security: gio.Security, ReturnAll: true,
		}
	}
}

type Inputs struct {
	Inputs       []Address
	TotalBalance uint64
}

func getTransfersDefaultOptions(options GetTransfersOptions) GetTransfersOptions {
	if options.Security == 0 {
		options.Security = SecurityLevelMedium
	}
	return options
}

type GetTransfersOptions struct {
	Start           uint64
	End             *uint64
	InclusionStates bool
	Security        SecurityLevel
}

func (gto GetTransfersOptions) ToGetNewAddressOptions() GetNewAddressOptions {
	opts := GetNewAddressOptions{}
	opts.Index = gto.Start
	opts.Security = gto.Security
	opts.ReturnAll = true
	if gto.End != nil {
		total := *gto.End - gto.Start
		opts.Total = &total
	}
	return opts
}

type PrepareTransfersOptions struct {
	// Inputs to fulfill the transfer's sum. If no inputs are provided, they are collected after
	// a best efford method. Provided inputs are not checked for spent state.
	Inputs []Address
	// The used security level when no Inputs and/or remainder address are supplied for computing
	// the corresponding addresses.
	Security SecurityLevel
	// The timestamp to use for each transaction in the resulting bundle.
	Timestamp *uint64
	// The address to send the remainder balance too. If no remainder address is supplied, then
	// the next available address is computed after a best efford method.
	RemainderAddress *Hash
}

type SendTransfersOptions struct {
	PrepareTransfersOptions
	Reference *Hash
}

type PrepareTransferProps struct {
	Transactions     transaction.Transactions
	Trytes           []Trytes
	Transfers        bundle.Transfers
	Seed             Trytes
	Security         SecurityLevel
	Inputs           []Address
	Timestamp        uint64
	RemainderAddress *Trytes
	HMACKey          *Trytes
}

func getPrepareTransfersDefaultOptions(options PrepareTransfersOptions) PrepareTransfersOptions {
	if options.Security == 0 {
		options.Security = SecurityLevelMedium
	}
	if options.Inputs == nil {
		options.Inputs = []Address{}
	}
	return options
}

type PromoteTransactionOptions struct {
	// Context which is used for cancellation signals during promotion.
	Ctx context.Context
	// Delay between promotions. (only used if Ctx is supplied)
	Delay *time.Duration
}

func getPromoteTransactionsDefaultOptions(options PromoteTransactionOptions) PromoteTransactionOptions {
	if options.Delay != nil && *options.Delay == 0 {
		t := time.Duration(1000) * time.Millisecond
		options.Delay = &t
	}
	return options
}

type GetAccountDataOptions struct {
	Start    uint64
	End      *uint64
	Security SecurityLevel
}

type ErrorResponse struct {
	Error     string `json:"error"`
	Exception string `json:"exception"`
}
