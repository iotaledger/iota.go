package api

import (
	"context"
	"github.com/iotaledger/iota.go/bundle"
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/transaction"
	. "github.com/iotaledger/iota.go/trinary"
	"time"
)

// Defines the query object which is sent to the connected node for the FindTransactions API call.
// Using multiple fields will return the intersection of the found results.
type FindTransactionsQuery struct {
	Addresses Hashes   `json:"addresses,omitempty"`
	Approvees Hashes   `json:"approvees,omitempty"`
	Bundles   Hashes   `json:"bundles,omitempty"`
	Tags      []Trytes `json:"tags,omitempty"`
}

// A simple non negative balance.
type Balance = uint64

// An Input is an address from which to withdraw the total available balance
// to fulfill a transfer's overall output value.
type Input struct {
	Balance
	Address  Hash
	KeyIndex uint64
	Security SecurityLevel
}

// Defines balances at a given milestone.
type Balances struct {
	Balances       []uint64 `json:"balances"`
	Milestone      string   `json:"milestone"`
	MilestoneIndex int64    `json:"milestoneIndex"`
}

type Neighbors = []Neighbor

// A Neighbor is a node which is connected to the current connected node and gossips transactions with it.
type Neighbor struct {
	Address                     string
	NumberOfAllTransactions     int64
	NumberOfInvalidTransactions int64
	NumberOfNewTransactions     int64
}

// Defines tips which can be approved by a new transaction.
type TransactionsToApprove struct {
	TrunkTransaction  Hash
	BranchTransaction Hash
}

// A simple object containing an account's current state derived from the available
// data given by nodes during the current snapshot epoch.
// AccountData should not be used to represent an account's state.
type AccountData struct {
	Addresses     Hashes
	Inputs        []Input
	Transfers     bundle.Bundles
	Transactions  Hashes
	LatestAddress Hash
	Balance       uint64
}

// Options for when creating a new AccountData object via GetAccountData().
type GetAccountDataOptions struct {
	// The index from which to start creating addresses from.
	Start uint64
	// The index up to which to generate addresses to.
	End *uint64
	// The security level used for generating addresses.
	Security SecurityLevel
}

func getAccountDAtaDefaultOptions(options GetAccountDataOptions) GetAccountDataOptions {
	if options.Security == 0 {
		options.Security = SecurityLevelMedium
	}
	return options
}

// Options for when generating new addresses via GetNewAddress().
type GetNewAddressOptions struct {
	// The index from which to start creating addresses from.
	Index uint64
	// The security level used for generating new addresses.
	Security SecurityLevel
	// Whether to append the checksum to the generated addresses.
	Checksum bool
	// The total amount of addresses to generate.
	Total *uint64
	// Whether to return all generated addresses and not just the new address.
	ReturnAll bool
}

func getNewAddressDefaultOptions(options GetNewAddressOptions) GetNewAddressOptions {
	if options.Security == 0 {
		options.Security = SecurityLevelMedium
	}
	return options
}

func getInputDefaultOptions(options GetInputsOptions) GetInputsOptions {
	if options.Security == 0 {
		options.Security = SecurityLevelMedium
	}
	return options
}

// Options for when gathering Inputs via GetInputs().
type GetInputsOptions struct {
	// The index to start from.
	Start uint64
	// The index up to which to generate addresses to.
	End *uint64
	// A threshold which must be fulfilled by the gathered Inputs.
	// GetInputs() will return an error if this value can't be fulfilled.
	Threshold *uint64
	// The security level used for generating new addresses.
	Security SecurityLevel
}

// ToGetNewAddressOptions converts GetInputsOptions to GetNewAddressOptions.
func (gio GetInputsOptions) ToGetNewAddressOptions() GetNewAddressOptions {
	if gio.End != nil {
		total := *gio.End - gio.Start
		return GetNewAddressOptions{
			Index: gio.Start, Total: &total, Security: gio.Security, ReturnAll: true,
		}
	}
	return GetNewAddressOptions{
		Index: gio.Start, Security: gio.Security, ReturnAll: true,
	}
}

// Defines a set of Inputs and the total balance from them.
type Inputs struct {
	Inputs       []Input
	TotalBalance uint64
}

func getTransfersDefaultOptions(options GetTransfersOptions) GetTransfersOptions {
	if options.Security == 0 {
		options.Security = SecurityLevelMedium
	}
	return options
}

// Options for when gathering Bundles via GetTransfers().
type GetTransfersOptions struct {
	// The index from which to start creating addresses from.
	Start uint64
	// The index up to which to generate addresses to.
	End *uint64
	// Whether to set the Persistence property on retrieved transactions.
	InclusionStates bool
	// The security level used for generating new addresses.
	Security SecurityLevel
}

// ToGetNewAddressOptions converts GetTransfersOptions to GetNewAddressOptions.
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

// Options for when preparing transfers via PrepareTransfers().
type PrepareTransfersOptions struct {
	// Inputs to fulfill the transfer's output sum. If no Inputs are provided, they are collected after
	// a best effort method (not recommended). Provided inputs are not checked for spent state.
	Inputs []Input
	// The security level used when no Inputs and/or remainder address are supplied for computing
	// the corresponding addresses.
	Security SecurityLevel
	// The timestamp to use for each transaction in the resulting bundle.
	Timestamp *uint64
	// The address to send the remainder balance too. If no remainder address is supplied, then
	// the next available address is computed after a best effort method using the first Input's key index.
	RemainderAddress *Hash
}

// Options for when sending bundle transaction trytes via SendTransfer().
type SendTransfersOptions struct {
	PrepareTransfersOptions
	// A hash of a transaction to use as reference in GetTransactionsToApprove().
	Reference *Hash
}

type prepareTransferProps struct {
	Transactions     transaction.Transactions
	Trytes           []Trytes
	Transfers        bundle.Transfers
	Seed             Trytes
	Security         SecurityLevel
	Inputs           []Input
	Timestamp        uint64
	RemainderAddress *Trytes
	HMACKey          *Trytes
}

func getPrepareTransfersDefaultOptions(options PrepareTransfersOptions) PrepareTransfersOptions {
	if options.Security == 0 {
		options.Security = SecurityLevelMedium
	}
	if options.Inputs == nil {
		options.Inputs = []Input{}
	}
	return options
}

// Options for when promoting a transaction via PromoteTransaction().
type PromoteTransactionOptions struct {
	// When given a Context, PromoteTransaction() will create new promotion transactions until
	// the Context is done.
	Ctx context.Context
	// Delay between promotions. Only used if a Context is given.
	Delay *time.Duration
}

func getPromoteTransactionsDefaultOptions(options PromoteTransactionOptions) PromoteTransactionOptions {
	if options.Delay != nil && *options.Delay == 0 {
		t := time.Duration(1000) * time.Millisecond
		options.Delay = &t
	}
	return options
}

// Encapsulates errors given by the connected node or parse errors.
type ErrorResponse struct {
	Error     string `json:"error"`
	Exception string `json:"exception"`
}
