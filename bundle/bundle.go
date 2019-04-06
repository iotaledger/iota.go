// Package bundle provides primitives to create and validate bundles.
package bundle

import (
	"github.com/iotaledger/iota.go/checksum"
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl"
	"github.com/iotaledger/iota.go/signing"
	"github.com/iotaledger/iota.go/transaction"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
	"math"
	"time"
)

// Bundles are a slice of Bundle.
type Bundles []Bundle

// BundlesByTimestamp are sorted bundles by attachment timestamp.
type BundlesByTimestamp Bundles

// ignore
func (a BundlesByTimestamp) Len() int { return len(a) }

// ignore
func (a BundlesByTimestamp) Swap(i int, j int) { a[i], a[j] = a[j], a[i] }

// ignore
func (a BundlesByTimestamp) Less(i int, j int) bool {
	return a[i][0].AttachmentTimestamp < a[j][0].AttachmentTimestamp
}

// PadTag pads the given trytes up to the length of a tag.
func PadTag(tag Trytes) Trytes {
	return Pad(tag, 27)
}

// Bundle represents grouped together transactions for creating a transfer.
type Bundle = transaction.Transactions

// BundleEntry is an object which gets translated into one or multiple transactions
// when used in conjunction with AddEntry().
type BundleEntry struct {
	// The amount of transactions to fulfill this entry.
	Length uint64
	// The address used for this entry.
	Address Hash
	// The value for this entry.
	Value int64
	// The tag for this entry.
	Tag Trytes
	// The timestamp for this entry.
	Timestamp uint64
	// One or multiple signature message fragments.
	SignatureMessageFragments []Trytes
}

// BundleEntries are a slice of BundleEntry.
type BundleEntries = []BundleEntry

// Transfers are a slice of Transfer.
type Transfers []Transfer

// Transfer represents the data/value to transfer to an address.
type Transfer struct {
	Address Hash
	Value   uint64
	Message Trytes
	Tag     Trytes
}

// EmptyTransfer is a transfer with 9s initialized values.
var EmptyTransfer = Transfer{
	Message: NullSignatureMessageFragmentTrytes,
	Value:   0,
	Tag:     NullTagTrytes,
	Address: NullAddressWithChecksum,
}

// TransfersToBundleEntries translates transfers to bundle entries.
func TransfersToBundleEntries(timestamp uint64, transfers ...Transfer) (BundleEntries, error) {
	entries := BundleEntries{}
	for i := range transfers {
		transfer := &transfers[i]
		msgLength := len(transfer.Message)
		length := int(math.Ceil(float64(msgLength) / SignatureMessageFragmentSizeInTrytes))
		if length == 0 {
			length = 1
		}
		addr, err := checksum.RemoveChecksum(transfer.Address)
		if err != nil {
			return nil, err
		}

		transfer.Message = Pad(transfer.Message, length*SignatureMessageFragmentSizeInTrytes)

		bndlEntry := BundleEntry{
			Address: addr, Value: int64(transfer.Value),
			Tag: transfer.Tag, Timestamp: timestamp,
			Length: uint64(length),
			SignatureMessageFragments: func() []Trytes {
				splitFrags := make([]Trytes, int(length))
				for i := 0; i < int(length); i++ {
					splitFrags[i] = transfer.Message[i*SignatureMessageFragmentSizeInTrytes : (i+1)*SignatureMessageFragmentSizeInTrytes]
				}
				return splitFrags
			}(),
		}

		entries = append(entries, bndlEntry)
	}

	return entries, nil
}

// AddEntry adds a new entry to the bundle. It automatically adds additional transactions if the signature
// message fragments don't fit into one transaction.
func AddEntry(txs Bundle, bndlEntry BundleEntry) Bundle {
	bndlEntry = getBundleEntryWithDefaults(bndlEntry)
	prevLastIndex := uint64(len(txs))
	lastIndex := uint64(len(txs)-1) + bndlEntry.Length
	tag := PadTag(bndlEntry.Tag)

	// set new last index on existing txs
	for i := range txs {
		txs[i].LastIndex = lastIndex
	}

	var i uint64
	for ; i < bndlEntry.Length; i++ {
		var v int64
		if i == 0 {
			v = bndlEntry.Value
		}
		txs = append(txs, transaction.Transaction{
			Address: bndlEntry.Address,
			Value:   v, Tag: tag, ObsoleteTag: tag,
			CurrentIndex: prevLastIndex + i,
			LastIndex:    lastIndex, Timestamp: bndlEntry.Timestamp,
			SignatureMessageFragment: bndlEntry.SignatureMessageFragments[i],
			TrunkTransaction:         NullHashTrytes, BranchTransaction: NullHashTrytes,
			Bundle: NullHashTrytes, Nonce: NullNonceTrytes, Hash: NullHashTrytes,
		})
	}

	return txs
}

