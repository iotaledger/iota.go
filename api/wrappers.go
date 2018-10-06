package api

import (
	"github.com/iotaledger/iota.go/address"
	"github.com/iotaledger/iota.go/api_errors"
	"github.com/iotaledger/iota.go/bundle"
	"github.com/iotaledger/iota.go/checksum"
	"github.com/iotaledger/iota.go/curl"
	"github.com/iotaledger/iota.go/signing"
	"github.com/iotaledger/iota.go/transaction"
	"github.com/iotaledger/iota.go/transaction_converter"
	. "github.com/iotaledger/iota.go/trinary"
	"math"
	"sort"
	"sync"
	"time"
)

func createBroadcastBundle(provider Provider) func(tailTxHash Hash) ([]Trytes, error) {
	var broadcastTransactions = createBroadcastTransactions(provider)
	var getBundle = createGetBundle(provider)
	return func(tailTxHash Hash) ([]Trytes, error) {
		// TODO: validate tx hash
		bndl, err := getBundle(tailTxHash)
		if err != nil {
			return nil, err
		}
		trytes := transaction.FinalTransactionTrytes(bndl)
		return broadcastTransactions(trytes...)
	}
}

func createGetAccountData(provider Provider) func(seed Trytes, options GetAccountDataOptions) (*AccountData, error) {
	var getNewAddress = createGetNewAddress(provider)
	var getBundlesFromAddresses = createGetBundlesFromAddresses(provider)
	var getBalances = createGetBalances(provider)
	var wereAddressesSpentFrom = createWereAddressesSpentFrom(provider)
	return func(seed Trytes, options GetAccountDataOptions) (*AccountData, error) {
		// TODO: validate start<->end, seed, security lvl

		var total *uint64
		if options.End != nil {
			t := *options.End - options.Start
			total = &t
		}

		addresses, err := getNewAddress(seed, GetNewAddressOptions{
			Index: options.Start, Total: total,
			ReturnAll: true, Security: options.Security,
		})
		if err != nil {
			return nil, err
		}

		var err1, err2, err3 error
		var bundles bundle.Bundles
		var balances *Balances
		var spentState []bool

		wg := sync.WaitGroup{}
		wg.Add(3)
		go func() {
			defer wg.Done()
			bundles, err1 = getBundlesFromAddresses(addresses)
		}()

		go func() {
			defer wg.Done()
			balances, err2 = getBalances(addresses, 100)
		}()

		go func() {
			defer wg.Done()
			spentState, err3 = wereAddressesSpentFrom(addresses...)
		}()

		wg.Wait()
		if err := firstNonNilErr(err1, err2, err3); err != nil {
			return nil, err
		}

		// extract tx hashes which operated on the account's addresses
		// as input or output tx
		var txsHashes Hashes
		for i := range bundles {
			bndl := &bundles[i]
			for j := range *bndl {
				tx := &(*bndl)[j]
				for x := range addresses {
					if tx.Address == addresses[x] {
						txsHashes = append(txsHashes, tx.Hash)
						break
					}
				}
			}
		}

		// compute balances
		inputs := []Address{}
		var totalBalance uint64
		for i := range addresses {
			value := balances.Balances[i]
			// this works because the balances and spent states are ordered
			if spentState[i] || value <= 0 {
				continue
			}
			totalBalance += value

			address := Address{
				Address: addresses[i], Security: options.Security,
				KeyIndex: options.Start + uint64(i), Balance: value,
			}
			inputs = append(inputs, address)
		}

		account := &AccountData{
			Transfers:     bundles,
			Transactions:  txsHashes,
			Inputs:        inputs,
			Balance:       totalBalance,
			LatestAddress: addresses[len(addresses)-1],
			Addresses:     addresses,
		}

		return account, nil
	}
}

func firstNonNilErr(errs ...error) error {
	for x := range errs {
		if errs[x] != nil {
			return errs[x]
		}
	}
	return nil
}

