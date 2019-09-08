// Package transaction provides functions for parsing transactions, extracting JSON data from them,
// conversions and validation.
package transaction

import (
	. "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/converter"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/guards"
	. "github.com/iotaledger/iota.go/trinary"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
	"strings"
)

// Transactions is a slice of Transaction.
type Transactions []Transaction

// Transaction represents a single transaction.
type Transaction struct {
	Hash                          Hash   `json:"hash"`
	SignatureMessageFragment      Trytes `json:"signatureMessageFragment"`
	Address                       Hash   `json:"address"`
	Value                         int64  `json:"value"`
	ObsoleteTag                   Trytes `json:"obsoleteTag"`
	Timestamp                     uint64 `json:"timestamp"`
	CurrentIndex                  uint64 `json:"currentIndex"`
	LastIndex                     uint64 `json:"lastIndex"`
	Bundle                        Hash   `json:"bundle"`
	TrunkTransaction              Hash   `json:"trunkTransaction"`
	BranchTransaction             Hash   `json:"branchTransaction"`
	Tag                           Trytes `json:"tag"`
	AttachmentTimestamp           int64  `json:"attachmentTimestamp"`
	AttachmentTimestampLowerBound int64  `json:"attachmentTimestampLowerBound"`
	AttachmentTimestampUpperBound int64  `json:"attachmentTimestampUpperBound"`
	Nonce                         Trytes `json:"nonce"`
	Confirmed                     *bool  `json:"confirmed,omitempty"`
	Persistence                   *bool  `json:"persistence,omitempty"`
}

// ParseTransaction parses the trits and returns a transaction object.
// The trits slice must be TransactionTrinarySize in length.
// If noHash is set to true, no transaction hash is calculated.
func ParseTransaction(trits Trits, noHash ...bool) (*Transaction, error) {
	var err error

	if len(trits) != TransactionTrinarySize {
		return nil, ErrInvalidTransaction
	}

	if err := ValidTrits(trits); err != nil {
		return nil, err
	}

	t := &Transaction{}
	t.SignatureMessageFragment = MustTritsToTrytes(trits[SignatureMessageFragmentTrinaryOffset:SignatureMessageFragmentTrinarySize])
	t.Address, err = TritsToTrytes(trits[AddressTrinaryOffset : AddressTrinaryOffset+AddressTrinarySize])
	if err != nil {
		return nil, err
	}
	t.Value = TritsToInt(trits[ValueOffsetTrinary : ValueOffsetTrinary+ValueSizeTrinary])
	t.ObsoleteTag = MustTritsToTrytes(trits[ObsoleteTagTrinaryOffset : ObsoleteTagTrinaryOffset+ObsoleteTagTrinarySize])
	t.Timestamp = uint64(TritsToInt(trits[TimestampTrinaryOffset : TimestampTrinaryOffset+TimestampTrinarySize]))
	t.CurrentIndex = uint64(TritsToInt(trits[CurrentIndexTrinaryOffset : CurrentIndexTrinaryOffset+CurrentIndexTrinarySize]))
	t.LastIndex = uint64(TritsToInt(trits[LastIndexTrinaryOffset : LastIndexTrinaryOffset+LastIndexTrinarySize]))
	if t.CurrentIndex > t.LastIndex {
		return nil, errors.Wrap(ErrInvalidIndex, "current index is bigger than last index")
	}
	t.Bundle = MustTritsToTrytes(trits[BundleTrinaryOffset : BundleTrinaryOffset+BundleTrinarySize])
	t.TrunkTransaction = MustTritsToTrytes(trits[TrunkTransactionTrinaryOffset : TrunkTransactionTrinaryOffset+TrunkTransactionTrinarySize])
	t.BranchTransaction = MustTritsToTrytes(trits[BranchTransactionTrinaryOffset : BranchTransactionTrinaryOffset+BranchTransactionTrinarySize])
	t.Tag = MustTritsToTrytes(trits[TagTrinaryOffset : TagTrinaryOffset+TagTrinarySize])
	t.AttachmentTimestamp = TritsToInt(trits[AttachmentTimestampTrinaryOffset : AttachmentTimestampTrinaryOffset+AttachmentTimestampTrinarySize])
	t.AttachmentTimestampLowerBound = TritsToInt(trits[AttachmentTimestampLowerBoundTrinaryOffset : AttachmentTimestampLowerBoundTrinaryOffset+AttachmentTimestampLowerBoundTrinarySize])
	t.AttachmentTimestampUpperBound = TritsToInt(trits[AttachmentTimestampUpperBoundTrinaryOffset : AttachmentTimestampUpperBoundTrinaryOffset+AttachmentTimestampUpperBoundTrinarySize])
	t.Nonce = MustTritsToTrytes(trits[NonceTrinaryOffset : NonceTrinaryOffset+NonceTrinarySize])
	if len(noHash) == 0 || noHash[0] == false {
		t.Hash = TransactionHash(t)
	}

	return t, nil
}

