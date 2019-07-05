package _examples

import (
	"fmt"
	"github.com/iotaledger/iota.go/account/store"
	"github.com/iotaledger/iota.go/account/store/inmemory"
	"log"
)

var db = inmemory.InMemoryStore{}
var id = "ABCDEFG"

// i: string, The id of the account to load.
// o: *AccountState, The state of the account.
// o: error, Returned for store errors.
func ExampleLoadAccount() {
	state, err := db.LoadAccount(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("last used key index: %d", state.KeyIndex)
}

// i: string, The id of the account to remove.
// o: error, Returned for store errors.
func ExampleRemoveAccount() {
	if err := db.RemoveAccount(id); err != nil {
		log.Fatal(err)
	}
}

// i: state, The account state to import into the store.
// o: error, Returned for store errors.
func ExampleImportAccount() {
	state := store.ExportedAccountState{}
	if err := db.ImportAccount(state); err != nil {
		log.Fatal(err)
	}
}

// o: ExportedAccountState, The account state in exported form
// o: error, Returned for store errors.
func ExampleExportAccount() {
	exportedAccountState, err := db.ExportAccount(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("exported account with id %s", exportedAccountState.ID)
}

// i: id, The id of the account.
// o: uint64, The last used key index by the account.
// o: error, Returned for store errors.
func ExampleReadIndex() {
	lastUsedKeyIndex, err := db.ReadIndex(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("last used key index: %d", lastUsedKeyIndex)
}

// i: id, The id of the account.
// i: index, The index to write as the new index of the account.
// o: error, Returned for store errors.
func ExampleWriteIndex() {
	if err := db.WriteIndex(id, 1337); err != nil {
		log.Fatal(err)
	}
}

// i: id, The id of the account.
// i: index, The index used for the deposit address.
// i: *StoredDepositAddress, The object describing the address.
func ExampleAddDepositAddress() {}

// i: id, The id of the account.
// i: index, The used index for the address to be removed from the store.
func ExampleRemoveDepositAddress() {}

// i: id, The id of the account.
// o: map[uint64]*StoredDepositAddress, The deposit addresses stored in the store.
// o: error, Returned for store errors.
func ExampleGetDepositAddress() {}

// i: id, The id of the account.
// i: originTailTxHash, The hash of the origin tail transaction.
// i: bundleTrytes, The trytes of the pending transfer.
// i: indices, The indices of the addresses which were used to fund the transfer.
// o: error, Returned for store errors.
func ExampleAddPendingTransfer() {}

// i: id, The id of the account.
// i: originTailTxHash, The hash of the origin tail transaction.
// o: error, Returned for store errors.
func ExampleRemovePendingTransfer() {}

// i: id, The id of the account.
// i; originTailTxHash, The hash of the origin tail transaction.
// i: newTailTxHash, The hash of the tail transaction to add to the store.
// o: error, returned for store errors.
func ExampleAddTailHash() {}

// i: id, The id of the account.
// o: map[string]*PendingTransfer, The pending transfers for this given account.
// o: error, Returned for store errors.
func ExampleGetPendingTransfers() {}