func createGetBundle(provider Provider) func(tailTxHash Hash) (bundle.Bundle, error) {
	var traverseBundle = createTraverseBundle(provider)
	return func(tailTxHash Hash) (bundle.Bundle, error) {
		// TODO: validate tail tx hash
		bndl := bundle.Bundle{}
		return traverseBundle(tailTxHash, bndl)
	}
}

func createGetBundlesFromAddresses(provider Provider) func(addresses Hashes, inclusionState ...bool) (bundle.Bundles, error) {
	var findTransactionObjects = createFindTransactionObjects(provider)
	var getLatestInclusion = createGetLatestInclusion(provider)
	return func(addresses Hashes, inclusionState ...bool) (bundle.Bundles, error) {
		txs, err := findTransactionObjects(FindTransactionsQuery{Addresses: addresses})
		if err != nil {
			return nil, err
		}

		bundleHashesSet := map[Trytes]struct{}{}
		for i := range txs {
			bundleHashesSet[txs[i].Bundle] = struct{}{}
		}

		bundleHashes := make([]Trytes, len(bundleHashesSet))
		for hash := range bundleHashesSet {
			bundleHashes = append(bundleHashes, hash)
		}

		allTxs, err := findTransactionObjects(FindTransactionsQuery{Bundles: bundleHashes})
		if err != nil {
			return nil, err
		}
		bundles := bundle.GroupTransactionsIntoBundles(allTxs)
		sort.Sort(bundle.BundlesByTimestamp(bundles))

		if len(inclusionState) > 0 && inclusionState[0] {
			// get tail tx hashes
			hashes := Hashes{}
			for i := range bundles {
				hashes = append(hashes, bundles[i][0].Hash)
			}

			states, err := getLatestInclusion(hashes)
			if err != nil {
				return nil, err
			}

			// set confirmed property on each tx
			// since bundles are atomic, each tx in the bundle
			// as the same 'confirmed' state
			for i := range bundles {
				bndl := &bundles[i]
				for j := range *bndl {
					tx := &(*bndl)[j]
					tx.Confirmed = &states[i]
				}
			}
		}

		return bundles, err
	}
}

func createGetLatestInclusion(provider Provider) func(transactions Hashes) ([]bool, error) {
	var getNodeInfo = createGetNodeInfo(provider)
	var getInclusionStates = createInclusionState(provider)
	return func(transactions Hashes) ([]bool, error) {
		nodeInfo, err := getNodeInfo()
		if err != nil {
			return nil, err
		}
		return getInclusionStates(transactions, nodeInfo.LatestSolidSubtangleMilestone)
	}
}

func createGetNewAddress(provider Provider) func(seed Trytes, options GetNewAddressOptions) ([]Trytes, error) {
	var isAddressUsed = createIsAddressUsed(provider)
	return func(seed Trytes, options GetNewAddressOptions) ([]Trytes, error) {
		// TODO: validate seed, index, security
		options = getNewAddressDefaultOptions(options)
		index := options.Index
		securityLvl := options.Security

		var addresses Hashes
		var err error

		if options.Total != nil && *options.Total > 0 {
			total := *options.Total
			addresses, err = address.GenerateAddresses(seed, index, total, securityLvl)
		} else {
			addresses, err = getUntilFirstUnusedAddress(isAddressUsed, seed, index, securityLvl, options.ReturnAll)
		}

		// TODO: apply checksum option
		return addresses, err
	}
}

func createIsAddressUsed(provider Provider) func(address Hash) (bool, error) {
	var wereAddressesSpentFrom = createWereAddressesSpentFrom(provider)
	var findTransactions = createFindTransactionObjects(provider)
	return func(address Hash) (bool, error) {
		// TODO: use goroutines to parallelize
		states, err := wereAddressesSpentFrom(address)
		if err != nil {
			return false, err
		}
		state := states[0]
		if state {
			return state, nil
		}
		txs, err := findTransactions(FindTransactionsQuery{Addresses: Hashes{address}})
		if err != nil {
			return false, err
		}
		return len(txs) > 0, nil
	}
}