func getBundleEntryWithDefaults(entry BundleEntry) BundleEntry {
	if entry.Length == 0 {
		entry.Length = 1
	}
	if len(entry.Address) == 0 {
		entry.Address = NullHashTrytes
	}
	if len(entry.Tag) == 0 {
		entry.Tag = NullTagTrytes
	}
	if entry.Timestamp == 0 {
		entry.Timestamp = uint64(time.Now().UnixNano() / int64(time.Second))
	}

	if entry.SignatureMessageFragments == nil || len(entry.SignatureMessageFragments) == 0 {
		entry.SignatureMessageFragments = make([]Trytes, entry.Length)
		var i uint64
		for ; i < entry.Length; i++ {
			entry.SignatureMessageFragments[i] = NullSignatureMessageFragmentTrytes
		}
	} else {
		for i := range entry.SignatureMessageFragments {
			entry.SignatureMessageFragments[i] = Pad(entry.SignatureMessageFragments[i], 2187)
		}
	}

	return entry
}

// Finalize finalizes the bundle by calculating the bundle hash and setting it on each transaction
// bundle hash field.
func Finalize(bundle Bundle) (Bundle, error) {
	var valueTrits = make([]Trits, len(bundle))
	var timestampTrits = make([]Trits, len(bundle))
	var currentIndexTrits = make([]Trits, len(bundle))
	var obsoleteTagTrits = make([]Trits, len(bundle))
	var lastIndexTrits = PadTrits(IntToTrits(int64(bundle[0].LastIndex)), 27)

	for i := range bundle {
		valueTrits[i] = PadTrits(IntToTrits(bundle[i].Value), 81)
		timestampTrits[i] = PadTrits(IntToTrits(int64(bundle[i].Timestamp)), 27)
		currentIndexTrits[i] = PadTrits(IntToTrits(int64(bundle[i].CurrentIndex)), 27)
		obsoleteTagTrits[i] = PadTrits(MustTrytesToTrits(bundle[i].ObsoleteTag), 81)
	}

	var bundleHash Hash
	for {
		k := kerl.NewKerl()

		for i := 0; i < len(bundle); i++ {
			relevantTritsForBundleHash := MustTrytesToTrits(
				bundle[i].Address +
					MustTritsToTrytes(valueTrits[i]) +
					MustTritsToTrytes(obsoleteTagTrits[i]) +
					MustTritsToTrytes(timestampTrits[i]) +
					MustTritsToTrytes(currentIndexTrits[i]) +
					MustTritsToTrytes(lastIndexTrits),
			)
			k.Absorb(relevantTritsForBundleHash)
		}

		bundleHashTrits, err := k.Squeeze(HashTrinarySize)
		if err != nil {
			return nil, err
		}
		bundleHash = MustTritsToTrytes(bundleHashTrits)

		// check whether normalized bundle hash can be computed
		normalizedBundleHash := signing.NormalizedBundleHash(bundleHash)
		ok := true
		for i := range normalizedBundleHash {
			if normalizedBundleHash[i] == 13 {
				ok = false
				break
			}
		}
		if ok {
			break
		}
		obsoleteTagTrits[0] = AddTrits(obsoleteTagTrits[0], Trits{1})
	}

	// set the computed bundle hash on each tx in the bundle
	for i := range bundle {
		tx := &bundle[i]
		if i == 0 {
			tx.ObsoleteTag = MustTritsToTrytes(obsoleteTagTrits[0])
		}
		tx.Bundle = bundleHash
	}

	return bundle, nil
}

// AddTrytes adds the given fragments to the txs in the bundle starting
// from the specified offset.
func AddTrytes(bndl Bundle, fragments []Trytes, offset int) Bundle {
	for i := range bndl {
		if i >= offset && i < offset+len(fragments) {
			bndl[i].SignatureMessageFragment = Pad(fragments[i-offset], 27*81)
		}
	}
	return bndl
}