// ValidTransactionTrytes checks whether the given trytes make up a valid transaction schematically.
func ValidTransactionTrytes(trytes Trytes) error {
	// verifies length and trytes values
	if !guards.IsTrytesOfExactLength(trytes, TransactionTrytesSize) {
		return ErrInvalidTrytes
	}

	if trytes[2279:2295] != "9999999999999999" {
		return ErrInvalidTrytes
	}

	return nil
}

// AsTransactionObject makes a new transaction from the given trytes.
// Optionally the computed transaction hash can be overwritten by supplying an own hash.
func AsTransactionObject(trytes Trytes, hash ...Hash) (*Transaction, error) {
	var tx *Transaction
	var err error

	if err := ValidTransactionTrytes(trytes); err != nil {
		return nil, err
	}

	skipHashCalc := len(hash) > 0
	if tx, err = ParseTransaction(MustTrytesToTrits(trytes), skipHashCalc); err != nil {
		return nil, err
	}

	if skipHashCalc {
		tx.Hash = hash[0]
	}

	return tx, nil
}

// AsTransactionObjects constructs new transactions from the given raw trytes.
func AsTransactionObjects(rawTrytes []Trytes, hashes Hashes) (Transactions, error) {
	txs := make(Transactions, len(rawTrytes))
	var tx *Transaction
	var err error
	for i := range rawTrytes {
		if hashes != nil && len(hashes) > 0 && len(hashes) > i {
			tx, err = AsTransactionObject(rawTrytes[i], hashes[i])
		} else {
			tx, err = AsTransactionObject(rawTrytes[i])
		}
		if err != nil {
			return nil, err
		}
		txs[i] = *tx
	}
	return txs, nil
}

// TransactionToTrits converts the transaction to trits.
func TransactionToTrits(t *Transaction) (Trits, error) {
	tr := make(Trits, TransactionTrinarySize)
	if !guards.IsTrytesOfExactLength(t.SignatureMessageFragment, SignatureMessageFragmentTrinarySize/3) {
		return nil, errors.Wrap(ErrInvalidTrytes, "invalid signature message fragment")
	}
	copy(tr, MustTrytesToTrits(t.SignatureMessageFragment))

	if !guards.IsTrytesOfExactLength(t.Address, AddressTrinarySize/3) {
		return nil, errors.Wrap(ErrInvalidTrytes, "invalid address")
	}
	copy(tr[AddressTrinaryOffset:], MustTrytesToTrits(t.Address))

	copy(tr[ValueOffsetTrinary:], IntToTrits(t.Value))
	if !guards.IsTrytesOfExactLength(t.ObsoleteTag, ObsoleteTagTrinarySize/3) {
		return nil, errors.Wrap(ErrInvalidTrytes, "invalid obsolete tag")
	}
	copy(tr[ObsoleteTagTrinaryOffset:], MustTrytesToTrits(t.ObsoleteTag))

	copy(tr[TimestampTrinaryOffset:], IntToTrits(int64(t.Timestamp)))
	if t.CurrentIndex > t.LastIndex {
		return nil, errors.Wrap(ErrInvalidIndex, "current index is bigger than last index")
	}

	copy(tr[CurrentIndexTrinaryOffset:], IntToTrits(int64(t.CurrentIndex)))
	copy(tr[LastIndexTrinaryOffset:], IntToTrits(int64(t.LastIndex)))
	if !guards.IsTrytesOfExactLength(t.Bundle, BundleTrinarySize/3) {
		return nil, errors.Wrap(ErrInvalidTrytes, "invalid bundle hash")
	}
	copy(tr[BundleTrinaryOffset:], MustTrytesToTrits(t.Bundle))

	if !guards.IsTrytesOfExactLength(t.TrunkTransaction, TrunkTransactionTrinarySize/3) {
		return nil, errors.Wrap(ErrInvalidTrytes, "invalid trunk tx hash")
	}
	copy(tr[TrunkTransactionTrinaryOffset:], MustTrytesToTrits(t.TrunkTransaction))

	if !guards.IsTrytesOfExactLength(t.BranchTransaction, BranchTransactionTrinarySize/3) {
		return nil, errors.Wrap(ErrInvalidTrytes, "invalid branch tx hash")
	}
	copy(tr[BranchTransactionTrinaryOffset:], MustTrytesToTrits(t.BranchTransaction))

	if !guards.IsTrytesOfExactLength(t.Tag, TagTrinarySize/3) {
		return nil, errors.Wrap(ErrInvalidTrytes, "invalid tag")
	}
	copy(tr[TagTrinaryOffset:], MustTrytesToTrits(t.Tag))
	copy(tr[AttachmentTimestampTrinaryOffset:], IntToTrits(t.AttachmentTimestamp))
	copy(tr[AttachmentTimestampLowerBoundTrinaryOffset:], IntToTrits(t.AttachmentTimestampLowerBound))
	copy(tr[AttachmentTimestampUpperBoundTrinaryOffset:], IntToTrits(t.AttachmentTimestampUpperBound))
	if !guards.IsTrytesOfExactLength(t.Nonce, NonceTrinarySize/3) {
		return nil, errors.Wrap(ErrInvalidTrytes, "invalid nonce")
	}
	copy(tr[NonceTrinaryOffset:], MustTrytesToTrits(t.Nonce))

	return tr, nil
}