func getUntilFirstUnusedAddress(
	isAddressUsed func(address Hash) (bool, error),
	seed Trytes, index uint64, security signing.SecurityLevel,
	returnAll bool,
) (Hashes, error) {
	addresses := Hashes{}

	for ; ; index++ {
		nextAddress, err := address.GenerateAddress(seed, index, security)
		if err != nil {
			return nil, err
		}

		if returnAll {
			addresses = append(addresses, nextAddress)
		}

		used, err := isAddressUsed(nextAddress)
		if err != nil {
			return nil, err
		}

		if used {
			continue
		}

		if !returnAll {
			addresses = append(addresses, nextAddress)
		}

		return addresses, nil
	}
}

func createGetTransactionObjects(provider Provider) func(hashes ...Hash) (transaction.Transactions, error) {
	var getTrytes = createGetTrytes(provider)
	return func(hashes ...Hash) (transaction.Transactions, error) {
		// TODO: validate hashes
		trytes, err := getTrytes(hashes...)
		if err != nil {
			return nil, err
		}
		return transaction.AsTransactionObjects(trytes, hashes)
	}
}

func createFindTransactionObjects(provider Provider) func(query FindTransactionsQuery) (transaction.Transactions, error) {
	var findTransactions = createFindTransactions(provider)
	var getTransactionObjects = createGetTransactionObjects(provider)
	return func(query FindTransactionsQuery) (transaction.Transactions, error) {
		txHashes, err := findTransactions(query)
		if err != nil {
			return nil, err
		}
		return getTransactionObjects(txHashes...)
	}
}

func createGetInputs(provider Provider) func(seed Trytes, options GetInputOptions) (*Inputs, error) {
	var getNewAddress = createGetNewAddress(provider)
	var getBalances = createGetBalances(provider)
	return func(seed Trytes, options GetInputOptions) (*Inputs, error) {
		// TODO: validate start, end, security, threshold
		opts := options.ToGetNewAddressOptions()
		addresses, err := getNewAddress(seed, opts)
		if err != nil {
			return nil, err
		}
		balances, err := getBalances(addresses, 100)
		if err != nil {
			return nil, err
		}

		inputs := createInputObjects(addresses, balances.Balances, opts.Index, opts.Security)

		// threshold is a hard cap for needed inputs to fulfil the threshold value
		if options.Threshold != nil {
			threshold := *options.Threshold

			if threshold > inputs.TotalBalance {
				return nil, api_errors.ErrInsufficientBalance
			}

			thresholdInputs := Inputs{}
			for i := range inputs.Inputs {
				if thresholdInputs.TotalBalance >= threshold {
					break
				}
				input := inputs.Inputs[i]
				thresholdInputs.Inputs = append(thresholdInputs.Inputs, input)
				thresholdInputs.TotalBalance += input.Balance
			}
			inputs = thresholdInputs
		}

		return &inputs, nil
	}
}

func createInputObjects(addresses Hashes, balances []uint64, start uint64, secLvl signing.SecurityLevel) Inputs {
	addrs := []Address{}
	var totalBalance uint64
	for i := range addresses {
		value := balances[i]
		addrs = append(addrs, Address{
			Address: addresses[i], Security: secLvl,
			Balance: value, KeyIndex: start + uint64(i)},
		)
		totalBalance += value
	}
	return Inputs{Inputs: addrs, TotalBalance: totalBalance}
}

func createGetTransfers(provider Provider) func(seed Trytes, options GetTransfersOptions) (bundle.Bundles, error) {
	var getNewAddress = createGetNewAddress(provider)
	var getBundlesFromAddresses = createGetBundlesFromAddresses(provider)
	return func(seed Trytes, options GetTransfersOptions) (bundle.Bundles, error) {
		// TODO: validate seed, sec lvl, start, end
		addresses, err := getNewAddress(seed, options.ToGetNewAddressOptions())
		if err != nil {
			return nil, err
		}
		return getBundlesFromAddresses(addresses, options.InclusionStates)
	}
}

