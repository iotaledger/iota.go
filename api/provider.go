package api

import (
	"context"
	"github.com/iotaledger/giota/bundle"
	"github.com/iotaledger/giota/pow"
	"github.com/iotaledger/giota/signing"
	"github.com/iotaledger/giota/transaction"
	. "github.com/iotaledger/giota/trinary"
	"github.com/pkg/errors"
	"time"
)

var (
	ErrInvalidSettingsType = errors.New("incompatible settings type supplied")
)

type FindTransactionsQuery struct {
	Addresses Hashes
	Approvees []Hash
	Bundles   []Hash
	Tags      []Trytes
}

type Balance = uint64

type Address struct {
	Balance
	Address  Hash
	KeyIndex uint64
	Security signing.SecurityLevel
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
	Security  signing.SecurityLevel
	Checksum  bool
	Total     *uint64
	ReturnAll bool
}

func getNewAddressDefaultOptions(options GetNewAddressOptions) GetNewAddressOptions {
	if options.Security == 0 {
		options.Security = signing.SecurityLevelMedium
	}
	return options
}

type GetInputOptions struct {
	Start     uint64
	End       *uint64
	Threshold *uint64
	Security  signing.SecurityLevel
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

type GetTransfersOptions struct {
	Start           uint64
	End             *uint64
	InclusionStates bool
	Security        signing.SecurityLevel
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
	Inputs           []Address
	RemainderAddress *Hash
	Security         signing.SecurityLevel
	HMACKey          *Trytes
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
	Security         signing.SecurityLevel
	Inputs           []Address
	Timestamp        uint64
	RemainderAddress *Trytes
	HMACKey          *Trytes
}

func getPrepareTransfersDefaultOptions(options PrepareTransfersOptions) PrepareTransfersOptions {
	if options.Security == 0 {
		options.Security = signing.SecurityLevelMedium
	}
	if options.Inputs == nil {
		options.Inputs = []Address{}
	}
	return options
}

type PromoteTransactionOptions struct {
	Delay time.Duration
	Ctx   context.Context
}

func getPromoteTransactionsDefaultOptions(options PromoteTransactionOptions) PromoteTransactionOptions {
	if options.Delay == 0 {
		options.Delay = 1000
	}
	return options
}

type GetAccountDataOptions struct {
	Start    uint64
	End      *uint64
	Security signing.SecurityLevel
}

type ErrorResponse struct {
	Error     string `json:"error"`
	Exception string `json:"exception"`
}

type API struct {
	// API commands

	AddNeighbors             func(uris ...string) (int64, error)
	AttachToTangle           func(trunkTxHash Hash, branchTxHash Hash, mwm uint64, trytes []Trytes) ([]Trytes, error)
	BroadcastTransactions    func(trytes ...Trytes) ([]Trytes, error)
	CheckConsistency         func(hashes ...Hash) (bool, error)
	FindTransactions         func(query FindTransactionsQuery) (Hashes, error)
	GetBalances              func(addresses Hashes, threshold uint64) (*Balances, error)
	GetInclusionStates       func(txHash Hashes, tips ...Hash) ([]bool, error)
	GetNeighbors             func() (Neighbors, error)
	GetNodeInfo              func() (*GetNodeInfoResponse, error)
	GetTips                  func() (Hashes, error)
	GetTransactionsToApprove func(depth uint64, reference ...Hash) (*TransactionsToApprove, error)
	GetTrytes                func(hashes ...Hash) ([]Trytes, error)
	InterruptAttachToTangle  func() error
	RemoveNeighbors          func(uris ...string) (int64, error)
	StoreTransactions        func(trytes ...Trytes) ([]Trytes, error)
	WereAddressesSpentFrom   func(addresses ...Hash) ([]bool, error)

	// wrapper methods

	BroadcastBundle         func(tailTxHash Hash) ([]Trytes, error)
	GetAccountData          func(seed Trytes, options GetAccountDataOptions) (*AccountData, error)
	GetBundle               func(tailTxHash Hash) (bundle.Bundle, error)
	GetBundlesFromAddresses func(addresses Hashes, inclusionState ...bool) (bundle.Bundles, error)
	GetLatestInclusion      func(transactions Hashes) ([]bool, error)
	GetNewAddress           func(seed Trytes, options GetNewAddressOptions) ([]Trytes, error)
	GetTransactionObjects   func(hashes ...Hash) (transaction.Transactions, error)
	FindTransactionObjects  func(query FindTransactionsQuery) (transaction.Transactions, error)
	GetInputs               func(seed Trytes, options GetInputOptions) (*Inputs, error)
	GetTransfers            func(seed Trytes, options GetTransfersOptions) (bundle.Bundles, error)
	IsPromotable            func(tailTxHash Hash) (bool, error)
	IsReattachable          func(inputAddresses ...Trytes) ([]bool, error)
	PrepareTransfers        func(seed Trytes, transfers bundle.Transfers, options PrepareTransfersOptions) ([]Trytes, error)
	PromoteTransactions     func(tailTxHash Hash, depth uint64, mwm uint64, spamTransfers bundle.Transfers, options PromoteTransactionOptions) (transaction.Transactions, error)
	ReplayBundle            func(tailTxhash Hash, depth uint64, mwm uint64, reference ...Hash) (bundle.Bundle, error)
	SendTrytes              func(trytes []Trytes, depth uint64, mwm uint64, reference ...Hash) (bundle.Bundle, error)
	StoreAndBroadcast       func(trytes []Trytes) ([]Trytes, error)
	TraverseBundle          func(trunkTxHash Hash, bndl bundle.Bundle) (transaction.Transactions, error)
}

type CreateProviderFunc func(settings interface{}) (Provider, error)

type AttachToTangle = func(trunkTxHash Hash, branchTxHash Hash, mwm uint64, trytes []Trytes) ([]Trytes, error)

type Settings interface {
	PowFunc() pow.PowFunc
}

func ComposeAPI(settings Settings, f ...CreateProviderFunc) (*API, error) {
	var provider Provider
	var err error
	if len(f) > 0 {
		provider, err = f[0](settings)
	} else {
		provider, err = NewHttpClient(settings)
	}
	if err != nil {
		return nil, err
	}

	var attachToTangle AttachToTangle
	if settings.PowFunc() != nil {
		attachToTangle = func(trunkTxHash Hash, branchTxHash Hash, mwm uint64, trytes []Trytes) ([]Trytes, error) {
			return bundle.DoPoW(trunkTxHash, branchTxHash, trytes, mwm, settings.PowFunc())
		}
	} else {
		attachToTangle = createAttachToTangle(provider)
	}

	api := &API{}
	api.AddNeighbors = createAddNeighbors(provider)
	api.AttachToTangle = attachToTangle
	api.BroadcastTransactions = createBroadcastTransactions(provider)
	api.CheckConsistency = createCheckConsistency(provider)
	api.FindTransactions = createFindTransactions(provider)
	api.GetBalances = createGetBalances(provider)
	api.GetInclusionStates = createInclusionState(provider)
	api.GetNeighbors = createGetNeighbors(provider)
	api.GetNodeInfo = createGetNodeInfo(provider)
	api.GetTips = createGetTips(provider)
	api.GetTransactionsToApprove = createGetTransactionsToApprove(provider)
	api.GetTrytes = createGetTrytes(provider)
	api.InterruptAttachToTangle = createInterruptAttachToTangle(provider)
	api.RemoveNeighbors = createRemoveNeighbors(provider)
	api.WereAddressesSpentFrom = createWereAddressesSpentFrom(provider)

	api.BroadcastBundle = createBroadcastBundle(provider)
	api.GetAccountData = createGetAccountData(provider)
	api.GetBundle = createGetBundle(provider)
	api.GetBundlesFromAddresses = createGetBundlesFromAddresses(provider)
	api.GetLatestInclusion = createGetLatestInclusion(provider)
	api.GetNewAddress = createGetNewAddress(provider)
	api.GetTransactionObjects = createGetTransactionObjects(provider)
	api.FindTransactionObjects = createFindTransactionObjects(provider)
	api.GetInputs = createGetInputs(provider)
	api.GetTransfers = createGetTransfers(provider)
	api.IsPromotable = createIsPromotable(provider)
	api.IsReattachable = createIsReattachable(provider)
	api.PrepareTransfers = createPrepareTransfers(provider)
	api.PromoteTransactions = createPromoteTransactions(provider, attachToTangle)
	api.ReplayBundle = createReplayBundle(provider, attachToTangle)
	api.SendTrytes = createSendTrytes(provider, attachToTangle)
	api.StoreAndBroadcast = createStoreAndBroadcast(provider)
	api.TraverseBundle = createTraverseBundle(provider)

	return api, nil
}

type Provider interface {
	Send(cmd interface{}, out interface{}) error
	SetSettings(settings interface{}) error
}
