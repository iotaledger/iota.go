package inmemory

import (
	"encoding/json"
	"github.com/iotaledger/iota.go/account/deposit"
	. "github.com/iotaledger/iota.go/account/store"
	"github.com/iotaledger/iota.go/trinary"
	"sync"
	"time"
)

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		accs: map[string]*AccountState{},
	}
}

type InMemoryStore struct {
	muAccs sync.Mutex
	accs   map[string]*AccountState
}

func (mem *InMemoryStore) Clear() {
	mem.muAccs.Lock()
	defer mem.muAccs.Unlock()
	mem.accs = map[string]*AccountState{}
}

func (mem *InMemoryStore) Dump() []byte {
	mem.muAccs.Lock()
	defer mem.muAccs.Unlock()
	dump, err := json.MarshalIndent(mem.accs, "", "   ")
	if err != nil {
		panic(err)
	}
	return dump
}
func (mem *InMemoryStore) LoadAccount(id string) (*AccountState, error) {
	mem.muAccs.Lock()
	defer mem.muAccs.Unlock()
	state, ok := mem.accs[id]
	if !ok {
		mem.accs[id] = NewAccountState()
		return mem.accs[id], nil
	}
	return state, nil
}

func (mem *InMemoryStore) RemoveAccount(id string) error {
	mem.muAccs.Lock()
	defer mem.muAccs.Unlock()
	_, ok := mem.accs[id]
	if !ok {
		return ErrAccountNotFound
	}
	delete(mem.accs, id)
	return nil
}

func (mem *InMemoryStore) ImportAccount(state ExportedAccountState) error {
	mem.muAccs.Lock()
	defer mem.muAccs.Unlock()
	mem.accs[state.ID] = &state.AccountState
	return nil
}

func (mem *InMemoryStore) ExportAccount(id string) (*ExportedAccountState, error) {
	mem.muAccs.Lock()
	defer mem.muAccs.Unlock()
	state, ok := mem.accs[id]
	if !ok {
		return nil, ErrAccountNotFound
	}
	stateCopy := AccountState{KeyIndex: state.KeyIndex}
	stateCopy.DepositAddresses = map[uint64]*StoredDepositAddress{}
	stateCopy.PendingTransfers = map[string]*PendingTransfer{}

	// copy deposit addresses
	for index, depositAddr := range state.DepositAddresses {
		depositAddressesCopy := &StoredDepositAddress{SecurityLevel: depositAddr.SecurityLevel}
		addrCopy := deposit.Conditions{MultiUse: depositAddr.MultiUse}
		if depositAddr.ExpectedAmount != nil {
			expectedAmountCopy := *depositAddr.ExpectedAmount
			addrCopy.ExpectedAmount = &expectedAmountCopy
		}
		if depositAddr.TimeoutAt != nil {
			timeoutAtCopy := *depositAddr.TimeoutAt
			addrCopy.TimeoutAt = &timeoutAtCopy
		}
		depositAddressesCopy.Conditions = addrCopy
		stateCopy.DepositAddresses[index] = depositAddressesCopy
	}
	// copy pending transfers
	for tailTx, pendingTransfer := range state.PendingTransfers {
		pendingTransferCopy := PendingTransfer{}
		if pendingTransfer.Bundle != nil {
			pendingTransferCopy.Bundle = make([]trinary.Trytes, len(pendingTransfer.Bundle))
			copy(pendingTransferCopy.Bundle, pendingTransfer.Bundle)
		}
		if pendingTransfer.Tails != nil {
			pendingTransferCopy.Tails = make(trinary.Hashes, len(pendingTransfer.Tails))
			copy(pendingTransferCopy.Tails, pendingTransfer.Tails)
		}
		stateCopy.PendingTransfers[tailTx] = &pendingTransferCopy
	}
	return &ExportedAccountState{ID: id, Date: time.Now(), AccountState: stateCopy}, nil
}

func (mem *InMemoryStore) ReadIndex(id string) (uint64, error) {
	mem.muAccs.Lock()
	defer mem.muAccs.Unlock()
	state, ok := mem.accs[id]
	if !ok {
		return 0, ErrAccountNotFound
	}
	return state.KeyIndex, nil
}