func createIsPromotable(provider Provider) func(tailTxHash Hash) (bool, error) {
	var checkConsistency = createCheckConsistency(provider)
	var getTrytes = createGetTrytes(provider)
	return func(tailTxHash Hash) (bool, error) {

		var err1, err2 error
		var isConsistent bool
		var trytes []Trytes
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			isConsistent, err1 = checkConsistency(tailTxHash)
		}()

		go func() {
			defer wg.Done()
			trytes, err2 = getTrytes(tailTxHash)
		}()

		wg.Wait()
		if err := firstNonNilErr(err1, err2); err != nil {
			return false, err
		}

		tx, err := transaction.NewTransaction(trytes[0])
		if err != nil {
			return false, err
		}

		return isConsistent && isAboveMaxDepth(tx.AttachmentTimestamp), nil
	}
}

const MilestoneInterval = 2 * 60 * 1000
const OneWayDelay = 1 * 60 * 1000
const maxDepth = 6

// checks whether by the given timestamp the transaction is to deep to be promoted
func isAboveMaxDepth(attachmentTimestamp int64) bool {
	nowMilli := time.Now().UnixNano() / int64(time.Millisecond)
	return attachmentTimestamp < nowMilli && nowMilli-attachmentTimestamp < maxDepth*MilestoneInterval*OneWayDelay
}

func createIsReattachable(provider Provider) func(inputAddresses ...Trytes) ([]bool, error) {
	var findTransactionObjects = createFindTransactionObjects(provider)
	var inclusionState = createInclusionState(provider)
	return func(inputAddresses ...Hash) ([]bool, error) {
		// TODO: make sure to remove checksums from addresses
		txs, err := findTransactionObjects(FindTransactionsQuery{Addresses: inputAddresses})
		if err != nil {
			return nil, err
		}

		// filter out zero value or receiving txs
		filteredTxs := transaction.Transactions{}
		for i := range txs {
			if txs[i].Value >= 0 {
				continue
			}
			filteredTxs = append(filteredTxs, txs[i])
		}

		// no spending tx found, therefore addresses
		// are all reattachtable
		if len(filteredTxs) == 0 {
			bools := []bool{}
			for i := range inputAddresses {
				bools[i] = true
			}
			return bools, nil
		}

		txHashes := make(Hashes, len(filteredTxs))
		for i := range filteredTxs {
			txHashes = append(txHashes, filteredTxs[i].Hash)
		}

		states, err := inclusionState(txHashes)
		if err != nil {
			return nil, err
		}

		for i := range filteredTxs {
			filteredTxs[i].Confirmed = &states[i]
		}

		// map addresses to whether any input tx is confirmed
		addrToOneConf := make([]bool, len(inputAddresses))
		for i := range inputAddresses {
			anyConfirmed := false
			for j := range filteredTxs {
				if *filteredTxs[j].Confirmed && filteredTxs[j].Address == inputAddresses[i] {
					anyConfirmed = true
					break
				}
			}

			// reverse state: isReattachable = negated inclusion state
			addrToOneConf = append(addrToOneConf, !anyConfirmed)
		}

		return addrToOneConf, nil
	}
}

