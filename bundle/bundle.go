package bundle

import (
	"errors"
	"github.com/iotaledger/giota/curl"
	"github.com/iotaledger/giota/kerl"
	"github.com/iotaledger/giota/signing"
	"github.com/iotaledger/giota/transaction"
	"github.com/iotaledger/giota/trinary"
	"github.com/iotaledger/giota/utils"
	"time"
)

var (
	ErrInvalidCurrentIndex  = errors.New("invalid current index")
	ErrInvalidLastIndex     = errors.New("invalid last index")
	ErrInvalidSignature     = errors.New("invalid signature")
	ErrInvalidBundleBalance = errors.New("summed up values of all txs in the bundle must be 0")
	ErrNonFinalizedBundle   = errors.New("bundle wasn't finalized")
)

type Bundles []Bundle

// BundlesByTimestamp are sorted bundles by attachment timestamp
type BundlesByTimestamp Bundles

func (a BundlesByTimestamp) Len() int {
	return len(a)
}
func (a BundlesByTimestamp) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a BundlesByTimestamp) Less(i, j int) bool {
	return a[i][0].AttachmentTimestamp.Trits().Value() < a[j][0].AttachmentTimestamp.Trits().Value()
}

// Bundle represents grouped together transactions for creating a transfer.
type Bundle []transaction.Transaction

// AddEntry adds a new transaction entry to the bundle. By the given num. fragments it adds
// one ore more transaction to accustom the resulting signature message fragments.
// Transaction properties not specified as parameters to this function are initialized with empty hash values.
func (bundle *Bundle) AddEntry(numFragments int, address signing.Address, value int64, timestamp time.Time, tag trinary.Trytes) {
	if tag == "" {
		tag = curl.EmptyHash[:27]
	}

	for i := 0; i < int(numFragments); i++ {
		var v int64

		if i == 0 {
			v = value
		}

		b := transaction.Transaction{
			SignatureMessageFragment:      signing.EmptySig,
			Address:                       address,
			Value:                         v,
			ObsoleteTag:                   trinary.Pad(tag, transaction.TagTrinarySize/3),
			Timestamp:                     timestamp,
			CurrentIndex:                  int64(len(*bundle) - 1),
			LastIndex:                     0,
			Bundle:                        curl.EmptyHash,
			TrunkTransaction:              curl.EmptyHash,
			BranchTransaction:             curl.EmptyHash,
			Tag:                           trinary.Pad(tag, transaction.TagTrinarySize/3),
			AttachmentTimestamp:           curl.EmptyHash,
			AttachmentTimestampLowerBound: curl.EmptyHash,
			AttachmentTimestampUpperBound: curl.EmptyHash,
			Nonce:                         curl.EmptyHash,
		}
		*bundle = append(*bundle, b)
	}
}

// Finalize adds the given signature message fragments to the transactions
// and initializes the indices and bundle hash properties.
func (bundle Bundle) Finalize(sig []trinary.Trytes) error {
	h, err := bundle.NormalizedHash()
	if err != nil {
		return err
	}

	lastIndex := int64(len(bundle) - 1)
	for i := range bundle {
		if i < len(sig) && sig[i] != "" {
			bundle[i].SignatureMessageFragment = trinary.Pad(sig[i], transaction.SignatureMessageFragmentTrinarySize/3)
		}

		bundle[i].CurrentIndex = int64(i)
		bundle[i].LastIndex = lastIndex
		bundle[i].Bundle = h
	}
	return nil
}

// Hash calculates the non normalized hash of the bundle.
func (bundle Bundle) Hash() (trinary.Trytes, error) {
	k := kerl.NewKerl()
	buf := make(trinary.Trits, 243+81*3)

	for i, b := range bundle {
		copyRelevantTritsForBundleHash(buf, &b, i, len(bundle))
		k.Absorb(buf)
	}

	h, err := k.Squeeze(curl.HashSize)
	return h.MustTrytes(), err
}

// NormalizedHash calculates a normalized hash of the bundle.
// The obsolete tag is incremented as many times as needed
// in order to prevent M/13 tryte values in the resulting bundle hash.
func (bundle Bundle) NormalizedHash() (trinary.Trytes, error) {
	k := kerl.NewKerl()
	hashedLen := transaction.BundleTrinaryOffset - transaction.AddressTrinaryOffset

	// copy all relevant trits from the transactions in the bundle into the buffer
	buf := make(trinary.Trits, hashedLen*len(bundle))
	for i, b := range bundle {
		copyRelevantTritsForBundleHash(buf[i*hashedLen:], &b, i, len(bundle))
	}

	for {
		k.Absorb(buf)
		hashTrits, err := k.Squeeze(curl.HashSize)
		if err != nil {
			return "", err
		}
		h := hashTrits.MustTrytes()
		n := h.Normalize()
		valid := true

		for _, v := range n {
			if v == 13 {
				valid = false
				break
			}
		}

		offset := transaction.ObsoleteTagTrinaryOffset - transaction.AddressTrinaryOffset

		if valid {
			bundle[0].ObsoleteTag = buf[offset : offset+transaction.ObsoleteTagTrinarySize].MustTrytes()
			return h, nil
		}

		k.Reset()
		trinary.IncTrits(buf[offset : offset+transaction.ObsoleteTagTrinarySize])
	}
}

