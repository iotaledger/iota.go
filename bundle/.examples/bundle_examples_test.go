package bundle_examples_test

import (
	"fmt"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/trinary"
	"strings"
	"time"
)

// i req: tag, The tag Trytes to pad.
// o: Trytes, The padded tag.
func ExamplePadTag() {}

// i req: timestamp, The timestamp (Unix epoch/seconds) for each entry/transaction.
// i req: transfers, Transfer objects to convert to BundleEntries.
// o: BundleEntries, The converted bundle entries.
// o: error, Returned for invalid addresses etc.
func ExampleTransfersToBundleEntries() {
	// Unix epoch in seconds
	ts := uint64(time.Now().UnixNano() / int64(time.Second))
	transfers := bundle.Transfers{
		{
			Address: strings.Repeat("9", 81),
			Tag:     strings.Repeat("9", 27),
			Value:   0,
			// if the message of the transfer would not fit
			// into one transaction, then TransfersToBundleEntries()
			// will create multiple entries for that transaction
			Message: "",
		},
	}
	bundleEntries, err := bundle.TransfersToBundleEntries(ts, transfers...)
	if err != nil {
		// handle error
		return
	}
	// add bundle entries to bundle with bundle.AddEntry()
	_ = bundleEntries
}

// i req: txs, The Bundle to which to add the entry to.
// i req: bndlEntry, The BundleEntry to add.
// o: Bundle, Returns a bundle with the newly added BundleEntry.
func ExampleAddEntry() {
	// Unix epoch in seconds
	ts := uint64(time.Now().UnixNano() / int64(time.Second))
	transfers := bundle.Transfers{
		{
			Address: strings.Repeat("9", 81),
			Tag:     strings.Repeat("9", 27),
			Value:   0,
			Message: "",
		},
	}
	bundleEntries, err := bundle.TransfersToBundleEntries(ts, transfers...)
	if err != nil {
		// handle error
		return
	}

	bndl := bundle.Bundle{}
	for _, entry := range bundleEntries {
		bundle.AddEntry(bndl, entry)
	}
}

// i req: bundle, The Bundle to finalize.
// o: Bundle, The finalized Bundle.
// o: error, Returned for invalid finalization.
func ExampleFinalize() {
	// Unix epoch in seconds
	ts := uint64(time.Now().UnixNano() / int64(time.Second))
	transfers := bundle.Transfers{
		{
			Address: strings.Repeat("9", 81),
			Tag:     strings.Repeat("9", 27),
			Value:   0,
			Message: "",
		},
	}
	bundleEntries, err := bundle.TransfersToBundleEntries(ts, transfers...)
	if err != nil {
		// handle error
		return
	}

	bndl := bundle.Bundle{}
	for _, entry := range bundleEntries {
		bundle.AddEntry(bndl, entry)
	}

	fmt.Println(len(bndl)) // 1

	finalizedBundle, err := bundle.Finalize(bndl)
	if err != nil {
		// handle error
		return
	}

	// finalized bundle, ready for PoW
	_ = finalizedBundle
}

// i req: bndl, The Bundle to add the Trytes to.
// i req: fragments, The Trytes fragments to add to the Bundle,
// i req: offset, The offset at which to start to add the Trytes into the Bundle.
// o: Bundle, The Bundle with the added fragments.
func ExampleAddTrytes() {
	bndl := bundle.Bundle{}
	// fragments get automatically padded
	bndl = bundle.AddTrytes(bndl, []trinary.Trytes{"ASDFEF..."}, 0)
}

// i req: bundle, The Bundle to validate.
// o: bool, Whether the signatures are valid or not.
// o: error, Returned if an error occurs during validation.
func ExampleValidateBundleSignatures() {
	bndl := bundle.Bundle{} // hypothetical finalized Bundle
	valid, err := bundle.ValidateBundleSignatures(bndl)
	if err != nil {
		// handle error
		return
	}
	switch valid {
	case true:
		fmt.Println("bundle is valid")
	case false:
		fmt.Println("bundle is invalid")
	}
}

// i req: bundle, The Bundle to validate.
// o: error, Returned for any failed validation.
func ExampleValidBundle() {}

// i req: txs, The transactions to group into different Bundles.
// o: Bundles, The different Bundles resulting from the group operation.
func ExampleGroupTransactionsIntoBundles() {}

// i req: bndl, The Bundle from which to get the tail transaction of.
func ExampleTailTransactionHash() {
	bndl := bundle.Bundle{
		{
			Hash:         "AAAA...",
			CurrentIndex: 0,
			// ...
		},
		{
			Hash:         "BBBB...",
			CurrentIndex: 1,
			// ...
		},
	}
	tailTxHash := bundle.TailTransactionHash(bndl)
	fmt.Println(tailTxHash) // "AAAA..."
}
