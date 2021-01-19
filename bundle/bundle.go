// Package bundle provides primitives to create and validate bundles.
package bundle

import (
	"math"
	"strings"
	"time"

	"github.com/iotaledger/iota.go/address"
	"github.com/iotaledger/iota.go/checksum"
	"github.com/iotaledger/iota.go/consts"
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/kerl"
	iotaGoMath "github.com/iotaledger/iota.go/math"
	"github.com/iotaledger/iota.go/signing"
	"github.com/iotaledger/iota.go/transaction"
	"github.com/iotaledger/iota.go/trinary"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
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
	return MustPad(tag, 27)
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

		transfer.Message = MustPad(transfer.Message, length*SignatureMessageFragmentSizeInTrytes)

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
			entry.SignatureMessageFragments[i] = MustPad(entry.SignatureMessageFragments[i], 2187)
		}
	}

	return entry
}

// Finalize finalizes the bundle by calculating and setting the bundle hash.
// It increments ObsoleteTag until the hash does not contain an M; this serves as a work-around for the so-called M-Bug.
func Finalize(bundle Bundle) (Bundle, error) {
	// bundle hash must not contain an M
	return finalize(bundle, normalizedHashHasNoM)
}

// FinalizeInsecure finalizes the bundle by calculating and setting the bundle hash.
// The bundle finalized with this could be susceptible to the so-called M-Bug, but it is much faster than Finalize.
// It must only be used for non-value transactions or for value transactions where the private key
// has been derived in an M-Bug resistant way.
func FinalizeInsecure(bundle Bundle) (Bundle, error) {
	// any bundle hash is valid
	return finalize(bundle, func(Hash) bool { return true })
}

func finalize(bundle Bundle, validator func(Hash) bool) (Bundle, error) {
	if len(bundle) == 0 {
		return bundle, nil
	}

	var (
		addresses         = make([]Trytes, len(bundle))
		values            = make([]Trytes, len(bundle))
		obsoleteTagsTrits = make([]Trits, len(bundle))
		timestamps        = make([]Trytes, len(bundle))
		currentIndexes    = make([]Trytes, len(bundle))
		lastIndexes       = make([]Trytes, len(bundle))
	)
	for i := range bundle {
		// make sure the last address trit is zero for backward compatibility
		addresses[i] = zeroLastTrit(bundle[i].Address)
		values[i] = IntToTrytes(bundle[i].Value, ValueSizeTrinary/TritsPerTryte)
		obsoleteTagsTrits[i] = MustPadTrits(MustTrytesToTrits(bundle[i].ObsoleteTag), TagTrinarySize)
		timestamps[i] = IntToTrytes(int64(bundle[i].Timestamp), TimestampTrinarySize/TritsPerTryte)
		currentIndexes[i] = IntToTrytes(int64(bundle[i].CurrentIndex), CurrentIndexTrinarySize/TritsPerTryte)
		lastIndexes[i] = IntToTrytes(int64(bundle[i].LastIndex), LastIndexTrinarySize/TritsPerTryte)
	}

	var bundleHash Hash
	for {
		k := kerl.NewKerl()
		for i := range bundle {
			var essence strings.Builder
			essence.Grow(2 * HashTrytesSize)

			essence.WriteString(addresses[i])
			essence.WriteString(values[i])
			essence.WriteString(MustTritsToTrytes(obsoleteTagsTrits[i]))
			essence.WriteString(timestamps[i])
			essence.WriteString(currentIndexes[i])
			essence.WriteString(lastIndexes[i])

			if err := k.AbsorbTrytes(essence.String()); err != nil {
				return nil, err
			}
		}

		bundleHash = k.MustSqueezeTrytes(HashTrinarySize)

		// check whether normalized bundle hash is valid
		if validator(bundleHash) {
			break
		}
		obsoleteTagsTrits[0] = AddTrits(obsoleteTagsTrits[0], Trits{1})
	}

	// update the ObsoleteTag
	bundle[0].ObsoleteTag = MustTritsToTrytes(obsoleteTagsTrits[0])

	// set the computed bundle hash on each tx in the bundle
	for i := range bundle {
		bundle[i].Bundle = bundleHash
	}

	return bundle, nil
}

func zeroLastTrit(hash Hash) Hash {
	lastTrits := MustTrytesToTrits(string(hash[HashTrytesSize-1]))
	if lastTrits[TritsPerTryte-1] == 0 {
		return hash
	}
	lastTrits[TritsPerTryte-1] = 0
	return hash[:HashTrytesSize-1] + MustTritsToTrytes(lastTrits)
}

func normalizedHashHasNoM(hash Hash) bool {
	normalizedHash := signing.NormalizedBundleHash(hash)
	for i := range normalizedHash {
		if normalizedHash[i] == MaxTryteValue {
			return false
		}
	}
	return true
}