// TransactionToTrytes converts the transaction to trytes.
func TransactionToTrytes(t *Transaction) (Trytes, error) {
	tr, err := TransactionToTrits(t)
	if err != nil {
		return "", err
	}
	return MustTritsToTrytes(tr), nil
}

// MustTransactionToTrytes converts the transaction to trytes.
func MustTransactionToTrytes(t *Transaction) Trytes {
	trytes, err := TransactionToTrytes(t)
	if err != nil {
		panic(err)
	}
	return trytes
}

// TransactionsToTrytes returns a slice of transaction trytes from the given transactions.
func TransactionsToTrytes(txs Transactions) ([]Trytes, error) {
	trytes := make([]Trytes, len(txs))
	var err error
	for i := range txs {
		trytes[i], err = TransactionToTrytes(&txs[i])
		if err != nil {
			return nil, err
		}
	}
	return trytes, nil
}

// MustTransactionsToTrytes returns a slice of transaction trytes from the given transactions.
func MustTransactionsToTrytes(txs Transactions) []Trytes {
	trytes := make([]Trytes, len(txs))
	for i := range txs {
		trytes[i] = MustTransactionToTrytes(&txs[i])
	}
	return trytes
}

// FinalTransactionTrytes returns a slice of transaction trytes from the given transactions.
// The order of the transactions is reversed in the output slice.
func FinalTransactionTrytes(txs Transactions) ([]Trytes, error) {
	trytes, err := TransactionsToTrytes(txs)
	if err != nil {
		return nil, err
	}
	for i, j := 0, len(trytes)-1; i < j; i, j = i+1, j-1 {
		trytes[i], trytes[j] = trytes[j], trytes[i]
	}
	return trytes, nil
}

// MustFinalTransactionTrytes returns a slice of transaction trytes from the given transactions.
// The order of the transactions is reversed in the output slice.
func MustFinalTransactionTrytes(txs Transactions) []Trytes {
	trytes := MustTransactionsToTrytes(txs)
	for i, j := 0, len(trytes)-1; i < j; i, j = i+1, j-1 {
		trytes[i], trytes[j] = trytes[j], trytes[i]
	}
	return trytes
}

// TransactionHash makes a transaction hash from the given transaction.
func TransactionHash(t *Transaction) Hash {
	return curl.MustHashTrytes(MustTransactionToTrytes(t))
}

// HasValidNonce checks if the transaction has the valid MinWeightMagnitude.
// MWM corresponds to the amount of zero trits at the end of the transaction hash.
func HasValidNonce(t *Transaction, mwm uint64) bool {
	if len(t.Hash) == 0 {
		t.Hash = TransactionHash(t)
	}
	return TrailingZeros(MustTrytesToTrits(t.Hash)) >= int64(mwm)
}