func (mem *InMemoryStore) WriteIndex(id string, index uint64) (error) {
	mem.muAccs.Lock()
	defer mem.muAccs.Unlock()
	state, ok := mem.accs[id]
	if !ok {
		return ErrAccountNotFound
	}
	state.KeyIndex = index
	return nil
}

func (mem *InMemoryStore) AddDepositAddress(id string, index uint64, depositAddress *StoredDepositAddress) error {
	mem.muAccs.Lock()
	defer mem.muAccs.Unlock()
	state, ok := mem.accs[id]
	if !ok {
		return ErrAccountNotFound
	}
	state.DepositAddresses[index] = depositAddress
	return nil
}

func (mem *InMemoryStore) RemoveDepositAddress(id string, index uint64) error {
	mem.muAccs.Lock()
	defer mem.muAccs.Unlock()
	state, ok := mem.accs[id]
	if !ok {
		return ErrAccountNotFound
	}
	_, ok = state.DepositAddresses[index]
	if !ok {
		return ErrDepositAddressNotFound
	}
	delete(state.DepositAddresses, index)
	return nil
}

func (mem *InMemoryStore) AddPendingTransfer(id string, tailTx trinary.Hash, bundleTrytes []trinary.Trytes, indices ...uint64) error {
	mem.muAccs.Lock()
	defer mem.muAccs.Unlock()
	state, ok := mem.accs[id]
	if !ok {
		return ErrAccountNotFound
	}

	// remove used deposit actions
	for _, index := range indices {
		delete(state.DepositAddresses, index)
	}

	pendingTransfer := TrytesToPendingTransfer(bundleTrytes)
	pendingTransfer.Tails = append(pendingTransfer.Tails, tailTx)
	state.PendingTransfers[tailTx] = &pendingTransfer
	return nil
}

func (mem *InMemoryStore) RemovePendingTransfer(id string, tailTx trinary.Hash) error {
	mem.muAccs.Lock()
	defer mem.muAccs.Unlock()
	state, ok := mem.accs[id]
	if !ok {
		return ErrAccountNotFound
	}
	if _, ok := state.PendingTransfers[tailTx]; !ok {
		return ErrPendingTransferNotFound
	}
	delete(state.PendingTransfers, tailTx)
	return nil
}

func (mem *InMemoryStore) GetDepositAddresses(id string) (map[uint64]*StoredDepositAddress, error) {
	mem.muAccs.Lock()
	defer mem.muAccs.Unlock()
	state, ok := mem.accs[id]
	if !ok {
		return nil, ErrAccountNotFound
	}
	depositAddresses := make(map[uint64]*StoredDepositAddress)
	// make a copy
	for k, v := range state.DepositAddresses {
		// copy value which is a pointer
		copyOfDepositAddress := *v
		depositAddresses[k] = &copyOfDepositAddress
	}
	return depositAddresses, nil
}

func (mem *InMemoryStore) AddTailHash(id string, tailTx trinary.Hash, newTailTxHash trinary.Hash) error {
	mem.muAccs.Lock()
	defer mem.muAccs.Unlock()
	state, ok := mem.accs[id]
	if !ok {
		return ErrAccountNotFound
	}

	pendingTransfer, ok := state.PendingTransfers[tailTx];
	if !ok {
		return ErrPendingTransferNotFound
	}
	pendingTransfer.Tails = append(pendingTransfer.Tails, newTailTxHash)
	return nil
}

func (mem *InMemoryStore) GetPendingTransfers(id string) (map[string]*PendingTransfer, error) {
	mem.muAccs.Lock()
	defer mem.muAccs.Unlock()
	state, ok := mem.accs[id]
	if !ok {
		return nil, ErrAccountNotFound
	}
	pendingTransfers := make(map[string]*PendingTransfer)
	for k, v := range state.PendingTransfers {
		copyOfPendTrans := *v
		pendingTransfers[k] = &copyOfPendTrans
	}
	return pendingTransfers, nil
}
