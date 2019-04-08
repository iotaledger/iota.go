package store

import (
	"encoding/gob"
	"github.com/iotaledger/iota.go/account/deposit"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/guards"
	"github.com/iotaledger/iota.go/transaction"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
	"time"
)

func init() {
	gob.Register(AccountState{})
}

func NewAccountState() *AccountState {
	return &AccountState{
		DepositAddresses: make(map[uint64]*StoredDepositAddress, 0),
		PendingTransfers: make(map[string]*PendingTransfer, 0),
	}
}

// AccountState is the underlying representation of the account data.
type AccountState struct {
	KeyIndex         uint64                           `json:"key_index" bson:"key_index"`
	DepositAddresses map[uint64]*StoredDepositAddress `json:"deposit_addresses" bson:"deposit_addresses"`
	PendingTransfers map[string]*PendingTransfer      `json:"pending_transfers" bson:"pending_transfers"`
}

func (state *AccountState) IsNew() bool {
	return len(state.DepositAddresses) == 0 && len(state.PendingTransfers) == 0
}

// ExportedAccountState represents an exported account state.
type ExportedAccountState struct {
	ID   string    `json:"id" bson:"_id"`
	Date time.Time `json:"date" bson:"date"`
	AccountState
}

// PendingTransfer defines a pending transfer in the store which is made up of the bundle's
// essence trytes and tail hashes of reattachments.
type PendingTransfer struct {
	Bundle []Trytes `json:"bundle" bson:"bundle"`
	Tails  Hashes   `json:"tails" bson:"tails"`
}

// StoredDepositAddress defines a stored deposit address.
// It differs from the normal deposit address only in having an additional field to hold the security level
// used to generate the deposit address.
type StoredDepositAddress struct {
	deposit.Conditions `bson:"inline"`
	SecurityLevel      consts.SecurityLevel `json:"security_level" bson:"security_level"`
}

// errors produced by the store package.
var (
	ErrAccountNotFound         = errors.New("account not found")
	ErrPendingTransferNotFound = errors.New("pending transfer not found")
	ErrDepositAddressNotFound  = errors.New("deposit address not found")
)

// Store defines a persistence layer which takes care of storing account data.
type Store interface {
	// LoadAccount loads an existing or allocates a new account state from/in the database and returns the state.
	LoadAccount(id string) (*AccountState, error)
	// RemoveAccount removes the account with the given id from the store.
	RemoveAccount(id string) error
	// ImportAccount imports the given account state into the persistence layer.
	// An existing account is overridden by this method.
	ImportAccount(state ExportedAccountState) error
	// ExportAccount exports the given account from the persistence layer.
	ExportAccount(id string) (*ExportedAccountState, error)
	// Returns the last used key index for the given account.
	ReadIndex(id string) (uint64, error)
	// WriteIndex stores the given index as the last used key index for the given account.
	WriteIndex(id string, index uint64) error
	// AddDepositAddress stores the deposit address under the given account with the used key index.
	AddDepositAddress(id string, index uint64, depositAddress *StoredDepositAddress) error
	// RemoveDepositAddress removes the deposit address with the given key index under the given account.
	RemoveDepositAddress(id string, index uint64) error
	// GetDepositAddresses loads the stored deposit addresses of the given account.
	GetDepositAddresses(id string) (map[uint64]*StoredDepositAddress, error)
	// AddPendingTransfer stores the pending transfer under the given account with the origin tail tx hash as a key and
	// removes all deposit addresses which correspond to the used key indices for the transfer.
	AddPendingTransfer(id string, originTailTxHash Hash, bundleTrytes []Trytes, indices ...uint64) error
	// RemovePendingTransfer removes the pending transfer with the given origin tail transaction hash
	// from the given account.
	RemovePendingTransfer(id string, originTailTxHash Hash) error
	// AddTailHash adds the given new tail transaction hash (presumably from a reattachment) under the given pending transfer
	// indexed by the given origin tail transaction hash.
	AddTailHash(id string, originTailTxHash Hash, newTailTxHash Hash) error
	// GetPendingTransfers returns all pending transfers of the given account.
	GetPendingTransfers(id string) (map[string]*PendingTransfer, error)
}

// TrytesToPendingTransfer converts the given trytes to its essence trits.
func TrytesToPendingTransfer(trytes []Trytes) PendingTransfer {
	essences := make([]Trytes, len(trytes))
	for i := 0; i < len(trytes); i++ {
		// if the transaction has a non empty signature message fragment, we store it in the store
		storeSigMsgFrag := !guards.IsEmptyTrytes(trytes[i][:consts.AddressTrinaryOffset/3])
		if storeSigMsgFrag {
			essences[i] = trytes[i][:consts.BundleTrinaryOffset/3]
		} else {
			essences[i] = trytes[i][consts.AddressTrinaryOffset/3 : consts.BundleTrinaryOffset/3]
		}
	}
	return PendingTransfer{Bundle: essences, Tails: Hashes{}}
}

// PendingTransferToBundle converts bundle essences to a (incomplete) bundle.
func PendingTransferToBundle(pt *PendingTransfer) (bundle.Bundle, error) {
	bndl := make(bundle.Bundle, len(pt.Bundle))
	in := 0
	for i := 0; i < len(bndl); i++ {
		essenceTrytes := pt.Bundle[i]
		// add empty trits for fields after the last index
		txTrytes := essenceTrytes + Pad("", (consts.TransactionTrinarySize-consts.BundleTrinaryOffset)/3)
		// add an empty signature message fragment if non was stored
		if len(txTrytes) != consts.TransactionTrinarySize/3 {
			txTrytes = Pad("", consts.SignatureMessageFragmentTrinarySize/3) + txTrytes
		}
		tx, err := transaction.ParseTransaction(MustTrytesToTrits(txTrytes), true)
		if err != nil {
			return nil, err
		}
		bndl[in] = *tx
		in++
	}
	b, err := bundle.Finalize(bndl)
	if err != nil {
		panic(err)
	}
	return b, nil
}