// AddTrytes adds the given fragments to the txs in the bundle starting
// from the specified offset.
func AddTrytes(bndl Bundle, fragments []Trytes, offset int) Bundle {
	for i := range bndl {
		if i >= offset && i < offset+len(fragments) {
			bndl[i].SignatureMessageFragment = MustPad(fragments[i-offset], 27*81)
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

const (
	// The minimum amount a migration bundle has to deposit.
	MigrationBundleMinDeposit = 1_000_000
)

// ValidBundle checks if a bundle is syntactically valid.
// Validates signatures and overall structure.
func ValidBundle(bundle Bundle, applyMigrationChecks ...bool) error {
	var totalSum int64
	var migrationChecks, isTransferringValue bool
	var prevInputAddr trinary.Hash
	if len(applyMigrationChecks) > 0 && applyMigrationChecks[0] {
		migrationChecks = true
	}

	changes := map[trinary.Trytes]int64{}
	k := kerl.NewKerl()

	lastIndex := uint64(len(bundle) - 1)
	for i := range bundle {
		tx := &bundle[i]

		if iotaGoMath.AbsInt64(tx.Value) > consts.TotalSupply {
			return errors.Wrapf(ErrInvalidValue, "tx value (%d) overflows/underflows total supply", tx.Value)
		}

		if tx.Value != 0 {
			isTransferringValue = true
		}

		if migrationChecks {
			switch {
			case i == 0 || tx.Value > 0:
				// migration output transaction must be the tail transaction.
				// this also enforces that max. 1 output transaction can exist
				if tx.CurrentIndex != 0 {
					return errors.Wrapf(ErrInvalidMigrationBundle, "tail transaction must be the output transaction")
				}

				// migration bundle must deposit at least 1'000'000	tokens
				if tx.Value < MigrationBundleMinDeposit {
					return errors.Wrapf(ErrInvalidMigrationBundle, "must deposit at least %d tokens via the output transaction", MigrationBundleMinDeposit)
				}

				// must deposit to a migration address
				if err := address.IsMigrationAddress(tx.Address); err != nil {
					return err
				}
			case tx.Value < 0:
				// the input address can not be a migration address itself (this prevents non-state mutating value bundles)
				if err := address.IsMigrationAddress(tx.Address); err == nil {
					return errors.Wrapf(ErrInvalidMigrationBundle, "input transactions must not have migration addresses")
				}
				// move to next input address
				prevInputAddr = tx.Address
			default:
				// zero-value transactions, must use the previous input transaction address, this enforces
				// that the migration bundle does not contain zero-value transactions which do not hold signature fragments.
				// in case they use the same address but don't actually hold the correct fragments, this bundle will be
				// deemed invalid up on signature validation
				if tx.Address != prevInputAddr {
					return errors.Wrapf(ErrInvalidMigrationBundle, "zero-value transactions must hold signature fragments")
				}
			}
		}

		totalSum += tx.Value
		if iotaGoMath.AbsInt64(totalSum) > consts.TotalSupply {
			return errors.Wrapf(ErrInvalidBundleTotalValue, "total sum of balance mutations (%d) overflows/underflows total supply", totalSum)
		}

		changes[tx.Address] += tx.Value
		if iotaGoMath.AbsInt64(changes[tx.Address]) > consts.TotalSupply {
			return errors.Wrapf(ErrInvalidBundleAddressValue, "balance mutation (%d) on address %v overflows/underflows total supply", changes[tx.Address], tx.Address)
		}

		if tx.CurrentIndex != uint64(i) {
			return errors.Wrapf(ErrInvalidBundle, "expected tx at index %d to have current index %d but got %d", i, i, tx.CurrentIndex)
		}
		if tx.LastIndex != lastIndex {
			return errors.Wrapf(ErrInvalidBundle, "expected tx at index %d to have last index %d but got %d", i, lastIndex, tx.LastIndex)
		}

		// absorb the bundle essence of this transaction
		txTrits, err := transaction.TransactionToTrits(tx)
		if err != nil {
			return err
		}
		essenceTrits := txTrits[consts.AddressTrinaryOffset:consts.BundleTrinaryOffset]
		// set the lest address trit to zero for backward compatibility
		// this can lead to transactions with different addresses having the same bundle hash
		essenceTrits[consts.HashTrinarySize-1] = 0
		if err := k.Absorb(essenceTrits); err != nil {
			return err
		}
	}

	// sum of all transaction must be 0
	if totalSum != 0 {
		return errors.Wrapf(ErrInvalidBundle, "bundle total sum should be 0 but got %d", totalSum)
	}

	if migrationChecks && !isTransferringValue {
		return errors.Wrapf(ErrInvalidMigrationBundle, "migration bundles must transfer value")
	}

	bundleHash, err := k.SqueezeTrytes(HashTrinarySize)
	if err != nil {
		return err
	}

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