func createPrepareTransfers(provider Provider) func(seed Trytes, transfers bundle.Transfers, options PrepareTransfersOptions) ([]Trytes, error) {
	var getInputs = createGetInputs(provider)
	var wereAddressesSpentFrom = createWereAddressesSpentFrom(provider)
	var getNewAddress = createGetNewAddress(provider)
	return func(seed Trytes, transfers bundle.Transfers, options PrepareTransfersOptions) ([]Trytes, error) {
		options = getPrepareTransfersDefaultOptions(options)

		props := PrepareTransferProps{
			Seed: seed, Security: options.Security, Inputs: options.Inputs,
			Timestamp: uint64(time.Now().UnixNano() / int64(time.Second)),
			Transfers: transfers, Transactions: transaction.Transactions{},
			Trytes: []Trytes{}, HMACKey: options.HMACKey, RemainderAddress: options.RemainderAddress,
		}

		var totalTransferValue uint64
		for i := range transfers {
			totalTransferValue += transfers[i].Value
		}

		// TODO: add HMAC placeholder txs

		// add transfers
		for i := range props.Transfers {
			transfer := &props.Transfers[i]
			msgLength := len(transfer.Message)
			length := math.Ceil(float64(msgLength)) / bundle.SignatureMessageFragmentSizeInTrytes
			if length == 0 {
				length = 1
			}
			addr, err := checksum.RemoveChecksum(transfer.Address)
			if err != nil {
				return nil, err
			}

			bndlEntry := bundle.BundleEntry{
				Address: addr, Value: int64(transfer.Value),
				Tag: transfer.Tag, Timestamp: props.Timestamp,
				Length: uint64(length),
				SignatureMessageFragments: func() []Trytes {
					splitFrags := make([]Trytes, int(length))
					for i := 0; i < int(length); i++ {
						splitFrags[i] = transfer.Message[i*bundle.SignatureMessageFragmentSizeInTrytes : (i+1)*bundle.SignatureMessageFragmentSizeInTrytes]
					}
					return splitFrags
				}(),
			}

			props.Transactions = bundle.AddEntry(props.Transactions, bndlEntry)
		}

		// gather inputs if we have a value transfer but no inputs were specified.
		// this would error out if the gathered inputs don't fulfill the threshold value
		if totalTransferValue != 0 && len(props.Inputs) == 0 {
			inputs, err := getInputs(seed, GetInputOptions{Security: props.Security, Threshold: &totalTransferValue})
			if err != nil {
				return nil, err
			}

			// filter out inputs which are already spent
			inputAddresses := make(Hashes, len(props.Inputs))
			for i := range props.Inputs {
				inputAddresses[i] = inputs.Inputs[i].Address
			}

			states, err := wereAddressesSpentFrom(inputAddresses...)
			if err != nil {
				return nil, err
			}
			for i, state := range states {
				if state {
					inputs.Inputs = append(inputs.Inputs[:i], inputs.Inputs[i+1:]...)
				}
			}

			props.Inputs = inputs.Inputs
		}

		// add input transactions
		var inputsTotal uint64
		for i := range props.Inputs {
			inputsTotal += props.Inputs[i].Balance
			input := &props.Inputs[i]
			addr, err := checksum.RemoveChecksum(input.Address)
			if err != nil {
				return nil, err
			}
			bndlEntry := bundle.BundleEntry{
				Address:   addr,
				Value:     -int64(input.Balance),
				Length:    uint64(input.Security),
				Timestamp: props.Timestamp,
			}
			props.Transactions = bundle.AddEntry(props.Transactions, bndlEntry)
		}

		// verify whether provided inputs fulfill threshold value
		if inputsTotal < totalTransferValue {
			return nil, api_errors.ErrInsufficientBalance
		}

		// TODO: document if inputs are provided by the caller, then they are not checked for spent state

		// compute remainder
		var remainder int64
		for i := range props.Transactions {
			remainder += props.Transactions[i].Value
		}

		if remainder > 0 {
			return nil, api_errors.ErrInsufficientBalance
		}

		// add remainder transaction if there's a remainder
		if remainder != 0 {
			// compute new remainder address if non supplied
			if totalTransferValue > 0 && props.RemainderAddress == nil {
				remainderAddressKeyIndex := props.Inputs[0].KeyIndex
				for i := range props.Inputs {
					keyIndex := props.Inputs[i].KeyIndex
					if keyIndex > remainderAddressKeyIndex {
						remainderAddressKeyIndex = keyIndex
					}
				}
				remainderAddressKeyIndex++
				addrs, err := getNewAddress(seed, GetNewAddressOptions{Security: props.Security, Index: remainderAddressKeyIndex})
				if err != nil {
					return nil, err
				}
				props.RemainderAddress = &addrs[0]
			}

			// add remainder transaction
			if totalTransferValue > 0 {
				bundle.AddEntry(props.Transactions, bundle.BundleEntry{
					Address: *props.RemainderAddress,
					Length:  1, Timestamp: props.Timestamp,
					Value: int64(math.Abs(float64(remainder))),
				})
			}
		}

		// verify that input txs don't send to the same address
		for i := range props.Transactions {
			tx := &props.Transactions[i]
			// only check output txs
			if tx.Value <= 0 {
				continue
			}
			// check whether any input uses the same address as the output tx
			for j := range props.Inputs {
				if props.Inputs[j].Address == tx.Address {
					return nil, api_errors.ErrSendingBackToInputs
				}
			}
		}

		// finalize bundle by adding the bundle hash
		finalizedBundle, err := bundle.Finalize(props.Transactions)
		if err != nil {
			return nil, err
		}

		// compute signatures for all input txs
		normalizedBundle := signing.NormalizedBundleHash(finalizedBundle[0].Bundle)

		signedFrags := []Trytes{}
		for i := range props.Inputs {
			input := &props.Inputs[i]
			subseed, err := signing.Subseed(seed, input.KeyIndex)
			if err != nil {
				return nil, err
			}
			var sec signing.SecurityLevel
			if input.Security == 0 {
				sec = signing.SecurityLevelMedium
			} else {
				sec = input.Security
			}

			prvKey, err := signing.Key(subseed, sec)
			if err != nil {
				return nil, err
			}

			frags := make([]Trytes, input.Security)
			for i := 0; i < int(input.Security); i++ {
				signedFragTrits, err := signing.SignatureFragment(
					normalizedBundle[i*curl.HashSize/3:(i+1)*curl.HashSize/3],
					prvKey[i*signing.KeyFragmentLength:(i+1)*signing.KeyFragmentLength],
				)
				if err != nil {
					return nil, err
				}
				frags[i] = MustTritsToTrytes(signedFragTrits)
			}

			signedFrags = append(signedFrags, frags...)
		}

		// add signed fragments to txs
		var indexFirstInputTx int
		for i := range props.Transactions {
			if props.Transactions[i].Value < 0 {
				indexFirstInputTx = i
				break
			}
		}

		props.Transactions = bundle.AddTrytes(props.Transactions, signedFrags, indexFirstInputTx)

		// TODO: add HMAC

		// finally return built up txs as raw trytes
		return transaction.FinalTransactionTrytes(props.Transactions), nil
	}
}