// IsTailTransaction checks if given transaction object is tail transaction.
// A tail transaction is one with currentIndex = 0.
func IsTailTransaction(t *Transaction) bool {
	return t.CurrentIndex == 0
}

var numericTrytesRegex = regexp.MustCompile(`^(RA|PA)?(UA|VA|WA|XA|YA|ZA|9B|AB|BB|CB)+((SA)(UA|VA|WA|XA|YA|ZA|9B|AB|BB|CB)+)?((TC|OB)(RA|PA)?(UA|VA|WA|XA|YA|ZA|9B|AB|BB|CB)+)?99`)
var numPadRegex = regexp.MustCompile(`^(.*)99`)

const boolFalseJSONTrytes = "UCPC9DGDTC"
const boolTrueJSONTrytes = "HDFDIDTC"
const nullJSONTrytes = "BDID9D9D"

// ExtractJSON extracts a JSON string from the given transactions.
// It supports JSON messages in the following format:
//
// - "{ \"message\": \"hello\" }"
//
// - "[1, 2, 3]"
//
// - "true", "false" and "null"
//
// - "hello"
//
// - 123
func ExtractJSON(txs Transactions) (string, error) {
	if txs == nil || len(txs) == 0 {
		return "", ErrInvalidBundle
	}

	switch {
	case txs[0].SignatureMessageFragment[:10] == boolFalseJSONTrytes:
		return "false", nil
	case txs[0].SignatureMessageFragment[:8] == boolTrueJSONTrytes:
		return "true", nil
	case txs[0].SignatureMessageFragment[:8] == nullJSONTrytes:
		return "null", nil
	}

	if numericTrytesRegex.MatchString(txs[0].SignatureMessageFragment) {
		num := txs[0].SignatureMessageFragment[:SignatureMessageFragmentSizeInTrytes-3]
		n, err := converter.TrytesToASCII(string(num))
		if err != nil {
			return "", err
		}
		n = strings.Replace(n, "\x00", "", -1)
		f, err := strconv.ParseFloat(n, 64)
		if err != nil {
			return "", errors.Wrap(err, "can't parse number")
		}
		return strconv.FormatFloat(f, 'f', -1, 64), nil
	}

	firstTrytePair := string(txs[0].SignatureMessageFragment[0]) + string(txs[0].SignatureMessageFragment[1])
	var lastTrytePair string

	switch {
	case firstTrytePair == "OD":
		lastTrytePair = "QD"
	case firstTrytePair == "GA":
		lastTrytePair = "GA"
	case firstTrytePair == "JC":
		lastTrytePair = "LC"
	default:
		return "", ErrInvalidTryteEncodedJSON
	}

	index := 0
	notEnded := true
	trytesChunk := ""
	trytesChecked := 0
	preliminaryStop := false
	finalJSON := ""

	for index < len(txs) && notEnded {
		messageChunk := txs[index].SignatureMessageFragment

		// iterate over message chunk 9 trytes at a time
		for i := 0; i < len(messageChunk); i += 9 {
			trytes := messageChunk[i : i+9]
			trytesChunk += trytes

			// get upper limit of trytes that need to be checked
			upperLimit := len(trytesChunk) - len(trytesChunk)%2
			trytesToCheck := trytesChunk[trytesChecked:upperLimit]

			// read 2 trytes at a time and check if it equals the closing bracket character
			for j := 0; j < len(trytesToCheck); j += 2 {
				trytePair := string(trytesToCheck[j]) + string(trytesToCheck[j+1])

				// if closing bracket was found and there are only trailing 9's
				// we quit and remove the 9's from trytesChunk
				if preliminaryStop && trytePair == "99" {
					notEnded = false
					break
				}

				data, err := converter.TrytesToASCII(trytePair)
				if err != nil {
					return "", err
				}
				finalJSON += data

				// set preliminary stop if close bracket was found
				if trytePair == lastTrytePair {
					preliminaryStop = true
				}
			}

			if !notEnded {
				break
			}

			trytesChecked += len(trytesToCheck)
		}

		// use the next tx in the bundle
		index++
	}

	if notEnded {
		return "", ErrInvalidTryteEncodedJSON
	}

	return finalJSON, nil
}
