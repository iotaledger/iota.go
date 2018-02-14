/*
MIT License

Copyright (c) 2017 Shinya Yagyu

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package giota

import (
	"errors"
	"math"
	"time"
)

// Number of random walks to perform. Currently IRI limits it to 5 to 27
const NumberOfWalks = 5

// GetUsedAddress generates a new address which is not found in the tangle
// and returns its new address and used addresses.
func GetUsedAddress(api *API, seed Trytes, security int) (Address, []Address, error) {
	var all []Address
	for index := 0; ; index++ {
		adr, err := NewAddress(seed, index, security)
		if err != nil {
			return "", nil, err
		}

		r := FindTransactionsRequest{
			Addresses: []Address{adr},
		}

		resp, err := api.FindTransactions(&r)
		if err != nil {
			return "", nil, err
		}

		if len(resp.Hashes) == 0 {
			return adr, all, nil
		}

		// reached the end of the loop, so must be used address, repeat until return
		all = append(all, adr)
	}
}

// GetInputs gets all possible inputs of a seed and returns them with the total balance.
// end must be under start+500.
func GetInputs(api *API, seed Trytes, start, end int, threshold int64, security int) (Balances, error) {
	var err error
	var adrs []Address

	if start > end || end > (start+500) {
		return nil, errors.New("Invalid start/end provided")
	}

	switch {
	case end > 0:
		adrs, err = NewAddresses(seed, start, end-start, security)
	default:
		_, adrs, err = GetUsedAddress(api, seed, security)
	}

	if err != nil {
		return nil, err
	}

	return api.Balances(adrs)
}

// Transfer is the  data to be transfered by bundles.
type Transfer struct {
	Address Address
	Value   int64
	Message Trytes
	Tag     Trytes
}

const sigSize = SignatureMessageFragmentTrinarySize / 3

func addOutputs(trs []Transfer) (Bundle, []Trytes, int64) {
	var (
		bundle Bundle
		frags  []Trytes
		total  int64
	)
	for _, tr := range trs {
		nsigs := 1

		// If message longer than 2187 trytes, increase signatureMessageLength (add 2nd transaction)
		switch {
		case len(tr.Message) > sigSize:
			// Get total length, message / maxLength (2187 trytes)
			n := int(math.Floor(float64(len(tr.Message)) / sigSize))
			nsigs += n

			// While there is still a message, copy it
			for k := 0; k < n; k++ {
				var fragment Trytes
				switch {
				case k == n-1:
					fragment = tr.Message[k*sigSize:]
				default:
					fragment = tr.Message[k*sigSize : (k+1)*sigSize]
				}

				// Pad remainder of fragment
				frags = append(frags, fragment)
			}
		default:
			frags = append(frags, tr.Message)
		}

		// Add first entries to the bundle
		// Slice the address in case the user provided a checksummed one
		bundle.Add(nsigs, tr.Address, tr.Value, time.Now(), tr.Tag)

		// Sum up total value
		total += tr.Value
	}
	return bundle, frags, total
}

// AddressInfo includes an address and its infomation for signing.
type AddressInfo struct {
	Seed     Trytes
	Index    int
	Security int
}

// Address makes an Address from an AddressInfo
func (a *AddressInfo) Address() (Address, error) {
	return NewAddress(a.Seed, a.Index, a.Security)
}

// Key makes a Key from an AddressInfo
func (a *AddressInfo) Key() (Trytes, error) {
	return NewKey(a.Seed, a.Index, a.Security)
}

func setupInputs(api *API, seed Trytes, inputs []AddressInfo, security int, total int64) (Balances, []AddressInfo, error) {
	var bals Balances
	var err error

	switch {
	case inputs == nil:
		//  Case 2: Get inputs deterministically
		//  If no inputs provided, derive the addresses from the seed and
		//  confirm that the inputs exceed the threshold

		// If inputs with enough balance
		bals, err = GetInputs(api, seed, 0, 100, total, security)
		if err != nil {
			return nil, nil, err
		}

		inputs = make([]AddressInfo, len(bals))
		for i := range bals {
			inputs[i].Index = bals[i].Index
			inputs[i].Security = security
			inputs[i].Seed = seed
		}
	default:
		//  Case 1: user provided inputs
		adrs := make([]Address, len(inputs))
		for i, ai := range inputs {
			adrs[i], err = ai.Address()
			if err != nil {
				return nil, nil, err
			}
		}

		//  Validate the inputs by calling getBalances (in call to Balances)
		bals, err = api.Balances(adrs)
		if err != nil {
			return nil, nil, err
		}
	}

	// Return not enough balance error
	if total > bals.Total() {
		return nil, nil, errors.New("Not enough balance")
	}
	return bals, inputs, nil
}

// PrepareTransfers gets an array of transfer objects as input, and then prepares
// the transfer by generating the correct bundle as well as choosing and signing the
// inputs if necessary (if it's a value transfer).
func PrepareTransfers(api *API, seed Trytes, trs []Transfer, inputs []AddressInfo, remainder Address, security int) (Bundle, error) {
	var err error

	bundle, frags, total := addOutputs(trs)

	// Get inputs if we are sending tokens
	if total <= 0 {
		// If no input required, don't sign and simply finalize the bundle
		bundle.Finalize(frags)
		return bundle, nil
	}

	bals, inputs, err := setupInputs(api, seed, inputs, security, total)
	if err != nil {
		return nil, err
	}

	err = addRemainder(api, bals, &bundle, security, remainder, seed, total)
	if err != nil {
		return nil, err
	}

	bundle.Finalize(frags)
	err = signInputs(inputs, bundle)
	return bundle, err
}

func addRemainder(api *API, in Balances, bundle *Bundle, security int, remainder Address, seed Trytes, total int64) error {
	for _, bal := range in {
		var err error

		// Add input as bundle entry
		bundle.Add(security, bal.Address, -bal.Value, time.Now(), EmptyHash)

		// If there is a remainder value add extra output to send remaining funds to
		if remain := bal.Value - total; remain > 0 {
			// If user has provided remainder address use it to send remaining funds to
			adr := remainder
			if adr == "" {
				// Generate a new Address by calling getNewAddress
				adr, _, err = GetUsedAddress(api, seed, security)
				if err != nil {
					return err
				}
			}

			// Remainder bundle entry
			bundle.Add(1, adr, remain, time.Now(), EmptyHash)
			return nil
		}

		// If multiple inputs provided, subtract the totalTransferValue by
		// the inputs balance
		if total -= bal.Value; total == 0 {
			return nil
		}
	}
	return nil
}

func signInputs(inputs []AddressInfo, bundle Bundle) error {
	//  Get the normalized bundle hash
	nHash := bundle.Hash().Normalize()

	// SIGNING OF INPUTS
	// Here we do the actual signing of the inputs. Iterate over all bundle transactions,
	// find the inputs, get the corresponding private key, and calculate signatureFragment
	for i, bd := range bundle {
		if bd.Value >= 0 {
			continue
		}

		// Get the corresponding keyIndex and security of the address
		var ai AddressInfo
		for _, in := range inputs {
			adr, err := in.Address()
			if err != nil {
				return err
			}

			if adr == bd.Address {
				ai = in
				break
			}
		}

		// Get corresponding private key of the address
		key, err := ai.Key()
		if err != nil {
			return err
		}

		// Calculate the new signatureFragment with the first bundle fragment
		bundle[i].SignatureMessageFragment = Sign(nHash[:27], key[:6561/3])

		// if user chooses higher than 27-tryte security
		// for each security level, add an additional signature
		for j := 1; j < ai.Security; j++ {
			//  Because the signature is > 2187 trytes, we need to find the subsequent
			// transaction to add the remainder of the signature same address as well
			// as value = 0 (as we already spent the input)
			if bundle[i+j].Address == bd.Address && bundle[i+j].Value == 0 {
				//  Calculate the new signature
				nfrag := Sign(nHash[(j%3)*27:(j%3)*27+27], key[6561*j/3:(j+1)*6561/3])
				//  Convert signature to trytes and assign it again to this bundle entry
				bundle[i+j].SignatureMessageFragment = nfrag
			}
		}
	}
	return nil
}

func doPow(tra *GetTransactionsToApproveResponse, depth int64, trytes []Transaction, mwm int64, pow PowFunc) error {
	var prev Trytes
	var err error
	for i := len(trytes) - 1; i >= 0; i-- {
		switch {
		case i == len(trytes)-1:
			trytes[i].TrunkTransaction = tra.TrunkTransaction
			trytes[i].BranchTransaction = tra.BranchTransaction
		default:
			trytes[i].TrunkTransaction = prev
			trytes[i].BranchTransaction = tra.TrunkTransaction
		}

		trytes[i].Nonce, err = pow(trytes[i].Trytes(), int(mwm))
		if err != nil {
			return err
		}

		prev = trytes[i].Hash()
	}
	return nil
}

// SendTrytes does attachToTangle and finally, it broadcasts the transactions.
func SendTrytes(api *API, depth int64, trytes []Transaction, mwm int64, pow PowFunc) error {
	tra, err := api.GetTransactionsToApprove(depth, NumberOfWalks, "")
	if err != nil {
		return err
	}

	switch {
	case pow == nil:
		at := AttachToTangleRequest{
			TrunkTransaction:   tra.TrunkTransaction,
			BranchTransaction:  tra.BranchTransaction,
			MinWeightMagnitude: mwm,
			Trytes:             trytes,
		}

		// attach to tangle - do pow
		attached, err := api.AttachToTangle(&at)
		if err != nil {
			return err
		}

		trytes = attached.Trytes
	default:
		err := doPow(tra, depth, trytes, mwm, pow)
		if err != nil {
			return err
		}
	}

	// Broadcast and store tx
	return api.BroadcastTransactions(trytes)
}

// Send sends tokens. If you need to do pow locally, you must specifiy pow func,
// otherwise this calls the AttachToTangle API
func Send(api *API, seed Trytes, security int, trs []Transfer, mwm int64, pow PowFunc) (Bundle, error) {
	bd, err := PrepareTransfers(api, seed, trs, nil, "", security)
	if err != nil {
		return nil, err
	}

	err = SendTrytes(api, Depth, []Transaction(bd), mwm, pow)
	return bd, err
}