func createSendTransfer(provider Provider, attachToTangle AttachToTangle) func(seed Trytes, depth uint64, mwm uint64, transfers bundle.Transfers, options *SendTransfersOptions) (bundle.Bundle, error) {
	var prepareTransfers = createPrepareTransfers(provider)
	var sendTrytes = createSendTrytes(provider, attachToTangle)
	return func(seed Trytes, depth uint64, mwm uint64, transfers bundle.Transfers, options *SendTransfersOptions) (bundle.Bundle, error) {
		// TODO: validate depth, mwm, seed and transfers
		var opts PrepareTransfersOptions
		refs := Hashes{}
		if options == nil {
			opts = getPrepareTransfersDefaultOptions(PrepareTransfersOptions{})
		} else {
			opts = getPrepareTransfersDefaultOptions(options.PrepareTransfersOptions)
			if options.Reference != nil {
				refs = append(refs, *options.Reference)
			}
		}

		trytes, err := prepareTransfers(seed, transfers, opts)
		if err != nil {
			return nil, err
		}

		return sendTrytes(trytes, depth, mwm, refs...)
	}
}

func createPromoteTransactions(provider Provider, attachToTangle AttachToTangle) func(tailTxHash Hash, depth uint64, mwm uint64, spamTransfers bundle.Transfers, options PromoteTransactionOptions) (transaction.Transactions, error) {
	var checkConsistency = createCheckConsistency(provider)
	var sendTransfer = createSendTransfer(provider, attachToTangle)
	return func(tailTxHash Hash, depth uint64, mwm uint64, spamTransfers bundle.Transfers, options PromoteTransactionOptions) (transaction.Transactions, error) {
		// TODO: validate tail tx and spam transfers
		options = getPromoteTransactionsDefaultOptions(options)

		consistent, err := checkConsistency(tailTxHash)
		if err != nil {
			return nil, err
		}

		if !consistent {
			return nil, api_errors.ErrInconsistentSubtangle
		}

		opts := SendTransfersOptions{Reference: &tailTxHash}
		opts.PrepareTransfersOptions = getPrepareTransfersDefaultOptions(opts.PrepareTransfersOptions)
		getPrepareTransfersDefaultOptions(PrepareTransfersOptions{})
		return sendTransfer(spamTransfers[0].Address, depth, mwm, spamTransfers, &opts)
	}
}

