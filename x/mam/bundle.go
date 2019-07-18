package mam

/*
#cgo CFLAGS: -Imam -Ientangled -Iuthash/include
#cgo LDFLAGS: -L. -lmam -lkeccak
#include <mam/api/api.h>
#include <common/trinary/flex_trit.h>

int num_flex_trits_for_trits(int trits_count) {
	return NUM_FLEX_TRITS_FOR_TRITS(trits_count);
}

bundle_transactions_t* new_bundle_transactions(){
	bundle_transactions_t *bundle = NULL;
	bundle_transactions_new(&bundle);
	return  bundle;
}

*/
import "C"

import (
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/transaction"
	"unsafe"
)

type c_Bundle = C.bundle_transactions_t;

var numOfFlexTritsForTxTrits = int(C.num_flex_trits_for_trits(C.int(consts.TransactionTrytesSize * 3)))

// converts a Go bundle to a C bundle. The C bundle's memory must be released manually.
func goBundleToCBundle(bndl bundle.Bundle) *c_Bundle {
	c_bundle := C.new_bundle_transactions()

	txsTrytes := transaction.MustTransactionsToTrytes(bndl)

	for _, txTrytes := range txsTrytes {
		trytesLen := len(txTrytes)
		c_trytesLen := C.size_t(trytesLen)

		// allocate buffer holding the flex trits
		flexTritBuf := make([]int8, numOfFlexTritsForTxTrits)

		// convert to C types
		c_numOfFlexTritsForTrits := C.size_t(numOfFlexTritsForTxTrits)
		c_flexTrits := (*C.flex_trit_t)(unsafe.Pointer(&flexTritBuf[0]))
		c_txTrytes := (*C.tryte_t)(unsafe.Pointer(&[]byte(txTrytes)[0]))

		// convert trnasaction trytes to its flex trits representation
		C.flex_trits_from_trytes(c_flexTrits, c_numOfFlexTritsForTrits, c_txTrytes, c_trytesLen, c_trytesLen)

		// deserialize flex trits to C transaction object
		c_tx := &C.iota_transaction_t{}
		C.transaction_deserialize_from_trits(c_tx, c_flexTrits, C._Bool(false))

		// now actually add it to the bundle
		C.bundle_transactions_add(c_bundle, c_tx)
	}

	return c_bundle
}

// converts a C bundle to a Go bundle and frees up the C bundle.
func cBundleToGoBundle(c_bundle *c_Bundle) (bundle.Bundle, error) {
	bndlLength := int(C.bundle_transactions_size(c_bundle))
	bndl := bundle.Bundle{}
	for i := 0; i < bndlLength; i++ {
		// get C transaction and convert it to flex trits
		c_tx := C.bundle_at(c_bundle, C.size_t(i))
		c_flexTrits := C.transaction_serialize(c_tx)

		// allocate buffer holding final tx trytes
		var txTrytesBuf [consts.TransactionTrytesSize]byte
		c_trytes := (*C.tryte_t)(unsafe.Pointer(&txTrytesBuf[0]))
		c_trytesLen := C.size_t(consts.TransactionTrytesSize)

		// convert transaction flex trits to their trytes representation
		c_numOfFlexTritsForTrits := C.size_t(numOfFlexTritsForTxTrits)
		C.flex_trits_to_trytes(c_trytes, c_trytesLen, c_flexTrits, c_numOfFlexTritsForTrits, c_numOfFlexTritsForTrits)

		tx, err := transaction.AsTransactionObject(string(txTrytesBuf[:]))
		if err != nil {
			return nil, err
		}
		bndl = append(bndl, *tx)
	}

	C.free(unsafe.Pointer(c_bundle))

	return bndl, nil
}