// copies the relevant trits for the bundle hash calculation from the
// the given transaction into the given buffer. Following properties are used:
// address, value, obsolete tag, timestamp, current index, last index
func copyRelevantTritsForBundleHash(buf trinary.Trits, b *transaction.Transaction, i, l int) {
	copy(buf, trinary.Trytes(b.Address).Trits())
	copy(buf[243:], trinary.IntToTrits(b.Value, transaction.ValueTrinarySize))
	copy(buf[243+81:], b.ObsoleteTag.Trits())
	copy(buf[243+81+81:], trinary.IntToTrits(b.Timestamp.Unix(), transaction.TimestampTrinarySize))
	copy(buf[243+81+81+27:], trinary.IntToTrits(int64(i), transaction.CurrentIndexTrinarySize))
	copy(buf[243+81+81+27+27:], trinary.IntToTrits(int64(l-1), transaction.LastIndexTrinarySize))
}

// Categorize categorizes a list of transfers into sent and received. It is important to
// note that zero value transfers (which for example, are being used for storing
// addresses in the Tangle), are seen as received in this function.
func (bundle Bundle) Categorize(adr signing.Address) (send Bundle, received Bundle) {
	send = make(Bundle, 0, len(bundle))
	received = make(Bundle, 0, len(bundle))

	for _, b := range bundle {
		switch {
		case b.Address != adr:
			continue
		case b.Value >= 0:
			received = append(received, b)
		default:
			send = append(send, b)
		}
	}
	return
}

// IsValid checks the validity of bundle. It checks whether the sum
// of all transactions in the bundle results to 0 (in+out txs must = 0) and whether all
// signatures are valid. Before calling this function, Finalize() must be called to
// correctly initialize the signature message fragments, indices and the bundle hash.
func (bundle Bundle) IsValid() error {
	var total int64

	sigs := make(map[signing.Address][]trinary.Trytes)

	for index, b := range bundle {
		total += b.Value

		switch {
		case b.CurrentIndex != int64(index):
			return ErrInvalidCurrentIndex
		case b.LastIndex != int64(len(bundle)-1):
			return ErrInvalidLastIndex
			// continue if output or signature tx
		case b.Value >= 0:
			continue
		}

		// check whether the signature message fragment isn't empty
		if utils.IsEmptyTrytes(b.SignatureMessageFragment) {
			return ErrNonFinalizedBundle
		}

		// here we have an input transaction (negative value)
		sigs[b.Address] = append(sigs[b.Address], b.SignatureMessageFragment)

		// find the subsequent txs containing the remaining signature
		// message fragments for this input transaction
		for i := index; i < len(bundle)-1; i++ {
			tx := &bundle[i+1]

			// check if the tx is part of the input transaction
			if tx.Address == b.Address && tx.Value == 0 {
				// append the signature message fragment
				sigs[tx.Address] = append(sigs[tx.Address], tx.SignatureMessageFragment)
			}
		}
	}

	// sum of all transaction must be 0
	if total != 0 {
		return ErrInvalidBundleBalance
	}

	// validate the signatures
	hash, err := bundle.Hash()
	if err != nil {
		return err
	}

	for addr, sig := range sigs {
		if !signing.IsValidSig(addr, sig, hash) {
			return ErrInvalidSignature
		}
	}

	return nil
}

// SignInputs signs the input transactions (txs with negative value) and their additional
// signature fragment holding txs (given the security level)
func (bundle Bundle) SignInputs(inputs []AddressInput) error {
	// compute normalized bundle hash
	hash, err := bundle.Hash()
	if err != nil {
		return err
	}
	normalizedBundleHash := hash.Normalize()

	// input signing:
	// find all input transactions (txs with negative value), get the corresponding private key
	// and compute the signature fragment
	for i, _ := range bundle {
		if bundle[i].Value >= 0 {
			continue
		}

		//  get the corresponding key index and security level of the address
		var ai AddressInput
		for _, in := range inputs {
			addr, err := in.Address()
			if err != nil {
				return err
			}

			if addr == bundle[i].Address {
				ai = in
				break
			}
		}

		// get the corresponding private key of the address
		key, err := ai.Key()
		if err != nil {
			return err
		}

		// calculate the new signature fragment with the first bundle fragment
		bundle[i].SignatureMessageFragment = signing.Sign(normalizedBundleHash[:27], key[:6561/3])

		// if user chooses higher than 27-trytes security
		// for each security level, add an additional signature
		for j := 1; j < int(ai.Security); j++ {
			// since the signature is > 2187 trytes, we need to find the subsequent
			// txs with the same address (and value = 0) to add the remainder of the signature fragment
			if bundle[i+j].Address != bundle[i].Address || bundle[i+j].Value != 0 {
				continue
			}
			// calculate the signature fragment
			nfrag := signing.Sign(normalizedBundleHash[(j%3)*27:(j%3)*27+27], key[6561*j/3:(j+1)*6561/3])
			// convert signature to trytes and assign it again to this bundle entry
			bundle[i+j].SignatureMessageFragment = nfrag
		}
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
					current.TrunkTransaction != txs[j].Hash() {
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