func createReplayBundle(provider Provider, attachToTangle AttachToTangle) func(tailTxhash Hash, depth uint64, mwm uint64, reference ...Hash) (bundle.Bundle, error) {
	var getBundle = createGetBundle(provider)
	var sendTrytes = createSendTrytes(provider, attachToTangle)
	return func(tailTxhash Hash, depth uint64, mwm uint64, reference ...Hash) (bundle.Bundle, error) {
		// TODO: validate tail tx hash, depth and mwm
		bndl, err := getBundle(tailTxhash)
		if err != nil {
			return nil, err
		}
		trytes := transaction.FinalTransactionTrytes(bndl)
		return sendTrytes(trytes, depth, mwm)
	}
}

func createSendTrytes(provider Provider, attachToTangle AttachToTangle) func(trytes []Trytes, depth uint64, mwm uint64, reference ...Hash) (bundle.Bundle, error) {
	var getTransactionsToApprove = createGetTransactionsToApprove(provider)
	var storeAndBroadcast = createStoreAndBroadcast(provider)
	return func(trytes []Trytes, depth uint64, mwm uint64, reference ...Hash) (bundle.Bundle, error) {
		// TODO: validate transaction trytes, depth and mwm
		tips, err := getTransactionsToApprove(depth, reference...)
		if err != nil {
			return nil, err
		}
		trytes, err = attachToTangle(tips.TrunkTransaction, tips.BranchTransaction, mwm, trytes)
		if err != nil {
			return nil, err
		}
		trytes, err = storeAndBroadcast(trytes)
		if err != nil {
			return nil, err
		}
		return transaction.AsTransactionObjects(trytes, nil)
	}
}

func createStoreAndBroadcast(provider Provider) func(trytes []Trytes) ([]Trytes, error) {
	var storeTransactions = createStoreTransactions(provider)
	var broadcastTransactions = createBroadcastTransactions(provider)
	return func(trytes []Trytes) ([]Trytes, error) {
		trytes, err := storeTransactions(trytes...)
		if err != nil {
			return nil, err
		}
		return broadcastTransactions(trytes...)
	}
}

func createTraverseBundle(provider Provider) func(tailTxhash Hash, bndl bundle.Bundle) (transaction.Transactions, error) {
	var getTrytes = createGetTrytes(provider)
	var f func(trunkTxHash Hash, bndl bundle.Bundle) (transaction.Transactions, error)
	f = func(trunkTxHash Hash, bndl bundle.Bundle) (transaction.Transactions, error) {
		// TODO: validate trunk tx hash
		tailTrytes, err := getTrytes(trunkTxHash)
		if err != nil {
			return nil, err
		}
		tx, err := transaction_converter.AsTransactionObject(tailTrytes[0], trunkTxHash)
		if err != nil {
			return nil, err
		}
		// tail tx ?
		if len(bndl) == 0 {
			if !transaction.IsTailTransaction(tx) {
				return nil, api_errors.ErrInvalidTailTransactionHash
			}
		}
		bndl = append(bndl, *tx)
		if tx.CurrentIndex == tx.LastIndex {
			return bndl, nil
		}
		return f(tx.TrunkTransaction, bndl)
	}
	return f
}