// ValidateBundleSignatures validates all signatures of the given bundle.
// Use ValidBundle() if you want to validate the overall structure of the bundle and the signatures.
func ValidateBundleSignatures(bundle Bundle) (bool, error) {
	for i := range bundle {
		tx := &bundle[i]

		// check whether input transaction
		if tx.Value >= 0 {
			continue
		}

		// it is unknown how many fragments there will be
		fragments := []Trytes{tx.SignatureMessageFragment}

		// find the subsequent txs containing the remaining signature
		// message fragments for this input transaction
		for j := i; j < len(bundle)-1; j++ {
			otherTx := &bundle[j+1]
			if otherTx.Value != 0 || otherTx.Address != tx.Address {
				continue
			}

			fragments = append(fragments, otherTx.SignatureMessageFragment)
		}

		valid, err := signing.ValidateSignatures(tx.Address, fragments, tx.Bundle)
		if err != nil {
			return false, err
		}
		if !valid {
			return false, nil
		}
	}
	return true, nil
}

// ValidBundle checks if a bundle is syntactically valid.
// Validates signatures and overall structure.
func ValidBundle(bundle Bundle) error {
	var totalSum int64

	sigs := make(map[Hash][]Trytes)
	k := kerl.NewKerl()

	lastIndex := uint64(len(bundle) - 1)
	for i := range bundle {
		tx := &bundle[i]
		totalSum += tx.Value

		if tx.CurrentIndex != uint64(i) {
			return errors.Wrapf(ErrInvalidBundle, "expected tx at index %d to have current index %d but got %d", i, i, tx.CurrentIndex)
		}
		if tx.LastIndex != lastIndex {
			return errors.Wrapf(ErrInvalidBundle, "expected tx at index %d to have last index %d but got %d", i, lastIndex, tx.LastIndex)
		}

		txTrits := MustTrytesToTrits(transaction.MustTransactionToTrytes(tx)[2187 : 2187+162])
		k.Absorb(txTrits)

		// continue if output or signature txbundle bundle
		if tx.Value >= 0 {
			continue
		}

		// here we have an input transaction (negative value)
		sigs[tx.Address] = append(sigs[tx.Address], tx.SignatureMessageFragment)

		// find the subsequent txs containing the remaining signature
		// message fragments for this input transaction
		for j := i; j < len(bundle)-1; j++ {
			tx2 := &bundle[j+1]

			// check if the tx is part of the input transaction
			if tx2.Address == tx.Address && tx2.Value == 0 {
				// append the signature message fragment
				sigs[tx.Address] = append(sigs[tx.Address], tx2.SignatureMessageFragment)
			}
		}
	}

	// sum of all transaction must be 0
	if totalSum != 0 {
		return errors.Wrapf(ErrInvalidBundle, "bundle total sum should be 0 but got %d", totalSum)
	}

	bundleHashTrits, err := k.Squeeze(HashTrinarySize)
	if err != nil {
		return err
	}

	bundleHash := MustTritsToTrytes(bundleHashTrits)

	if bundleHash != bundle[0].Bundle {
		return ErrInvalidBundleHash
	}

	// validate the signatures
	valid, err := ValidateBundleSignatures(bundle)
	if err != nil {
		return err
	}

	if !valid {
		return ErrInvalidSignature
	}

	return nil
}

// GroupTransactionsIntoBundles groups the given transactions into groups of bundles.
// Note that the same bundle can exist in the return slice multiple times, though they
// are reattachments of the same transfer.
func GroupTransactionsIntoBundles(txs transaction.Transactions) Bundles {
	bundles := Bundles{}

	for i := range txs {
		tx := &txs[i]
		if tx.CurrentIndex != 0 {
			continue
		}

		bundle := Bundle{*tx}
		lastIndex := int(tx.LastIndex)
		current := tx
		for x := 1; x <= lastIndex; x++ {
			// get all txs belonging into this bundle
			found := false
			for j := range txs {
				if current.Bundle != txs[j].Bundle ||
					txs[j].CurrentIndex != current.CurrentIndex+1 ||
					current.TrunkTransaction != txs[j].Hash {
					continue
				}
				found = true
				bundle = append(bundle, txs[j])
				current = &txs[j]
				break
			}
			if !found {
				break
			}
		}
		bundles = append(bundles, bundle)
	}

	return bundles
}

// TailTransactionHash returns the tail transaction's hash.
func TailTransactionHash(bndl Bundle) Hash {
	if bndl == nil || len(bndl) == 0 {
		return ""
	}
	for i := range bndl {
		tx := &bndl[i]
		if tx.CurrentIndex != 0 {
			continue
		}
		if len(tx.Hash) > 0 {
			return tx.Hash
		}
		return transaction.TransactionHash(tx)
	}
	return ""
}
